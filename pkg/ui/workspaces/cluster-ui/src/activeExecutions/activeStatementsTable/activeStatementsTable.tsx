// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React from "react";
import { Link } from "react-router-dom";
import { isSelectedColumn } from "../../columnsSelector/utils";
import { ColumnDescriptor } from "../../sortedtable";
import {
  activeStatementColumnsFromCommon,
  ExecutionsColumn,
  executionsTableTitles,
  getLabel,
} from "../execTableCommon";
import { ActiveStatement } from "../types";

export function makeActiveStatementsColumns(
  isCockroachCloud: boolean,
): ColumnDescriptor<ActiveStatement>[] {
  return [
    activeStatementColumnsFromCommon.executionID,
    {
      name: "execution",
      title: executionsTableTitles.execution("statement"),
      cell: (item: ActiveStatement) => (
        <Link to={`/execution/statement/${item.statementID}`}>
          {item.query}
        </Link>
      ),
      sort: (item: ActiveStatement) => item.query,
    },
    activeStatementColumnsFromCommon.status,
    activeStatementColumnsFromCommon.startTime,
    activeStatementColumnsFromCommon.elapsedTime,
    !isCockroachCloud
      ? activeStatementColumnsFromCommon.timeSpentWaiting
      : null,
    activeStatementColumnsFromCommon.applicationName,
  ].filter(col => col != null);
}

export function getColumnOptions(
  columns: ColumnDescriptor<ActiveStatement>[],
  selectedColumns: string[] | null,
): { label: string; value: string; isSelected: boolean }[] {
  return columns
    .filter(col => !col.alwaysShow)
    .map(col => ({
      value: col.name,
      label: getLabel(col.name as ExecutionsColumn, "statement"),
      isSelected: isSelectedColumn(selectedColumns, col),
    }));
}
