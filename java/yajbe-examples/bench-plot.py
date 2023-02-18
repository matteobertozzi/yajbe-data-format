#!/usr/bin/env python3
#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import numpy as np
import matplotlib.pyplot as plt
import sys

if len(sys.argv) < 2:
  print('usage: bench-plot.py <result.csv>')
  sys.exit(1)

csv_name = sys.argv[1]
file_name = csv_name[:-4]

data = np.genfromtxt(csv_name, delimiter=',', names=True, dtype=None)
print(data.dtype)

x   = data['Score']
y   = np.arange(len(data['Benchmark']))
err = data['Samples']
labels = []
for test_name, format, dataset_name in zip(data['Benchmark'], data['Param_format'], data['Param_dataSetName']):
  format = format.decode()
  test_name = test_name.decode()
  dataset_name = dataset_name.decode()
  test_name = test_name[len('"io.github.matteobertozzi.yajbe.examples.bench.BenchEncoding.'):-1]
  dataset_name = dataset_name[dataset_name.rfind('/'):]
  labels.append('%s %s %5s' % (dataset_name, test_name, format))

plt.rcdefaults()
plt.barh(y, x, color='blue', ecolor='red', alpha=0.4, align='center')
plt.yticks(y, labels)
plt.xlabel("Performance (ops/s)")
plt.title("Benchmark")
plt.savefig(file_name + '.png', bbox_inches='tight')

