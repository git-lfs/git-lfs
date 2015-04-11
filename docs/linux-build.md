## Building on Linux

## Ubuntu 14.04 (Trusty Tahr)

### Building

```
sudo apt-get install golang-go git

./script/bootstrap
```

That will place a git-lfs binary in the `bin/` directory. Copy the binary to a directory in your path:

```
sudo cp bin/git-lfs /usr/local/bin
```

Try it:

```
[949][rubiojr@octox] git lfs
git-lfs v0.0.1

[~]
[949][rubiojr@octox] git lfs init
git lfs initialized
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

`git help lfs` should show the git-lfs man pages now.
