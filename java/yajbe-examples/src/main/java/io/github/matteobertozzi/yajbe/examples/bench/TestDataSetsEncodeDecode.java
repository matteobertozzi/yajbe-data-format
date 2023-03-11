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

package io.github.matteobertozzi.yajbe.examples.bench;


import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;

import com.fasterxml.jackson.databind.JsonNode;

import io.github.matteobertozzi.yajbe.examples.util.AbstractTestEncodeDecode;
import io.github.matteobertozzi.yajbe.examples.util.HumansTableView;

public class TestDataSetsEncodeDecode extends AbstractTestEncodeDecode {
  public static void main(final String[] args) throws Exception {
    final ArrayList<TestData> testData = new ArrayList<>();
    foreachTestData(new File("../../test-data/"), (file, node) -> {
      testData.add(new TestData(file.getName(), JsonNode.class, node));
    });
    testData.sort((a, b) -> Long.compare(b.jsonEnc().length, a.jsonEnc().length));


    final HumansTableView results = runEncodeDecode(testData);
    //final HumansTableView results = runSeparateEncodeDecode(testData);
    System.out.println(results.addHumanView(new StringBuilder()));
    Files.writeString(Path.of("test-datasets.csv"), results.toCsv());
  }
}
