# benchmark

This is a simple benchmark utility that reads either a JSONL dataset or
a dataset consisting of documents separated by `\0` (i.e., the null terminator),
and outputs statistics.

The default testing dataset is the Simple English Wikipedia, which you can
download using the included `fetch_wikipedia_simple.py` script. Make sure you have
the Huggingface `datasets` package installed and updated.

## Example Output

```
--- final stats ---
Tokens:   53619552 | Bytes:    216352627 | Elapsed:         8.474501552s
Elapsed sec:     8.4740
Bytes/token:       4.03
Tokens/sec:  6327537.41
Bytes/sec:   25531346.12
--- ----------- ---
```
