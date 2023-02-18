/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.github.matteobertozzi.yajbe.examples.util;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Function;

public class HumansTableView {
  private static final int COLUMN_WRAP_LENGTH = 80;

  private final ArrayList<Function<Object, String>> columnConverters = new ArrayList<>();
  private final ArrayList<String> columns = new ArrayList<>();
  private final ArrayList<Object[]> rows = new ArrayList<>();

  public HumansTableView addColumn(final String name, final Function<Object, String> converter) {
    this.columns.add(name);
    this.columnConverters.add(converter);
    return this;
  }

  public HumansTableView addRow(final Object... rowValues) {
    this.rows.add(rowValues);
    return this;
  }

  public HumansTableView addSeparator() {
    rows.add(null);
    return this;
  }

  public StringBuilder addHumanView(final StringBuilder builder) {
    return addHumanView(builder, true);
  }

  public StringBuilder addHumanView(final StringBuilder builder, final boolean drawHeader) {
    final List<String[]> strRows = convertRowsToString();
    final int[] columnsLength = calcColumnsLength(strRows);

    if (drawHeader) {
      drawHeaderBorder(builder, columnsLength);
      builder.append(drawRow(columnsLength, columns)).append('\n');
    }
    drawHeaderBorder(builder, columnsLength);
    for (final String[] row : strRows) {
      if (row == null) {
        drawHeaderBorder(builder, columnsLength);
      } else {
        builder.append(drawRow(columnsLength, List.of(row))).append('\n');
      }
    }
    drawHeaderBorder(builder, columnsLength);
    return builder;
  }

  public String toCsv() {
    final StringBuilder builder = new StringBuilder();
    for (int i = 0; i < columns.size(); ++i) {
      if (i > 0) builder.append(";");
      builder.append(columns.get(i));
    }
    builder.append(System.lineSeparator());
    for (final Object[] row: this.rows) {
      for (int i = 0; i < row.length; ++i) {
        if (i > 0) builder.append(";");
        builder.append(row[i]);
      }
      builder.append(System.lineSeparator());
    }
    return builder.toString();
  }

  private void drawHeaderBorder(final StringBuilder builder, final int[] columnsLength) {
    for (int i = 0; i < columnsLength.length; ++i) {
      builder.append("+-");
      builder.append("-".repeat(columnsLength[i]));
      builder.append('-');
    }
    builder.append("+\n");
  }

  private String drawRow(final int[] columnsLength, final List<String> values) {
    final StringBuilder buf = new StringBuilder();
    final ArrayList<String> truncatedColumns = new ArrayList<>();
    boolean hasTruncation = false;
    for (int i = 0; i < columnsLength.length; ++i) {
      final String colValue = values.get(i);
      buf.append("| ");
      if (columnsLength[i] < colValue.length()) {
        truncatedColumns.add(colValue.substring(columnsLength[i]));
        buf.append(colValue, 0, columnsLength[i]);
        hasTruncation = true;
      } else {
        truncatedColumns.add("");
        buf.append(colValue);
        for (int k = 0, kN = (columnsLength[i] - colValue.length()); k < kN; ++k) {
          buf.append(' ');
        }
      }
      buf.append(' ');
    }
    buf.append('|');

    if (hasTruncation) {
      buf.append('\n');
      buf.append(drawRow(columnsLength, truncatedColumns));
    }
    return buf.toString();
  }

  private List<String[]> convertRowsToString() {
    final ArrayList<String[]> strRows = new ArrayList<>(rows.size());
    for (final Object[] row: rows) {
      final String[] strRow = new String[columns.size()];
      for (int i = 0; i < columns.size(); ++i) {
        strRow[i] = valueOf(columnConverters.get(i), row[i]);
      }
      strRows.add(strRow);
    }
    return strRows;
  }

  private static String valueOf(final Function<Object, String> converter, final Object input) {
    if (converter != null) return converter.apply(input);
    if (input == null) return "(null)";

    final String value = String.valueOf(input);
    if (value.isEmpty()) return "";

    return value.replace('\t', ' ').replace('\n', ' ').replaceAll("\\s+", " ");
  }

  private int[] calcColumnsLength(final List<String[]> rows) {
    final int[] columnsLength = new int[columns.size()];
    for (int i = 0, n = columnsLength.length; i < n; ++i) {
      columnsLength[i] = calcColumnLength(rows, i);
    }
    return columnsLength;
  }

  private int calcColumnLength(final List<String[]> rows, final int index) {
    int length = columns.get(index).length();
    for (final String[] row : rows) {
      if (row == null)
        continue;
      length = Math.max(length, row[index].length());
    }
    return Math.min(length, COLUMN_WRAP_LENGTH);
  }
}