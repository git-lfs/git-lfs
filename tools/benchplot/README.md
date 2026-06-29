# benchplot

A Go tool to parse Go benchmark results and generate plots.

## Building

```bash
cd tools/benchplot
go build -o ../../bin/benchplot .
```

Or from the repository root:

```bash
make benchplot
```

## Usage

Run with default settings (reads `benchmark/sweep_final.txt` and writes plots to `benchmark/`):

```bash
./bin/benchplot
```

Run with custom input file:

```bash
./bin/benchplot -input=/path/to/results.txt
```

Run with custom output directory:

```bash
./bin/benchplot -outdir=/path/to/output -prefix=my_plot
```

## Options

- `-input`: Path to benchmark results file (default: `benchmark/sweep_final.txt`)
- `-outdir`: Output directory for plots (default: `benchmark`)
- `-prefix`: Output file prefix (default: `sweep_plot`)

## Generating Benchmark Data

Run benchmarks and save results:

```bash
go test -bench=. -benchmem -count=3 ./commands > benchmark/sweep_final.txt
```

Or use the Makefile:

```bash
make test-bench > benchmark/sweep_final.txt
```

Then generate plots:

```bash
./bin/benchplot
```

## Output

The tool generates 22 PNG plots showing various benchmark metrics:

- Pattern count sweeps for GetAutoTrackSize
- File size sweeps for clean operations
- Throughput comparisons
- Pattern shape analysis
- And more...

All plots are saved to the `benchmark/` directory (which is gitignored).
