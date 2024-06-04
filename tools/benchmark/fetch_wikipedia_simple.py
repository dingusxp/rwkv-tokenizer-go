#!/usr/bin/env python3
import json, sys, tqdm
from datasets import load_dataset

ds = load_dataset("wikipedia", "20220301.simple")

with open('wikipedia_simple.jsonl', 'w') as fp:
  for row in tqdm.tqdm(ds["train"]):
    fp.write(json.dumps({'text':row["text"]}, ensure_ascii=False) + '\n')
