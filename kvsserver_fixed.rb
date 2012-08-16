require 'rubygems'
require 'bud'

class KvsServer
  include Bud

  state do
    channel :kvdel_channel, [:@address, :key] #kvs.seed:4
    channel :kvget_response_channel, [:@address, :reqid] => [:key, :value] #kvs.seed:6
    scratch :kvput, [:key] => [:value] #kvs.seed:3
    scratch :kvget, [:key] => [:client, :reqid] #kvs.seed:5
    table :kvstate, [:key] => [:value] #kvs.seed:8
    channel :kvput_channel, [:@address, :key] => [:value] #kvs.seed:3
    # channel :kvget_channel, [:@address, :key] => [:reqid] #kvs.seed:5
    channel :kvget_channel, [:@address, :key] => [:client, :reqid] #kvs.seed:5
    scratch :kvdel, [:key] #kvs.seed:4
    # scratch :kvget_response, [:reqid] => [:key, :value] #kvs.seed:6
    scratch :kvget_response, [:client, :reqid] => [:key, :value] #kvs.seed:6
  end

  bloom do
    kvdel <= kvdel_channel.payloads #kvs.seed:4
    kvput <= kvput_channel.payloads #kvs.seed:3
    kvget <= kvget_channel.payloads #kvs.seed:5
    # kvget_response_channel <~ kvget_response.payloads #kvs.seed:6
    kvget_response_channel <~ kvget_response #kvs.seed:6
    kvstate <+-  kvput #kvs.seed:11
    # kvget_response <= (kvget * kvstate).pairs do |g, t| [g.reqid, t.key, t.value] end #kvs.seed:14
    kvget_response <= (kvget * kvstate).pairs do |g, t|
  [g.client, g.reqid, t.key, t.value]
end #kvs.seed:14
    # kvstate <-  kvdel #kvs.seed:19 ##FIXED##
    kvstate <- (kvstate * kvdel).matches do |s, d| s end
  end

  # added to show state change
  bloom :output do
    stdio <~ kvstate.inspected
  end
end

## Below added by hand
program = KvsServer.new(:ip => "127.0.0.1", :port => 3000)
program.run_fg
