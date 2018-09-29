#!/usr/bin/env bash
docker-compose up -d
aws --endpoint=http://127.0.0.1:5000 \
  ssm put-parameter \
    --name 'testtest' \
    --type 'String' \
    --value '12345678901234567890' > /dev/null
