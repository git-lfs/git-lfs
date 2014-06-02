# Git Media Commands

## Init

    $ git media init

Sets up global Git configuration settings for Git Media.

## Sync

    $ git media sync

This pushes queued files in `.git/media` to the Git Media server.

## Paths

    $ git media paths add "*.zip"

    $ git media paths remove "*.zip"

    $ git media paths

This manages the media paths in the `.gitattributes` file.
