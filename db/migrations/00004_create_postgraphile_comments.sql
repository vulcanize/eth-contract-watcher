-- +goose Up
COMMENT ON TABLE public.nodes IS E'@name NodeInfo';
COMMENT ON COLUMN public.nodes.node_id IS E'@name ChainNodeID';
COMMENT ON COLUMN public.headers.node_id IS E'@name HeaderNodeID';
