# Coin-flip economy simulator

CLI tool that loads players and configuration from CSV, runs a sequential coin-flip interaction model at scale (simulation or analytical), and writes one aggregated CSV row.

## Run

From the repository root:

```bash
go run ./cmd/simulator \
  --input input_table.csv \
  --config config_table.csv \
  --output output_results.csv \
  --threshold 1500 \
  --churn-multiplier 1.3 \
  --seed 12345
```

Flags:

| Flag | Meaning |
|------|---------|
| `--input` | Players CSV path |
| `--config` | Config key/value CSV path |
| `--output` | Output CSV path |
| `--threshold` | Player-level total points threshold for `players_above_threshold` |
| `--churn-multiplier` | Optional; default **1.3**. Applied to success probabilities when `about_to_churn` is true |
| `--seed` | Optional deterministic RNG seed for simulation mode |
| `--workers` | Parallel workers (`1` = single-threaded streaming) |
| `--analytical` | Closed-form expectations instead of RNG |

Build a binary:

```bash
go build -o simulator ./cmd/simulator
```

### Churn multiplier behavior

When `about_to_churn` is **true** for a player (and the column exists), each baseline success probability `p_success_n` is multiplied by `--churn-multiplier`, then **clamped to `[0, 1]`**. This models a retention uplift without allowing invalid probabilities.

If the `about_to_churn` column is omitted from the input CSV, churn adjustments are **not** applied (`HasChurnColumn` is false).

### RNG and parallelism

- **`--workers 1`** (default): uses one global `math/rand/v2` stream. With `--seed`, the entire run is reproducible bit-for-bit.
- **`--workers N` with `N > 1`**: each player gets an independent PCG stream derived from `--seed` (or an OS-random base) and the player’s **stable CSV row index**. Different worker counts still produce **different totals** than `--workers 1` because the **global** RNG consumption order no longer matches the single-stream case. Parallel runs are reproducible for fixed `(seed, CSV, workers)` but are **not** numerically identical to single-threaded mode.

### Assumptions

1. **Number of interactions per player** is `floor(rolls_sink / avg_multiplier)` (see `aggregation.FloorInteractions`). Non-finite or non-positive values of the ratio are treated as **zero** interactions.
2. **Points per interaction**:
   - The simulation engine computes **cumulative base points** from config by success depth (sum of `points_0 … points_{k-1}` for `k` successes in that interaction).
   - **Then** the result is multiplied by the player’s `avg_multiplier` to get the final points for that interaction and for aggregation. The engine itself stays free of per-player economy rules; scaling is done in `internal/run` (and mirrored in analytical mode).
3. **`players_above_threshold`**: a player is counted only if their **total** points (sum over all of that player’s interactions) is **strictly greater** than `--threshold` (`>`, not `>=`).
4. **Sequential flips**: each interaction tries flips `0 .. max_successes-1`. Each flip succeeds with probability `p_success_n` (after churn adjustment). On the **first failure**, the interaction stops. Total successes is in `[0, max_successes]`.
5. **Analytical mode** (`--analytical`): uses **expected** values from the closed-form distribution of one interaction (not a path-by-path replay of the Monte Carlo process). It uses the same `floor(rolls_sink / avg_multiplier)` interaction count and the same `avg_multiplier` scaling of expected per-interaction points. Expected bucket contributions are converted to integer counts with `math.Round` when adding to `success_k_count` fields, so the row is a **summary of expectations**, not a single synthetic run.
6. **Parallel mode** may produce different aggregate numbers than a single worker for the same seed (see “RNG and parallelism”): different RNG consumption order, not a different product rule set.

## Project layout

- `cmd/simulator` — CLI entrypoint (flags only)
- `internal/domain` — `Player`, `SimulationConfig`, `Aggregates`
- `internal/input` — streaming player CSV
- `internal/config` — key/value config CSV
- `internal/simulation` — RNG engine + optional analytical distribution
- `internal/aggregation` — counters, merge, interaction helpers
- `internal/output` — aggregated CSV writer
- `internal/run` — orchestration

## Tests and benchmarks

```bash
go test ./...
go test -bench=. -benchmem ./internal/simulation/
```

On Apple M1 Pro (example), `BenchmarkSimulateOneInteraction` is about **30 ns/op with 0 heap allocations** per interaction in the micro-benchmark — hot paths reuse scratch slices for churn-adjusted probabilities.

For large inputs, prefer **`--workers 1`** unless CPU-bound throughput matters; parallelism trades numerical agreement with `--workers 1` (see above) while keeping per-player sub-stream determinism when worker count is fixed.

## Input / config formats

**Players** (`input_table.csv`): header must include `user_id`, `rolls_sink`, `avg_multiplier`. Optional: `about_to_churn` (`true` / `false` / `1` / `0` / `yes`).

**Config** (`config_table.csv`): two columns with headers **`Input`** and **`Value`** (case-insensitive; whitespace around names is OK). The loader also accepts legacy headers **`key`** and **`value`**.

Probabilities can be provided either as percentages (for example, `60%`) or decimals in the `0..1` range (for example, `0.6`).

Assignment format (1-based indices in the file):

- `max_successes` — positive integer `N`
- `p_success_1` … `p_success_N` — success probabilities per depth, each as a decimal in `[0,1]` **or** a percentage like `60%` (meaning `0.6`)
- `points_success_1` … `points_success_N` — non-negative point awards

Legacy format (0-based `p_success_0` … `p_success_{N-1}` and `points_0` … `points_{N-1}`) remains supported when `p_success_0` or `points_0` is present.

Example `config_table.csv` is included in the repo root.
