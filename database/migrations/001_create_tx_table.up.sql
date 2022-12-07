CREATE TABLE transactions(chain VARCHAR(64), tx_hash VARCHAR(256), block_height BIGINT, tx_bytes bytea, PRIMARY KEY (chain, tx_hash, block_height));
