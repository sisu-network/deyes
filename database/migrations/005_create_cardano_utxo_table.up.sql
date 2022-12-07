CREATE TABLE utxo(id VARCHAR(64), chain  VARCHAR(64), tx_hash VARCHAR(256), tx_index INTEGER, serialized bytea, PRIMARY KEY(id, chain));
