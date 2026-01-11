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
("name", rarity, influence_id, job_id, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
VALUES('Fiore', 5, 3, 1, NULL, NULL, NULL, NULL, NULL, NULL);


INSERT INTO public.m_user
(username, "password", "token", created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
VALUES('isla', '$2a$14$D0JCfJ3BGUiLVfb0NkOdqudJFe64umW6nmcOl4nSloWtJkIb0M8QW', '1', NULL, NULL, NULL, NULL, NULL, NULL);
