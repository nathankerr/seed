syn spell toplevel
syn include @rubyTop syntax/ruby.vim

syn keyword Statement input output table state
syn match Comment /^#.*$/

syn region seedDoBlock start="do" end="end" contains=@rubyTop