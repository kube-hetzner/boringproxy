/*
Copyright 2024 kube-hetzner

This file is dual-licensed under the MIT License and the Apache License, Version 2.0.
You may choose either license to govern your use of this file.

Apache License, Version 2.0 (the "Apache License"):
    You may not use this file except in compliance with the Apache License.
    You may obtain a copy of the Apache License at:
        http://www.apache.org/licenses/LICENSE-2.0

MIT License (the "MIT License"):
    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:
        The above copyright notice and this permission notice shall be included in
        all copies or substantial portions of the Software.

Unless required by applicable law or agreed to in writing, software distributed
under either license is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the respective licenses
for the specific language governing permissions and limitations under those licenses.
*/

package main

import (
	"os"
)

// Gets environment variable with optional default value
func getEnv(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if value == "" && len(defaultValue) != 0 {
		return defaultValue[0]
	}
	return value
}
