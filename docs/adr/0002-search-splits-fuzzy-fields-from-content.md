# Search matches short fields with fzf and Content through FTS5

Title, Description, and Tags are fuzzy-matched in memory with fzf, and Content goes through an FTS5 index with the trigram tokenizer. The short fields are cheap to scan even across many Snippets, and fuzzy matching is what makes them worth searching. Content is the field that grows, so it's the one an index pays for. Trigram keeps today's substring match, so "son" still finds "json". The default tokenizer matches whole words only.

## Considered Options

- **FTS5 for every field.** This loses fuzzy matching on Title, Description, and Tags, because FTS5 has no subsequence match. bm25 ranks would also replace the fzf scores the weight table is built for.
- **Everything in memory.** This is what `memsearch` does until the FTS5 index exists, and it's the stand-in for Content until then.

## Consequences

- FTS5 replaces only the Content half of `memsearch`. The weight table stays in `domain`, and moving search ranking into the adapter ([#77](https://github.com/DannyFestor/TuiSnip/issues/77), part 2) isn't needed.
- The trigram index can't serve queries shorter than three characters. How those queries match Content, and how smart case and the Content score carry over, is left to the FTS5 ticket.
