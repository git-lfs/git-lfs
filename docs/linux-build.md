## Building on Linux

## Ubuntu 14.04 (Trusty Tahr)

```
sudo apt-get install golang-go git

./script/bootstrap
```

That will place a git-media binary in the `bin/` directory. Copy the binary to a directory in your path:

```
sudo cp bin/git-media /usr/local/bin 
```

Try it:

```
[949][rubiojr@octox] git media
git-media v0.0.1

[~]
[949][rubiojr@octox] git media init
git media initialized
```
