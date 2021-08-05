// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package colexec

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecbase"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecbase/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

// externalSorterState indicates the current state of the external sorter.
type externalSorterState int

const (
	// externalSorterNewPartition indicates that the next batch we read should
	// start a new partition. A zero-length batch in this state indicates that
	// the input to the external sorter has been fully consumed and we should
	// proceed to merging the partitions.
	externalSorterNewPartition externalSorterState = iota
	// externalSorterSpillPartition indicates that the next batch we read should
	// be added to the last partition so far. A zero-length batch in this state
	// indicates that the end of the partition has been reached and we should
	// transition to starting a new partition. If maxNumberPartitions is reached
	// in this state, the sorter will transition to externalSorterRepeatedMerging
	// to reduce the number of partitions.
	externalSorterSpillPartition
	// externalSorterRepeatedMerging indicates that we need to merge
	// maxNumberPartitions into one and spill that new partition to disk. When
	// finished, the sorter will transition to externalSorterNewPartition.
	externalSorterRepeatedMerging
	// externalSorterFinalMerging indicates that we have fully consumed the input
	// and can merge all of the partitions in one step. We then transition to
	// externalSorterEmitting state.
	externalSorterFinalMerging
	// externalSorterEmitting indicates that we are ready to emit output. A zero-
	// length batch in this state indicates that we have emitted all tuples and
	// should transition to externalSorterFinished state.
	externalSorterEmitting
	// externalSorterFinished indicates that all tuples from all partitions have
	// been emitted and from now on only a zero-length batch will be emitted by
	// the external sorter. This state is also responsible for closing the
	// partitions.
	externalSorterFinished
)

// In order to make progress when merging we have to merge at least two
// partitions into a new third one.
const externalSorterMinPartitions = 3

// externalSorter is an Operator that performs external merge sort. It works in
// two stages:
// 1. it will use a combination of an input partitioner and in-memory sorter to
// divide up all batches from the input into partitions, sort each partition in
// memory, and write sorted partitions to disk
// 2. it will use OrderedSynchronizer to merge the partitions.
//
// The (simplified) diagram of the components involved is as follows:
//
//                      input
//                        |
//                        ↓
//                 input partitioner
//                        |
//                        ↓
//                 in-memory sorter
//                        |
//                        ↓
//    ------------------------------------------
//   |             external sorter              |
//   |             ---------------              |
//   |                                          |
//   | partition1     partition2 ... partitionN |
//   |     |              |              |      |
//   |     ↓              ↓              ↓      |
//   |      merger (ordered synchronizer)       |
//    ------------------------------------------
//                        |
//                        ↓
//                      output
//
// There are a couple of implicit upstream links in the setup:
// - input partitioner checks the allocator used by the in-memory sorter to see
// whether a new partition must be started
// - external sorter resets in-memory sorter (which, in turn, resets input
// partitioner) once the full partition has been spilled to disk.
//
// What is hidden in the diagram is the fact that at some point we might need
// to merge several partitions into a new one that we spill to disk in order to
// reduce the number of "active" partitions. This requirement comes from the
// need to limit the number of "active" partitions because each partition uses
// some amount of RAM for its buffer. This is determined by
// maxNumberPartitions variable.
type externalSorter struct {
	OneInputNode
	NonExplainable
	closerHelper

	// mergeUnlimitedAllocator is used to track the memory under the batches
	// dequeued from partitions during the merge operation.
	mergeUnlimitedAllocator *colmem.Allocator
	// outputUnlimitedAllocator is used to track the memory under the output
	// batch in the merge operation.
	outputUnlimitedAllocator *colmem.Allocator

	totalMemoryLimit int64
	state            externalSorterState
	inputTypes       []*types.T
	ordering         execinfrapb.Ordering
	// columnOrdering is the same as ordering used when creating mergers.
	columnOrdering     colinfo.ColumnOrdering
	inMemSorter        ResettableOperator
	inMemSorterInput   *inputPartitioningOperator
	partitioner        colcontainer.PartitionedQueue
	partitionerCreator func() colcontainer.PartitionedQueue
	// partitionerToOperators stores all partitionerToOperator instances that we
	// have created when merging partitions. This allows for reusing them in
	// case we need to perform repeated merging (namely, we'll be able to reuse
	// their internal batches from dequeueing from disk).
	partitionerToOperators []*partitionerToOperator
	// numPartitions is the current number of partitions.
	numPartitions int
	// firstPartitionIdx is the index of the first partition to merge next.
	firstPartitionIdx   int
	maxNumberPartitions int

	// fdState is used to acquire file descriptors up front.
	fdState struct {
		fdSemaphore semaphore.Semaphore
		acquiredFDs int
	}

	emitter colexecbase.Operator

	testingKnobs struct {
		// delegateFDAcquisitions if true, means that a test wants to force the
		// PartitionedDiskQueues to track the number of file descriptors the hash
		// joiner will open/close. This disables the default behavior of acquiring
		// all file descriptors up front in Next.
		delegateFDAcquisitions bool
	}
}

