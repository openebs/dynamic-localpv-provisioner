# Copyright 2020-2021 The OpenEBS Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

cd datadir
error_check()
{
	if [ $? != 0 ]
	then
		echo "error caught"
		exit 1
	fi
}

perform_data_write()
{
        value=space_left
	i=0
	while [ $i -le $value ]
	do
		ls -all
		error_check
		touch file$i
		error_check
		dd if=/dev/urandom of=file$i bs=4k count=5000
		error_check
		sync
		read_data
		i=$(( i + 1 ))
		error_check
	done
}
read_data()
{
	touch testfile
	error_check
	echo "OpenEBS Released newer version"
	error_check
	cat testfile
	error_check
	rm testfile
}
perform_data_write
