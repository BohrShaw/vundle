# Vundle

`vundle` is a command line tool for managing Vim packages/plugins/bundles.
Be aware that this is currently a **personalized** tool.

## Usage

    go get -u github.com/bohrshaw/vundle
    vundle -h

1. The bundle list is get by parsing `~/.vim/init..vim` with a simplified
   Vim pattern `Bundle\a*(['"]\zs[^'"]+`.
1. A bundle is of this format: `author/project[:[branch]][/sub/directory]`.
1. Bundles are installed into `~/.vim/bundle`.
1. To generate Vim help tags files, a function `helptags#(overwrite_or_not)`
   must be defined. This function should populate `&runtimepath` as it would be
   invoked like `vim -Nes --cmd 'call helptags#(1/0) | qall!'`.
