# Vundle

`vundle` is a command line tool for managing Vim packages/plugins/bundles.

## Usage

    go get -u github.com/bohrshaw/vundle
    vundle -h

1. `vundle` get the bundle list by running an Vim instance.
1. Thus On Vim's side, there should be a variable `g:dundles` specifying the bundle list.
1. A bundle has this format: `author/project[:[branch]][/sub/directory]`.
1. A variable `g:vundle` is defined when Vim is run by `vundle`.
1. Bundles are installed into `~/.vim/bundle`.
1. To generate Vim help tags files, a function `helptags#(overwrite)` must be defined.
    This function should populate `&runtimepath` as it would be invoked like:
    `vim -Nes --cmd 'call helptags#(1/0) | qall!'`
