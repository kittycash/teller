# kitty-api
Stores kitty info and other stuff.

## Test

**Clone repository:**

```bash
# Ensure directory exists.
mkdir -p ${GOPATH}/src/github.com/kittycash

# Go into directory.
cd ${GOPATH}/src/github.com/kittycash

# Clone.
git clone https://github.com/kittycash/kitty-api.git

# Checkout 'develop' branch.
git checkout develop
```

**Start redis server locally:**

```bash
docker run -d -p 6379:6379 redislabs/redisearch:latest
```

**Run executable (test mode):**

```bash
# Go into directory.
cd ${GOPATH}/src/github.com/kittycash/kitty-api

# Run with environment variable declared.
bash run_test_mode.sh

# Inject some kitties.
go run ${GOPATH}/src/github.com/kittycash/kitty-api/cmd/testcli/testcli.go
```

### docker-compose

As an alternative to the docker/`run.sh` combination above, a `docker-compose.yml` file is provided that automates the process.

Clone the repo and check out the develop branch as above, and then just run:

```bash
docker-compose up
```

**Add some entries for testing:**

*TODO: Complete (cli needed)*
