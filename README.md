# Vundle

`vundle` is a command line tool for managing Vim packages/plugins/bundles. It's
fast to clone/update/clean bundles. See `vundle -h` for features.

Be aware that this is currently a **personalized** tool.

## Specification

1. The bundle list is get by parsing files like `~/.vim/init?.vim` for lines
   matching `Bundle\w*\(['"]`, and in which lines extracting the strings like
   `[domain.com(/|:)]author/project[:[branch]][/sub/directory]`.
1. Bundles are installed into `~/.vim/bundle`.
1. To generate Vim help tags files, a function `helptags#(overwrite_or_not)`
   must be defined. This function should populate `&runtimepath` as it would be
   invoked like `vim -Nes --cmd 'call helptags#(1/0) | qall!'`.
