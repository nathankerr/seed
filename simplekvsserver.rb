require 'rubygems'
require 'bud'

class SimpleKvsServer
  include Bud

  state do
    # kvdel <~ [["127.0.0.1:3000", 1]]
    channel :kvdel, [:@address, :key]

    # kvput <~ [["127.0.0.1:3000", 1, 2],["127.0.0.1:3000", 3, 4]]
    channel :kvput, [:@address, :key] => [:value]

    # kvget <~ [["127.0.0.1:3000", "127.0.0.1:46637", 1]]
    channel :kvget, [:@address, :client, :key]
    channel :kvget_response, [:@client, :key] => [:value]

    table :kvstate, [:key] => [:value]
  end

  bloom do
    kvstate <+-  kvput.payloads
    kvstate <- (kvstate * kvdel).lefts(:key => :key)

    kvget_response <~ (kvget * kvstate).pairs(:key => :key) do |g, s| [g.client, s.key, s.value] end
  end

  # added to show state change
  bloom :output do
    stdio <~ kvstate.inspected
  end
end


## Below added by hand
program = SimpleKvsServer.new(:ip => "127.0.0.1", :port => 3000, :dump_rewrite => true)
program.run_fg