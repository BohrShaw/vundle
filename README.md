# Vundle

`vundle` is a command line tool for managing Vim packages/plugins/bundles.

## Usage

    go get -u github.com/bohrshaw/vundle
    vundle -h

1. `vundle` get the bundle list by running an Vim instance.
2. Thus On Vim's side, there should be a variable `g:dundles` specifying the bundle list.
3. A bundle has this format: `author/project[:[branch]][/sub/directory]`.
4. A variable `g:vundle` is defined when Vim is run by `vundle`.
5. Bundles are installed into `~/.vim/bundle`.
