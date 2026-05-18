-- MoonGen example: simple L3 load + latency measurement
-- Usage: ./build/MoonGen benchmarks/moongen/l3_load_latency.lua --dev 0 --dev 1 --rate 10000 --duration 60

local mg = require "moongen"
local device = require "device"
local hist = require "histogram"
local timer = require "timer"
local memory = require "memory"
local proto = require "proto"

function master(args)
    local devRx = device.config{port = args.dev0, rxQueues = 1}
    local devTx = device.config{port = args.dev1, txQueues = 1}
    device.waitForLinks()

    local txQueue = devTx:getTxQueue(0)
    mg.startTask("loadSlave", txQueue, args.rate, args.duration)
    mg.waitForTasks()
end

function loadSlave(queue, rate, duration)
    local mem = memory.createMemPool(function(buf)
        buf:getUdpPacket():fill{ethSrc = "aa:aa:aa:aa:aa:aa", ethDst = "bb:bb:bb:bb:bb:bb",
            ip4Src = "192.0.2.1", ip4Dst = "198.51.100.1", udpSrc = 1234, udpDst = 4321}
    end)

    local bufs = mem:bufArray()
    local t = timer.new(duration)
    local histo = hist:new()

    while t:running() do
        bufs:alloc(64)
        for i = 1, bufs.size do
            local pkt = bufs[i]
            -- Optionally add timestamp in payload for latency
            pkt:getUdpPacket():getUdpPayload():storeUInt64(0, mg.getTime())
        end
        queue:send(bufs)
        mg.sleepMillis(0)
    end
end

-- argument parsing convenience
-- Provide defaults when run via MoonGen examples
if not ... then
    print("Run this with MoonGen: ./build/MoonGen benchmarks/moongen/l3_load_latency.lua --dev0 0 --dev1 1 --rate 10000 --duration 60")
end
