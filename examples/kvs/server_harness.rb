require 'rubygems'
require 'bud'
require './build/kvsserver.rb'

class Server
	include Bud
	include KvsServer

	# added to show state change
	bloom :output do
	  stdio <~ kvstate.inspected
	end
end

program = Server.new(:ip => "127.0.0.1", :port => 3000)

# program.kvstate_replicants <= [["127.0.0.1:3001"]]

program.run_fg