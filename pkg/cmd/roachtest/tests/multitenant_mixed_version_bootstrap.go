// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"runtime"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/roachtestutil/clusterupgrade"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

func registerMultiTenantMixedVersionBootstrap(r registry.Registry) {
	r.Add(registry.TestSpec{
		Name:              "multitenant-mixed-version-bootstrap",
		Cluster:           r.MakeClusterSpec(2),
		Owner:             registry.OwnerMultiTenant,
		NonReleaseBlocker: false,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			if c.IsLocal() && runtime.GOARCH == "arm64" {
				t.Skip("Skip under ARM64. See https://github.com/cockroachdb/cockroach/issues/89268")
			}
			runMultiTenantMixedVersionBootstrap(ctx, t, c, *t.BuildVersion())
		},
	})
}

// Sketch of the test:
//   - Create a tenant `123` and get its KV datums.
//   - Upgrade the storage binary but do not finalize.
//   - Create another tenant `456` and get its KV datums.
//   - Ensure that KV datums from tenant `123` match the ones from tenant `456`.
func runMultiTenantMixedVersionBootstrap(ctx context.Context, t test.Test, c cluster.Cluster, v version.Version) {
	predecessor, err := clusterupgrade.PredecessorVersion(v)
	require.NoError(t, err)

	// currentBinary := uploadVersion(ctx, t, c, c.All(), clusterupgrade.MainVersion)
	predecessorBinary := uploadVersion(ctx, t, c, c.All(), predecessor)

	kvNodes := c.Node(1)
	settings := install.MakeClusterSettings(install.BinaryOption(predecessorBinary), install.SecureOption(true))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, kvNodes)

	const tenant123ID = 123
	const tenant456ID = 456
	runner := sqlutils.MakeSQLRunner(c.Conn(ctx, t.L(), 1))
	// We'll sometimes have to wait out the backoff of the storage cluster
	// auto-update loop (at the time of writing 30s), plus some migrations may be
	// genuinely long-running.
	runner.SucceedsSoonDuration = 5 * time.Minute

	t.Status("creating tenants")
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1::INT)`, tenant123ID)
	runner.Exec(t, `SELECT crdb_internal.create_tenant($1::INT)`, tenant456ID)

	scanQuery := `SELECT crdb_internal.pretty_key(key, 1) as key, substring(encode(val, 'hex') from 9) as value 
					FROM crdb_internal.scan(crdb_internal.tenant_span($1::INT)) as t(key, val);`
	t.Status("getting kv datums")
	kvDatumsForTenant123 := runner.QueryStr(t, scanQuery, tenant123ID)
	kvDatumsForTenant456 := runner.QueryStr(t, scanQuery, tenant456ID)

	require.Equal(t, kvDatumsForTenant123, kvDatumsForTenant456)
}
