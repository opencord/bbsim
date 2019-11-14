/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"github.com/opencord/voltha-protos/v2/go/openolt"
	"strconv"
)

func OnuSnToString(sn *openolt.SerialNumber) string {
	s := string(sn.VendorId)
	for _, i := range sn.VendorSpecific {
		s = s + strconv.FormatInt(int64(i/16), 16) + strconv.FormatInt(int64(i%16), 16)
	}
	return s
}