var _ ResettableOperator = &externalSorter{}
var _ closableOperator = &externalSorter{}

// NewExternalSorter returns a disk-backed general sort operator.
// - unlimitedAllocators must have been created with a memory account derived
// from an unlimited memory monitor. They will be used by several internal
// components of the external sort which is responsible for making sure that
// the components stay within the memory limit.
// - maxNumberPartitions (when non-zero) overrides the semi-dynamically
// computed maximum number of partitions to have at once.
// - delegateFDAcquisitions specifies whether the external sorter should let
// the partitioned disk queue acquire file descriptors instead of acquiring
// them up front in Next. This should only be true in tests.
func NewExternalSorter(
	sortUnlimitedAllocator *colmem.Allocator,
	mergeUnlimitedAllocator *colmem.Allocator,
	outputUnlimitedAllocator *colmem.Allocator,
	input colexecbase.Operator,
	inputTypes []*types.T,
	ordering execinfrapb.Ordering,
	memoryLimit int64,
	maxNumberPartitions int,
	delegateFDAcquisitions bool,
	diskQueueCfg colcontainer.DiskQueueCfg,
	fdSemaphore semaphore.Semaphore,
	diskAcc *mon.BoundAccount,
) colexecbase.Operator {
	if diskQueueCfg.CacheMode != colcontainer.DiskQueueCacheModeReuseCache {
		colexecerror.InternalError(errors.Errorf("external sorter instantiated with suboptimal disk queue cache mode: %d", diskQueueCfg.CacheMode))
	}
	if diskQueueCfg.BufferSizeBytes > 0 && maxNumberPartitions == 0 {
		// With the default limit of 256 file descriptors, this results in 16
		// partitions. This is a hard maximum of partitions that will be used by the
		// external sorter
		// TODO(asubiotto): this number should be tuned.
		maxNumberPartitions = fdSemaphore.GetLimit() / 16
	}
	if maxNumberPartitions < externalSorterMinPartitions {
		maxNumberPartitions = externalSorterMinPartitions
	}
	estimatedOutputBatchMemSize := colmem.EstimateBatchSizeBytes(inputTypes, coldata.BatchSize())
	// Each disk queue will use up to BufferSizeBytes of RAM, so we reduce the
	// memoryLimit of the partitions to sort in memory by those cache sizes. To
	// be safe, we also estimate the size of the output batch and subtract that
	// as well.
	singlePartitionSize := memoryLimit - int64(maxNumberPartitions*diskQueueCfg.BufferSizeBytes+estimatedOutputBatchMemSize)
	if singlePartitionSize < 1 {
		// If the memory limit is 0, the input partitioning operator will return
		// a zero-length batch, so make it at least 1.
		singlePartitionSize = 1
	}
	inputPartitioner := newInputPartitioningOperator(input, singlePartitionSize)
	inMemSorter, err := newSorter(
		sortUnlimitedAllocator, newAllSpooler(sortUnlimitedAllocator, inputPartitioner, inputTypes),
		inputTypes, ordering.Columns,
	)
	if err != nil {
		colexecerror.InternalError(err)
	}
	partitionedDiskQueueSemaphore := fdSemaphore
	if !delegateFDAcquisitions {
		// To avoid deadlocks with other disk queues, we manually attempt to acquire
		// the maximum number of descriptors all at once in Next. Passing in a nil
		// semaphore indicates that the caller will do the acquiring.
		partitionedDiskQueueSemaphore = nil
	}
	es := &externalSorter{
		OneInputNode:             NewOneInputNode(inMemSorter),
		mergeUnlimitedAllocator:  mergeUnlimitedAllocator,
		outputUnlimitedAllocator: outputUnlimitedAllocator,
		totalMemoryLimit:         memoryLimit,
		inMemSorter:              inMemSorter,
		inMemSorterInput:         inputPartitioner.(*inputPartitioningOperator),
		partitionerCreator: func() colcontainer.PartitionedQueue {
			return colcontainer.NewPartitionedDiskQueue(inputTypes, diskQueueCfg, partitionedDiskQueueSemaphore, colcontainer.PartitionerStrategyCloseOnNewPartition, diskAcc)
		},
		inputTypes:     inputTypes,
		ordering:       ordering,
		columnOrdering: execinfrapb.ConvertToColumnOrdering(ordering),
		// TODO(yuzefovich): maxNumberPartitions should also be semi-dynamically
		// limited based on the sizes of batches. Consider a scenario when each
		// input batch is 1 GB in size - if we don't put any limiting in place,
		// we might try to use 16 partitions at once, which means that during
		// merging we will be keeping 16 batches (i.e. 16GB of data) in memory.
		maxNumberPartitions: maxNumberPartitions,
	}
	es.fdState.fdSemaphore = fdSemaphore
	es.testingKnobs.delegateFDAcquisitions = delegateFDAcquisitions
	return es
}

