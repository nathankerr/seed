require 'rubygems'
require 'bud'
require './build/timeserver.rb'
require 'active_support/time'

class Server
	include Bud
	include TimeServer

	def current_time_in(timezone)
		ActiveSupport::TimeZone['America/New_York'].at(Time.now).strftime('%F %T.%6N %z %Z')
	end
end

program = Server.new(:ip => "127.0.0.1", :port => 3000)

program.run_fg