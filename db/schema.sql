--
-- PostgreSQL database dump
--

-- Dumped from database version 12.1
-- Dumped by pg_dump version 12.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: checked_headers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.checked_headers (
    id integer NOT NULL,
    header_id integer NOT NULL
);


--
-- Name: checked_headers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.checked_headers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: checked_headers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.checked_headers_id_seq OWNED BY public.checked_headers.id;


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.goose_db_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.goose_db_version_id_seq OWNED BY public.goose_db_version.id;


--
-- Name: headers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.headers (
    id integer NOT NULL,
    hash character varying(66),
    block_number bigint,
    raw jsonb,
    block_timestamp numeric,
    check_count integer DEFAULT 0 NOT NULL,
    node_id integer NOT NULL,
    eth_node_fingerprint character varying(128)
);


--
-- Name: headers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.headers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: headers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.headers_id_seq OWNED BY public.headers.id;


--
-- Name: nodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nodes (
    id integer NOT NULL,
    client_name character varying,
    genesis_block character varying(66),
    network_id character varying,
    node_id character varying(128)
);


--
-- Name: nodes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nodes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nodes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nodes_id_seq OWNED BY public.nodes.id;


--
-- Name: checked_headers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.checked_headers ALTER COLUMN id SET DEFAULT nextval('public.checked_headers_id_seq'::regclass);


--
-- Name: goose_db_version id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version ALTER COLUMN id SET DEFAULT nextval('public.goose_db_version_id_seq'::regclass);


--
-- Name: headers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.headers ALTER COLUMN id SET DEFAULT nextval('public.headers_id_seq'::regclass);


--
-- Name: nodes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes ALTER COLUMN id SET DEFAULT nextval('public.nodes_id_seq'::regclass);


--
-- Name: checked_headers checked_headers_header_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.checked_headers
    ADD CONSTRAINT checked_headers_header_id_key UNIQUE (header_id);


--
-- Name: checked_headers checked_headers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.checked_headers
    ADD CONSTRAINT checked_headers_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: headers headers_block_number_hash_eth_node_fingerprint_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.headers
    ADD CONSTRAINT headers_block_number_hash_eth_node_fingerprint_key UNIQUE (block_number, hash, eth_node_fingerprint);


--
-- Name: headers headers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.headers
    ADD CONSTRAINT headers_pkey PRIMARY KEY (id);


--
-- Name: nodes node_uc; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes
    ADD CONSTRAINT node_uc UNIQUE (genesis_block, network_id, node_id);


--
-- Name: nodes nodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nodes
    ADD CONSTRAINT nodes_pkey PRIMARY KEY (id);


--
-- Name: headers_block_number; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX headers_block_number ON public.headers USING btree (block_number);


--
-- Name: headers_block_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX headers_block_timestamp ON public.headers USING btree (block_timestamp);


--
-- Name: checked_headers checked_headers_header_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.checked_headers
    ADD CONSTRAINT checked_headers_header_id_fkey FOREIGN KEY (header_id) REFERENCES public.headers(id) ON DELETE CASCADE;


--
-- Name: headers headers_node_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.headers
    ADD CONSTRAINT headers_node_id_fkey FOREIGN KEY (node_id) REFERENCES public.nodes(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

