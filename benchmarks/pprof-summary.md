# pprof summary

## pprof-afxdp-top.txt
```
File: afxdp.test
Build ID: adedca7c136cf5120d0983c07870f9d4876c5ca5
Type: cpu
Time: 2026-05-15 01:40:42 UTC
Duration: 4.72s, Total samples = 5920ms (125.51%)
Showing nodes accounting for 3580ms, 60.47% of 5920ms total
Dropped 108 nodes (cum <= 29.60ms)
Showing top 20 nodes out of 137
      flat  flat%   sum%        cum   cum%
     650ms 10.98% 10.98%     2260ms 38.18%  runtime.selectgo
     590ms  9.97% 20.95%      590ms  9.97%  runtime.nanotime (inline)
     480ms  8.11% 29.05%      610ms 10.30%  runtime.unlock2
     270ms  4.56% 33.61%      300ms  5.07%  runtime.lock2
     170ms  2.87% 36.49%      170ms  2.87%  runtime.usleep
     150ms  2.53% 39.02%     2660ms 44.93%  smip-mwp/internal/datapath/afxdp.(*Forwarder).RunXDPLoop
     140ms  2.36% 41.39%      140ms  2.36%  runtime.futex
     130ms  2.20% 43.58%      130ms  2.20%  runtime.memclrNoHeapPointers
     120ms  2.03% 45.61%      250ms  4.22%  runtime.sellock
     100ms  1.69% 47.30%      100ms  1.69%  runtime.(*mLockProfile).store (inline)
      90ms  1.52% 48.82%      710ms 11.99%  runtime.mallocgc
      90ms  1.52% 50.34%      590ms  9.97%  runtime.mallocgcSmallScanNoHeader
      80ms  1.35% 51.69%      110ms  1.86%  runtime.(*mspan).writeHeapBitsSmall
      80ms  1.35% 53.04%       80ms  1.35%  runtime.nextFreeFast (inline)
      80ms  1.35% 54.39%      160ms  2.70%  smip-mwp/internal/routing.(*Table).LookupNextHop
      80ms  1.35% 55.74%      100ms  1.69%  smip-mwp/internal/wire.ParseHeader
      70ms  1.18% 56.93%       70ms  1.18%  internal/runtime/atomic.(*Uint32).CompareAndSwap (inline)
      70ms  1.18% 58.11%      200ms  3.38%  runtime.(*timer).modify
      70ms  1.18% 59.29%       70ms  1.18%  runtime.(*timer).needsAdd (inline)
      70ms  1.18% 60.47%      420ms  7.09%  runtime.makechan
```
## pprof-crypto-top.txt
```
File: crypto.test
Build ID: dcfec2ff23f858f18f4c7177b6f3e236dc674851
Type: cpu
Time: 2026-05-15 01:40:17 UTC
Duration: 23.53s, Total samples = 30.13s (128.05%)
Showing nodes accounting for 24.08s, 79.92% of 30.13s total
Dropped 218 nodes (cum <= 0.15s)
Showing top 20 nodes out of 107
      flat  flat%   sum%        cum   cum%
     8.42s 27.95% 27.95%      8.42s 27.95%  crypto/internal/fips140/aes/gcm.gcmAesEnc
     5.09s 16.89% 44.84%      5.09s 16.89%  crypto/internal/fips140/aes/gcm.gcmAesDec
     2.83s  9.39% 54.23%      2.83s  9.39%  runtime.memmove
     1.23s  4.08% 58.31%      1.23s  4.08%  runtime.memclrNoHeapPointers
     0.75s  2.49% 60.80%      1.04s  3.45%  runtime.typePointers.next
     0.69s  2.29% 63.09%      0.79s  2.62%  runtime.findObject
     0.62s  2.06% 65.15%      0.62s  2.06%  runtime.futex
     0.61s  2.02% 67.18%      0.61s  2.02%  runtime.procyieldAsm
     0.60s  1.99% 69.17%      2.81s  9.33%  runtime.scanObject
     0.58s  1.92% 71.09%      0.58s  1.92%  crypto/internal/fips140/aes.encryptBlockAsm
     0.48s  1.59% 72.69%      0.48s  1.59%  runtime.madvise
     0.48s  1.59% 74.28%      2.29s  7.60%  runtime.sweepone
     0.25s  0.83% 75.11%      1.34s  4.45%  runtime.(*mcentral).cacheSpan
     0.23s  0.76% 75.87%      1.55s  5.14%  runtime.(*sweepLocked).sweep
     0.22s  0.73% 76.60%      0.23s  0.76%  runtime.(*fixalloc).alloc
     0.22s  0.73% 77.33%      0.22s  0.73%  runtime.typePointers.nextFast (inline)
     0.20s  0.66% 78.00%      0.20s  0.66%  encoding/binary.bigEndian.Uint64
     0.20s  0.66% 78.66%      0.21s   0.7%  runtime.(*mspan).init
     0.19s  0.63% 79.29%      0.32s  1.06%  runtime.(*spanSet).push
     0.19s  0.63% 79.92%      0.41s  1.36%  runtime.mallocgcTiny
```
