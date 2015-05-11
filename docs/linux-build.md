## Building on Linux

## Ubuntu 14.04 (Trusty Tahr)

### Building

```
sudo apt-get install golang-go git
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
gvm install go1.4.2 # or something 1.3.1 or newer
gvm use go1.4.2
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
sudo apt-get install build-essential
sudo apt-get install ruby2.0 ruby2.0-dev
sudo rm /usr/bin/ruby && sudo ln -s /usr/bin/ruby2.0 /usr/bin/ruby
sudo rm -fr /usr/bin/gem && sudo ln -s /usr/bin/gem2.0 /usr/bin/gem
sudo gem install ronn
./script/man
sudo mkdir -p /usr/local/share/man/man1
sudo cp man/*.1 /usr/local/share/man/man1
```

`git help lfs` should show the git-lfs man pages now.
