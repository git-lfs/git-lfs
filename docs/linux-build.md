## Building on Linux

## Ubuntu 14.04 (Trusty Tahr)

### Building

```
sudo apt-get install golang-go git

./script/bootstrap
```

That will place a git-hawser binary in the `bin/` directory. Copy the binary to a directory in your path:

```
sudo cp bin/git-hawser /usr/local/bin
```

Try it:

```
[949][rubiojr@octox] git hawser
git-hawser v0.0.1

[~]
[949][rubiojr@octox] git hawser init
git hawser initialized
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

`git help hawser` should show the git-hawser manpage now.