func (s *externalSorter) Init() {
	s.input.Init()
	s.state = externalSorterNewPartition
}

func (s *externalSorter) Next(ctx context.Context) coldata.Batch {
	for {
		switch s.state {
		case externalSorterNewPartition:
			b := s.input.Next(ctx)
			if b.Length() == 0 {
				// The input has been fully exhausted, and it is always the case that
				// the number of partitions is less than the maximum number since
				// externalSorterSpillPartition will check and re-merge if not.
				// Proceed to the final merging state.
				s.state = externalSorterFinalMerging
				continue
			}
			newPartitionIdx := s.firstPartitionIdx + s.numPartitions
			if s.partitioner == nil {
				s.partitioner = s.partitionerCreator()
			}
			if err := s.partitioner.Enqueue(ctx, newPartitionIdx, b); err != nil {
				handleErrorFromDiskQueue(err)
			}
			s.state = externalSorterSpillPartition
			continue
		case externalSorterSpillPartition:
			curPartitionIdx := s.firstPartitionIdx + s.numPartitions
			b := s.input.Next(ctx)
			if b.Length() == 0 {
				// The partition has been fully spilled, so we reset the in-memory
				// sorter (which will do the "shallow" reset of
				// inputPartitioningOperator).
				s.inMemSorterInput.interceptReset = true
				s.inMemSorter.reset(ctx)
				s.numPartitions++
				if s.numPartitions == s.maxNumberPartitions-1 {
					// We have reached the maximum number of active partitions that we
					// know that we'll be able to merge without exceeding the limit, so
					// we need to merge all of them and spill the new partition to disk
					// before we can proceed on consuming the input.
					s.state = externalSorterRepeatedMerging
					continue
				}
				s.state = externalSorterNewPartition
				continue
			}
			if !s.testingKnobs.delegateFDAcquisitions && s.fdState.fdSemaphore != nil && s.fdState.acquiredFDs == 0 {
				toAcquire := s.maxNumberPartitions
				if err := s.fdState.fdSemaphore.Acquire(ctx, toAcquire); err != nil {
					colexecerror.InternalError(err)
				}
				s.fdState.acquiredFDs = toAcquire
			}
			if err := s.partitioner.Enqueue(ctx, curPartitionIdx, b); err != nil {
				handleErrorFromDiskQueue(err)
			}
			continue
		case externalSorterRepeatedMerging:
			// We will merge all partitions in range [s.firstPartitionIdx,
			// s.firstPartitionIdx+s.numPartitions) and will spill all the
			// resulting batches into a new partition with the next available
			// index.
			//
			// The merger will be using some amount of RAM for the output batch,
			// will register it with the output unlimited allocator and will
			// *not* release that memory from the allocator, so we have to do it
			// ourselves.
			//
			// Note that the memory used for dequeueing batches from the
			// partitions is retained and is registered with the merge unlimited
			// allocator.
			merger, err := s.createMergerForPartitions(ctx)
			if err != nil {
				colexecerror.InternalError(err)
			}
			merger.Init()
			newPartitionIdx := s.firstPartitionIdx + s.numPartitions
			for b := merger.Next(ctx); b.Length() > 0; b = merger.Next(ctx) {
				if err := s.partitioner.Enqueue(ctx, newPartitionIdx, b); err != nil {
					handleErrorFromDiskQueue(err)
				}
			}
			// We are now done with the merger, so we can release the memory
			// used for the output batches (all of which have been enqueued into
			// the new partition).
			s.outputUnlimitedAllocator.ReleaseMemory(s.outputUnlimitedAllocator.Used())
			// Reclaim disk space by closing the inactive read partitions. Since
			// the merger must have exhausted all inputs, this is all the
			// partitions just read from.
			if err := s.partitioner.CloseInactiveReadPartitions(ctx); err != nil {
				colexecerror.InternalError(err)
			}
			s.firstPartitionIdx += s.numPartitions
			s.numPartitions = 1
			s.state = externalSorterNewPartition
			continue
		case externalSorterFinalMerging:
			if s.numPartitions == 0 {
				s.state = externalSorterFinished
				continue
			} else if s.numPartitions == 1 {
				s.createPartitionerToOperators()
				s.emitter = s.partitionerToOperators[0]
			} else {
				var err error
				s.emitter, err = s.createMergerForPartitions(ctx)
				if err != nil {
					colexecerror.InternalError(err)
				}
			}
			s.emitter.Init()
			s.state = externalSorterEmitting
			continue
		case externalSorterEmitting:
			b := s.emitter.Next(ctx)
			if b.Length() == 0 {
				s.state = externalSorterFinished
				continue
			}
			return b
		case externalSorterFinished:
			if err := s.Close(ctx); err != nil {
				colexecerror.InternalError(err)
			}
			return coldata.ZeroBatch
		default:
			colexecerror.InternalError(errors.AssertionFailedf("unexpected externalSorterState %d", s.state))
		}
	}
}

