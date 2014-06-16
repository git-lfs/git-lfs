## Building on Linux

## Ubuntu 14.04 (Trusty Tahr)

### Building

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

### Installing the man pages

You'll need ruby and rubygems to install the `ronn` gem:


```
sudo apt-get install ruby build-essential
sudo gem install ronn
./script/man
sudo mkdir -p /usr/local/share/man/man1
sudo cp man/*.1 /usr/local/share/man/man1
```

`git help media` should show the git-media manpage now.
