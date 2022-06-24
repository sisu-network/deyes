CREATE TABLE utxo(id VARCHAR(64), chain  VARCHAR(64), tx_hash VARCHAR(256), tx_index INTEGER, serialized BLOB, PRIMARY KEY(id, chain));
