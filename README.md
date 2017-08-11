# Mjolnir

Mjolnir is a simple command-line utility to work with BoltDB from the CLI. I wrote this as I found myself wanting to make quick entries on my terminal without having to work with golang directly. 

## Examples
**Show Buckets**
```bash
shellfu@shellfu:~$ mjolnir -f foobar.db -p
|-------------|
| BUCKET NAME |
|-------------|
| barBucket   |
| fooBucket   |
|-------------|
 total buckets: 2
```

**Print entire bucket to stdout**
```bash
shellfu@shellfu:~$ mjolnir -f foobar.db -b fooBucket -p
|  KEY   |       VALUE       |
|--------|-------------------|
| barKey | "barValue"        |
| fooKey | { "foo" : "bar" } |
 database: foobar.db | bucket: fooBucket | total records: 2
```

**Add a Key Value Pair**
```bash
shellfu@shellfu:~$ mjolnir -f foobar.db -b fooBucket -a -key fooKey -value '{ "foo" : "bar" }'
|  KEY   |     VALUE     |
|--------|---------------|
| fooKey | {"foo":"bar"} |
```

**Delete a Key Value Pair**
```bash
shellfu@shellfu:~$ mjolnir -f foobar.db -b fooBucket -d -key fooKey
```

**Read a Key Value Pair**
```bash
shellfu@shellfu-mac:~$ mjolnir -f foobar.db -b fooBucket -r -key fooKey
|  KEY   |   VALUE    |
|--------|------------|
| fooKey | "fooValue" |
```

**Update a Key Value Pair**
```bash
shellfu@shellfu:~$ mjolnir -f foobar.db -b fooBucket -u -key fooKey -value fooValue
|  KEY   |   VALUE    |
|--------|------------|
| fooKey | "fooValue" |
```