func (s *externalSorter) reset(ctx context.Context) {
	if r, ok := s.input.(resetter); ok {
		r.reset(ctx)
	}
	s.state = externalSorterNewPartition
	if err := s.Close(ctx); err != nil {
		colexecerror.InternalError(err)
	}
	// Reset closed so that the sorter may be closed again.
	s.closed = false
	s.firstPartitionIdx = 0
	s.numPartitions = 0
}

func (s *externalSorter) Close(ctx context.Context) error {
	if !s.close() {
		return nil
	}
	var lastErr error
	if s.partitioner != nil {
		lastErr = s.partitioner.Close(ctx)
		s.partitioner = nil
	}
	if err := s.inMemSorterInput.Close(ctx); err != nil {
		lastErr = err
	}
	if !s.testingKnobs.delegateFDAcquisitions && s.fdState.fdSemaphore != nil && s.fdState.acquiredFDs > 0 {
		s.fdState.fdSemaphore.Release(s.fdState.acquiredFDs)
		s.fdState.acquiredFDs = 0
	}
	return lastErr
}

// createPartitionerToOperators updates s.partitionerToOperators to correspond
// to all current partitions.
func (s *externalSorter) createPartitionerToOperators() {
	oldPartitioners := s.partitionerToOperators
	if len(oldPartitioners) < s.numPartitions {
		s.partitionerToOperators = make([]*partitionerToOperator, s.numPartitions)
		copy(s.partitionerToOperators, oldPartitioners)
		for i := len(oldPartitioners); i < s.numPartitions; i++ {
			s.partitionerToOperators[i] = newPartitionerToOperator(
				s.mergeUnlimitedAllocator, s.inputTypes, s.partitioner,
			)
		}
	}
	for i := 0; i < s.numPartitions; i++ {
		// We only need to set the partitioner and partitionIdx fields because
		// all others will not change when these operators are reused.
		s.partitionerToOperators[i].partitioner = s.partitioner
		s.partitionerToOperators[i].partitionIdx = s.firstPartitionIdx + i
	}
}

