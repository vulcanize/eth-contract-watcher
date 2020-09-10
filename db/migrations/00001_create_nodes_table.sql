-- +goose Up
CREATE TABLE nodes (
  id            SERIAL PRIMARY KEY,
  client_name   VARCHAR,
  genesis_block VARCHAR(66),
  network_id    VARCHAR,
  node_id       VARCHAR(128),
  chain_id      INTEGER,
  CONSTRAINT node_uc UNIQUE (genesis_block, network_id, node_id, chain_id)
);

-- +goose Down
DROP TABLE nodes;
