# E2E tests

## Insert single user via API

```bash
BASE_DN=your.domain.here.com
curl -k --noproxy '*' -X POST -d @sample_user.json https://feeder.$BASE_DN/users
```

## Performance tests

Performance tests of Feeder API

```bash
BASE_DN=your.domain.here.com
hey -cpus 1 -n 10000 -c 1 -D sample_user.json -m POST https://feeder.$BASE_DN/users
```

Performance tests of Web API

```bash
BASE_DN=your.domain.here.com
hey -z 5m https://web-api.$BASE_DN/api/users?field=city\&l=100\&query=wroclaw\&s=0
```
