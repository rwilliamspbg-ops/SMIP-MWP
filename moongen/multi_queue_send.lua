-- multi_queue_send.lua
-- MoonGen script: send UDP packets across multiple TX queues (multi-queue)
-- Usage example:
--   ./build/MoonGen moongen/multi_queue_send.lua --txPort 0 --txQueues 16 --pktSize 64 --flows 16 --rate 10000000000 --duration 30

local mg = require "moongen"
local device = require "device"
local memory = require "memory"
local log = require "log"
local timer = require "timer"

function configure(parser)
    parser:option("--txPort", "Transmit port (DPDK port id)"):default(0)
    parser:option("--txQueues", "Number of TX queues to use"):default(16)
    parser:option("--pktSize", "Packet size in bytes"):default(64)
    parser:option("--flows", "Distinct flows (IP variation)"):default(16)
    parser:option("--rate", "Target line rate in bps (per-queue not enforced)"):default(10000000000)
    parser:option("--duration", "Run duration in seconds"):default(30)
    return parser:parse()
end

function master(args)
    if args.txQueues < 1 then
        log:error("txQueues must be >= 1")
        return
    end
    local dev = device.config{port = args.txPort, txQueues = args.txQueues}
    device.waitForLinks()

    local pktSize = tonumber(args.pktSize)
    local flows = tonumber(args.flows)
    local duration = tonumber(args.duration)

    local mem = memory.createMemPool(function(buf)
        local pkt = buf:getUdpPaddedPacket()
        pkt:fill{
            ethSrc = "52:54:00:12:34:56",
            ethDst = "52:54:00:65:43:21",
            ip4Src = "10.0.0.1",
            ip4Dst = "10.0.0.100",
            udpSrc = 1234,
            udpDst = 1234,
            pktLength = pktSize
        }
    end)

    -- start one sender per TX queue
    for q=0, args.txQueues-1 do
        mg.startTask("sender", dev:getTxQueue(q), mem, pktSize, flows, duration, q)
    end
    mg.waitForTasks()
end

function sender(queue, mem, pktSize, flows, duration, qid)
    local bufs = mem:bufArray()
    local t = timer:new(duration)
    local flowBase = (qid * flows) % 250
    local seq = 0

    while t:running() do
        bufs:alloc(128)
        for i=1, bufs.size do
            local pkt = bufs[i]
            pkt:setSize(pktSize)
            local udp = pkt:getUdpPacket()
            -- vary destination IP among flows to create per-flow hashing spread
            local octet = (flowBase + (seq % flows)) % 250 + 1
            local dstIp = string.format("192.168.%d.%d", qid % 254, octet)
            udp.ip4.dst:set(dstIp)
            udp.udp:setDst(1234)
            seq = seq + 1
        end
        queue:send(bufs)
    end
end
