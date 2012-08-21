syn spell toplevel
syn include @rubyTop syntax/ruby.vim

syn keyword Type input output table state
syn match Comment /^#.*$/

syn region DoBlock start="do" end="end" contains=@rubyTop