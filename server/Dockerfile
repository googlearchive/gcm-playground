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

FROM golang:1.4.2-wheezy


# -----------------
# Install dependencies
# -----------------

RUN apt-get update && apt-get install -y build-essential


# -----------------
# Install Go dependencies
# -----------------

RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/jinzhu/gorm
RUN go get -u github.com/mattn/go-sqlite3
RUN go get -u github.com/googollee/go-socket.io
RUN go get -u github.com/rs/cors
RUN go get -u github.com/google/go-gcm

# -----------------
# Copy files over
# -----------------

RUN mkdir -p /src/gcm-playground/server/
ADD . /src/gcm-playground/server/
WORKDIR /src/gcm-playground/server/


# -----------------
# Set the server
# -----------------

EXPOSE 4260