// createMergerForPartitions creates an ordered synchronizer that will merge
// partitions in [firstPartitionIdx, firstPartitionIdx+numPartitions) range.
func (s *externalSorter) createMergerForPartitions(
	ctx context.Context,
) (colexecbase.Operator, error) {
	s.createPartitionerToOperators()
	syncInputs := make([]SynchronizerInput, s.numPartitions)
	for i := range syncInputs {
		syncInputs[i].Op = s.partitionerToOperators[i]
	}
	// TODO(yuzefovich): we should calculate a more precise memory limit taking
	// into account how many partitions are currently being merged and the
	// average batch size in each one of them.
	return NewOrderedSynchronizer(
		s.outputUnlimitedAllocator, s.totalMemoryLimit, syncInputs, s.inputTypes, s.columnOrdering,
	)
}

func newInputPartitioningOperator(
	input colexecbase.Operator, memoryLimit int64,
) ResettableOperator {
	return &inputPartitioningOperator{
		OneInputNode: NewOneInputNode(input),
		memoryLimit:  memoryLimit,
	}
}

// inputPartitioningOperator is an operator that returns the batches from its
// input until the memory footprint of all emitted batches reaches the memory
// limit. From that point, the operator returns a zero-length batch (until it is
// reset).
type inputPartitioningOperator struct {
	OneInputNode
	NonExplainable

	// memoryLimit determines the size of each partition.
	memoryLimit int64
	// alreadyUsedMemory tracks the size of the current partition so far.
	alreadyUsedMemory int64
	// interceptReset determines whether the reset method will be called on
	// the input to this operator when the latter is being reset. This field is
	// managed by externalSorter.
	// NOTE: this field itself is set to 'false' when inputPartitioningOperator
	// is being reset, regardless of the original value.
	//
	// The reason for having this knob is that we need two kinds of behaviors
	// when resetting the inputPartitioningOperator:
	// 1. ("shallow" reset) we need to reset alreadyUsedMemory because the
	// external sorter is moving on spilling the data into a new partition.
	// However, we *cannot* propagate the reset further up because it might
	// delete the data that the external sorter has not yet spilled. This
	// behavior is needed in externalSorter when resetting the in-memory sorter
	// when spilling the next "chunk" of data into the new partition.
	// 2. ("deep" reset) we need to do the full reset of the whole chain of
	// operators. This behavior is needed when the whole external sorter is
	// being reset.
	interceptReset bool
}

var _ ResettableOperator = &inputPartitioningOperator{}

func (o *inputPartitioningOperator) Init() {
	o.input.Init()
}

func (o *inputPartitioningOperator) Next(ctx context.Context) coldata.Batch {
	if o.alreadyUsedMemory >= o.memoryLimit {
		return coldata.ZeroBatch
	}
	b := o.input.Next(ctx)
	if b.Length() == 0 {
		return b
	}
	// This operator is an input to sortOp which will spool all the tuples and
	// buffer them (by appending into the buffered batch), so we need to account
	// for memory proportionally to the length of the batch. (Note: this is not
	// exactly true for Bytes type, but it's ok if we have some deviation. This
	// numbers matter only to understand when to start a new partition, and the
	// memory will be actually accounted for correctly.)
	o.alreadyUsedMemory += colmem.GetProportionalBatchMemSize(b, int64(b.Length()))
	return b
}

func (o *inputPartitioningOperator) reset(ctx context.Context) {
	if !o.interceptReset {
		if r, ok := o.input.(resetter); ok {
			r.reset(ctx)
		}
	}
	o.interceptReset = false
	o.alreadyUsedMemory = 0
}

func (o *inputPartitioningOperator) Close(ctx context.Context) error {
	o.alreadyUsedMemory = 0
	return nil
}
