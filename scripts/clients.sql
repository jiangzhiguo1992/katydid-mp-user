--
-- PostgreSQL database dump
--

-- Dumped from database version 17.2 (Homebrew)
-- Dumped by pg_dump version 17.0

-- Started on 2025-01-21 08:52:56 CST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
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
-- TOC entry 219 (class 1259 OID 24594)
-- Name: client; Type: TABLE; Schema: clients; Owner: jiang
--

CREATE TABLE clients.client (
    id bigint NOT NULL,
    create_at bigint NOT NULL,
    update_at bigint NOT NULL,
    delete_at bigint,
    ip integer NOT NULL,
    part integer NOT NULL,
    enable boolean NOT NULL,
    online_at bigint NOT NULL,
    offline_at bigint NOT NULL,
    organization text NOT NULL,
    extra jsonb NOT NULL
)
WITH (fillfactor='90');


ALTER TABLE clients.client OWNER TO jiang;

--
-- TOC entry 218 (class 1259 OID 24593)
-- Name: client_id_seq; Type: SEQUENCE; Schema: clients; Owner: jiang
--

CREATE SEQUENCE clients.client_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE clients.client_id_seq OWNER TO jiang;

--
-- TOC entry 3695 (class 0 OID 0)
-- Dependencies: 218
-- Name: client_id_seq; Type: SEQUENCE OWNED BY; Schema: clients; Owner: jiang
--

ALTER SEQUENCE clients.client_id_seq OWNED BY clients.client.id;


--
-- TOC entry 3536 (class 2604 OID 24597)
-- Name: client id; Type: DEFAULT; Schema: clients; Owner: jiang
--

ALTER TABLE ONLY clients.client ALTER COLUMN id SET DEFAULT nextval('clients.client_id_seq'::regclass);


--
-- TOC entry 3539 (class 2606 OID 24601)
-- Name: client idx_ip_part; Type: CONSTRAINT; Schema: clients; Owner: jiang
--

ALTER TABLE ONLY clients.client
    ADD CONSTRAINT idx_ip_part UNIQUE (ip, part) WITH (fillfactor='100');


--
-- TOC entry 3544 (class 2606 OID 24599)
-- Name: client pk; Type: CONSTRAINT; Schema: clients; Owner: jiang
--

ALTER TABLE ONLY clients.client
    ADD CONSTRAINT pk PRIMARY KEY (id) WITH (fillfactor='100');


--
-- TOC entry 3537 (class 1259 OID 24604)
-- Name: idx_enable; Type: INDEX; Schema: clients; Owner: jiang
--

CREATE INDEX idx_enable ON clients.client USING btree (enable DESC NULLS LAST) WITH (deduplicate_items='true', fillfactor='99');


--
-- TOC entry 3540 (class 1259 OID 24607)
-- Name: idx_offlineAt; Type: INDEX; Schema: clients; Owner: jiang
--

CREATE INDEX "idx_offlineAt" ON clients.client USING btree (offline_at DESC NULLS LAST) WITH (fillfactor='90', deduplicate_items='false');


--
-- TOC entry 3541 (class 1259 OID 24608)
-- Name: idx_onlineAt; Type: INDEX; Schema: clients; Owner: jiang
--

CREATE INDEX "idx_onlineAt" ON clients.client USING btree (online_at DESC) WITH (fillfactor='90', deduplicate_items='false');


--
-- TOC entry 3542 (class 1259 OID 24609)
-- Name: idx_organization; Type: INDEX; Schema: clients; Owner: jiang
--

CREATE INDEX idx_organization ON clients.client USING btree (organization) WITH (fillfactor='99', deduplicate_items='true');


-- Completed on 2025-01-21 08:52:56 CST

--
-- PostgreSQL database dump complete
--

