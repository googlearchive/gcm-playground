# Copyright 2015 Google Inc. All Rights Reserved.
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

# init.sh
# Install dependencies for the web server to run.
# Make sure you have the following installed:
#   - Node.js, npm
#   - git
#   - Docker
#   - Docker Compose


echo "==> Working in web/"
cd web/

echo "==> Install gulp"
if hash gulp 2>/dev/null; then
  echo "gulp installed."
else
    sudo npm install -g gulp
fi

echo "==> Install bower"
if hash gulp 2>/dev/null; then
  echo "bower installed."
else
    sudo npm install -g bower
fi

echo "==> Install npm dependencies"
npm install

echo "==> Install bower dependencies"
bower install --allow-root

echo "==> Running gulp for building"
gulp

cd ..

echo "==> Starting Docker containers"
docker-compose up
