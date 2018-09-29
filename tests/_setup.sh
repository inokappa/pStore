#!/usr/bin/env bash
moto_server ssm --port=5000 > /dev/null 2>&1 &
aws --endpoint=http://127.0.0.1:5000 \
  ssm put-parameter \
    --name 'testtest' \
    --type 'String' \
    --value '12345678901234567890' > /dev/null
