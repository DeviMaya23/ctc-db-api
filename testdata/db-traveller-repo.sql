CREATE TABLE public.m_influence (
	id serial4 NOT NULL,
	"name" varchar(50) NOT NULL,
	created_at timestamptz NULL,
	created_by varchar(50) NULL,
	updated_at timestamptz NULL,
	updated_by varchar(50) NULL,
	deleted_at timestamptz NULL,
	deleted_by varchar(50) NULL,
	CONSTRAINT m_influence_pk PRIMARY KEY (id)
);

CREATE TABLE public.m_accessory (
	id serial4 NOT NULL,
	"name" varchar(50) NOT NULL,
	hp int4 NULL,
	sp int4 NULL,
	patk int4 NULL,
	pdef int4 NULL,
	eatk int4 NULL,
	edef int4 NULL,
	spd int4 NULL,
	crit int4 NULL,
	effect text NULL,
	created_at timestamptz NULL,
	created_by varchar(50) NULL,
	updated_at timestamptz NULL,
	updated_by varchar(50) NULL,
	deleted_at timestamptz NULL,
	deleted_by varchar(50) NULL,
	CONSTRAINT m_accessory_pk PRIMARY KEY (id)
);


CREATE TABLE public.m_traveller (
	id serial NOT NULL,
	"name" varchar(50) NOT NULL,
	rarity int2 NOT NULL,
	influence_id int4 NULL,
	created_at timestamptz NULL,
	created_by varchar(50) NULL,
	updated_at timestamptz NULL,
	updated_by varchar(50) NULL,
	deleted_at timestamptz NULL,
	deleted_by varchar(50) NULL,
	accessory_id int4 NULL,
	CONSTRAINT m_traveller_pk PRIMARY KEY (id)
);



-- public.m_traveller foreign keys

ALTER TABLE public.m_traveller ADD CONSTRAINT m_traveller_m_influence_fk FOREIGN KEY (influence_id) REFERENCES public.m_influence(id);
ALTER TABLE public.m_traveller ADD CONSTRAINT m_traveller_m_accessory_fk FOREIGN KEY (accessory_id) REFERENCES public.m_accessory(id);


INSERT INTO public.m_influence
("name")
VALUES('Wealth');
INSERT INTO public.m_influence
("name")
VALUES('Power');
INSERT INTO public.m_influence
("name")
VALUES('Fame');
INSERT INTO public.m_influence
("name")
VALUES('Opulence');
INSERT INTO public.m_influence
("name")
VALUES('Dominance');
INSERT INTO public.m_influence
("name")
VALUES('Prestige');

INSERT INTO public.m_traveller
("name", rarity, influence_id, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
VALUES('Fiore', 5, 3, NULL, NULL, NULL, NULL, NULL, NULL);