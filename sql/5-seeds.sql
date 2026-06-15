-- Seed data : catégories classiques FFJ
-- age_max NULL = pas de limite supérieure (Veteran)
-- weight_max NULL = catégorie ouverte (+)

-- ============================================================
-- Utilisateurs de test
-- Mots de passe (bcrypt cost 10) :
--   superadmin@clubmanager.fr  →  Superadmin123!
--   user@clubmanager.fr        →  User123!
-- ============================================================
INSERT INTO users (username, email, password, phonenumber, is_valid, is_superadmin) VALUES
  (
    'superadmin',
    'superadmin@clubmanager.fr',
    '$2a$10$ouW.kYMhzPdrTNCrDkER5.zgBS6.ydAqjh0brDY6rQ2i8LFh6Lh66',
    NULL,
    true,
    true
  ),
  (
    'testuser',
    'user@clubmanager.fr',
    '$2a$10$LSJN9usTiG2R/BIFJr40e.KN25DatoD.wpOE66AlbLbzd6HjZYXZO',
    '+33 6 12 34 56 78',
    true,
    false
  ),
  (
    'testuser2',
    'user2@clubmanager.fr',
    '$2a$10$6UMuLiTZ8RXeNZnuap382epnaaBuAKnlvzIrnkwNHJ3hpndIu1I42',
    '+33 6 98 76 54 32',
    true,
    false
  )
ON CONFLICT (email) DO NOTHING;

-- ============================================================
-- Club de test : A3M Mitry Mory
-- Statut active (SIREN vérifié manuellement)
-- ============================================================
INSERT INTO clubs (siren, name, address, city, postal_code, country, phonenumber, status)
VALUES (
  '382048650',
  'A3M',
  '46 AVENUE JEAN JAURES',
  'Mitry-Mory',
  '77290',
  'FRANCE',
  '+33 1 60 26 00 00',
  'active'
)
ON CONFLICT (siren) DO NOTHING;

-- ============================================================
-- Membre : Redwane Bouselham — profil principal de testuser
-- ============================================================
INSERT INTO members (user_id, firstname, lastname, birthdate, gender, is_primary)
SELECT
  u.id,
  'Redwane',
  'Bouselham',
  '1996-08-30',
  'man',
  true
FROM users u
WHERE u.email = 'user@clubmanager.fr'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Membership : Redwane est membre validé de A3M
-- ============================================================
INSERT INTO club_memberships (member_id, club_id, is_valid)
SELECT
  m.id,
  c.id,
  true
FROM members m
JOIN users u ON u.id = m.user_id
JOIN clubs c ON c.siren = '382048650'
WHERE u.email = 'user@clubmanager.fr'
  AND m.firstname = 'Redwane'
  AND m.lastname  = 'Bouselham'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Membre : Mehdi Bouselham — profil principal de testuser2
-- ============================================================
INSERT INTO members (user_id, firstname, lastname, birthdate, gender, is_primary)
SELECT
  u.id,
  'Mehdi',
  'Bouselham',
  '2000-01-01',
  'man',
  true
FROM users u
WHERE u.email = 'user2@clubmanager.fr'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Membre : Adam Bouselham — profil secondaire de testuser2
-- ============================================================
INSERT INTO members (user_id, firstname, lastname, birthdate, gender, is_primary)
SELECT
  u.id,
  'Adam',
  'Bouselham',
  '2020-04-14',
  'man',
  false
FROM users u
WHERE u.email = 'user2@clubmanager.fr'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Membership : Adam est membre validé de A3M
-- ============================================================
INSERT INTO club_memberships (member_id, club_id, is_valid)
SELECT m.id, c.id, true
FROM members m
JOIN users u ON u.id = m.user_id
JOIN clubs c ON c.siren = '382048650'
WHERE u.email = 'user2@clubmanager.fr' AND m.firstname = 'Adam'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Licence : Redwane — saison 2025-2026
-- ============================================================
INSERT INTO licences (member_id, licence_number, valid_from, valid_until, status)
SELECT m.id, 'LIC-REDWANE-2526', '2025-09-01', '2026-08-31', 'active'
FROM members m
JOIN users u ON u.id = m.user_id
WHERE u.email = 'user@clubmanager.fr' AND m.firstname = 'Redwane'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Licence : Mehdi — saison 2025-2026
-- ============================================================
INSERT INTO licences (member_id, licence_number, valid_from, valid_until, status)
SELECT m.id, 'LIC-MEHDI-2526', '2025-09-01', '2026-08-31', 'active'
FROM members m
JOIN users u ON u.id = m.user_id
WHERE u.email = 'user2@clubmanager.fr' AND m.firstname = 'Mehdi'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Membership : Mehdi est membre validé de A3M
-- ============================================================
INSERT INTO club_memberships (member_id, club_id, is_valid)
SELECT m.id, c.id, true
FROM members m
JOIN users u ON u.id = m.user_id
JOIN clubs c ON c.siren = '382048650'
WHERE u.email = 'user2@clubmanager.fr' AND m.firstname = 'Mehdi'
ON CONFLICT DO NOTHING;

-- ============================================================
-- Rôle : Mehdi est associate dans A3M
-- ============================================================
INSERT INTO roles (user_id, club_id, role)
SELECT u.id, c.id, 'associate'
FROM users u, clubs c
WHERE u.email = 'user2@clubmanager.fr' AND c.siren = '382048650'
ON CONFLICT (user_id, club_id, role) DO NOTHING;

-- ============================================================
-- Rôle : testuser est président de A3M
-- ============================================================
INSERT INTO roles (user_id, club_id, role)
SELECT
  u.id,
  c.id,
  'president'
FROM users u, clubs c
WHERE u.email = 'user@clubmanager.fr'
  AND c.siren  = '382048650'
ON CONFLICT (user_id, club_id, role) DO NOTHING;


INSERT INTO judo_categories (age_group, gender, age_min, age_max, weight_label, weight_max) VALUES
-- Eveil Mixte (4-5 ans)
('Eveil', 'mixed', 4, 5, 'Toutes', NULL),

-- Poussinet Mixte (6-7 ans)
('Poussinet', 'mixed', 6, 7, 'Toutes', NULL),

-- Poussin Mixte (7-8 ans)
('Poussin', 'mixed', 7, 8, 'Toutes', NULL),

-- Benjamin Masculin (10-11 ans)
('Benjamin', 'man', 10, 11, '-30',  30),
('Benjamin', 'man', 10, 11, '-34',  34),
('Benjamin', 'man', 10, 11, '-38',  38),
('Benjamin', 'man', 10, 11, '-42',  42),
('Benjamin', 'man', 10, 11, '-46',  46),
('Benjamin', 'man', 10, 11, '-50',  50),
('Benjamin', 'man', 10, 11, '-55',  55),
('Benjamin', 'man', 10, 11, '-60',  60),
('Benjamin', 'man', 10, 11, '+60',  NULL),

-- Benjamin Féminin (10-11 ans)
('Benjamin', 'woman', 10, 11, '-28',  28),
('Benjamin', 'woman', 10, 11, '-32',  32),
('Benjamin', 'woman', 10, 11, '-36',  36),
('Benjamin', 'woman', 10, 11, '-40',  40),
('Benjamin', 'woman', 10, 11, '-44',  44),
('Benjamin', 'woman', 10, 11, '-48',  48),
('Benjamin', 'woman', 10, 11, '-52',  52),
('Benjamin', 'woman', 10, 11, '+52',  NULL),

-- Minime Masculin (12-13 ans)
('Minime', 'man', 12, 13, '-34',  34),
('Minime', 'man', 12, 13, '-38',  38),
('Minime', 'man', 12, 13, '-42',  42),
('Minime', 'man', 12, 13, '-46',  46),
('Minime', 'man', 12, 13, '-50',  50),
('Minime', 'man', 12, 13, '-55',  55),
('Minime', 'man', 12, 13, '-60',  60),
('Minime', 'man', 12, 13, '-66',  66),
('Minime', 'man', 12, 13, '+66',  NULL),

-- Minime Féminin (12-13 ans)
('Minime', 'woman', 12, 13, '-32',  32),
('Minime', 'woman', 12, 13, '-36',  36),
('Minime', 'woman', 12, 13, '-40',  40),
('Minime', 'woman', 12, 13, '-44',  44),
('Minime', 'woman', 12, 13, '-48',  48),
('Minime', 'woman', 12, 13, '-52',  52),
('Minime', 'woman', 12, 13, '-57',  57),
('Minime', 'woman', 12, 13, '+57',  NULL),

-- Cadet Masculin (14-15 ans)
('Cadet', 'man', 14, 15, '-46',  46),
('Cadet', 'man', 14, 15, '-50',  50),
('Cadet', 'man', 14, 15, '-55',  55),
('Cadet', 'man', 14, 15, '-60',  60),
('Cadet', 'man', 14, 15, '-66',  66),
('Cadet', 'man', 14, 15, '-73',  73),
('Cadet', 'man', 14, 15, '-81',  81),
('Cadet', 'man', 14, 15, '+81',  NULL),

-- Cadet Féminin (14-15 ans)
('Cadet', 'woman', 14, 15, '-40',  40),
('Cadet', 'woman', 14, 15, '-44',  44),
('Cadet', 'woman', 14, 15, '-48',  48),
('Cadet', 'woman', 14, 15, '-52',  52),
('Cadet', 'woman', 14, 15, '-57',  57),
('Cadet', 'woman', 14, 15, '-63',  63),
('Cadet', 'woman', 14, 15, '-70',  70),
('Cadet', 'woman', 14, 15, '+70',  NULL),

-- Junior Masculin (16-17 ans)
('Junior', 'man', 16, 17, '-55',  55),
('Junior', 'man', 16, 17, '-60',  60),
('Junior', 'man', 16, 17, '-66',  66),
('Junior', 'man', 16, 17, '-73',  73),
('Junior', 'man', 16, 17, '-81',  81),
('Junior', 'man', 16, 17, '-90',  90),
('Junior', 'man', 16, 17, '-100', 100),
('Junior', 'man', 16, 17, '+100', NULL),

-- Junior Féminin (16-17 ans)
('Junior', 'woman', 16, 17, '-44',  44),
('Junior', 'woman', 16, 17, '-48',  48),
('Junior', 'woman', 16, 17, '-52',  52),
('Junior', 'woman', 16, 17, '-57',  57),
('Junior', 'woman', 16, 17, '-63',  63),
('Junior', 'woman', 16, 17, '-70',  70),
('Junior', 'woman', 16, 17, '-78',  78),
('Junior', 'woman', 16, 17, '+78',  NULL),

-- Senior Masculin (18-34 ans)
('Senior', 'man', 18, 34, '-60',  60),
('Senior', 'man', 18, 34, '-66',  66),
('Senior', 'man', 18, 34, '-73',  73),
('Senior', 'man', 18, 34, '-81',  81),
('Senior', 'man', 18, 34, '-90',  90),
('Senior', 'man', 18, 34, '-100', 100),
('Senior', 'man', 18, 34, '+100', NULL),

-- Senior Féminin (18-34 ans)
('Senior', 'woman', 18, 34, '-48',  48),
('Senior', 'woman', 18, 34, '-52',  52),
('Senior', 'woman', 18, 34, '-57',  57),
('Senior', 'woman', 18, 34, '-63',  63),
('Senior', 'woman', 18, 34, '-70',  70),
('Senior', 'woman', 18, 34, '-78',  78),
('Senior', 'woman', 18, 34, '+78',  NULL),

-- Veteran Masculin (35+ ans)
('Veteran', 'man', 35, NULL, '-60',  60),
('Veteran', 'man', 35, NULL, '-66',  66),
('Veteran', 'man', 35, NULL, '-73',  73),
('Veteran', 'man', 35, NULL, '-81',  81),
('Veteran', 'man', 35, NULL, '-90',  90),
('Veteran', 'man', 35, NULL, '-100', 100),
('Veteran', 'man', 35, NULL, '+100', NULL),

-- Veteran Féminin (35+ ans)
('Veteran', 'woman', 35, NULL, '-48',  48),
('Veteran', 'woman', 35, NULL, '-52',  52),
('Veteran', 'woman', 35, NULL, '-57',  57),
('Veteran', 'woman', 35, NULL, '-63',  63),
('Veteran', 'woman', 35, NULL, '-70',  70),
('Veteran', 'woman', 35, NULL, '-78',  78),
('Veteran', 'woman', 35, NULL, '+78',  NULL)

ON CONFLICT (age_group, gender, weight_label) DO NOTHING;

-- ============================================================
-- Compétition test : Compétition interne A3M (dimanche prochain)
-- Créée par Redwane, ouverte, toutes catégories Senior (H+F)
-- Pesée à 08h30, début le 2026-06-14
-- ============================================================
INSERT INTO events (club_id, created_by, title, description, type, status, location,
                    registration_open_at, registration_close_at, date, max_participants)
SELECT
  c.id, u.id,
  'Compétition interne A3M',
  'Compétition ouverte à tous les membres du club.',
  'competition', 'open',
  'Dojo A3M — Mitry-Mory',
  '2026-06-11 00:00:00+00',
  '2026-06-12 23:59:00+00',
  '2026-06-14 09:00:00+00',
  NULL
FROM clubs c, users u
WHERE c.siren = '382048650' AND u.email = 'user@clubmanager.fr'
ON CONFLICT DO NOTHING;

-- Post d'annonce lié à la compétition
INSERT INTO posts (club_id, author_id, event_id, title, content, status, visibility)
SELECT
  c.id,
  u.id,
  e.id,
  'Compétition interne A3M — Inscriptions ouvertes',
  'Les inscriptions sont maintenant ouvertes !' || E'\n\n' ||
  'Date : 14 juin 2026' || E'\n' ||
  'Lieu : Dojo A3M — Mitry-Mory' || E'\n\n' ||
  'Compétition ouverte à tous les membres du club.',
  'published',
  'adherent'
FROM events e
JOIN clubs c ON c.id = e.club_id
JOIN users u ON u.email = 'user@clubmanager.fr'
WHERE c.siren = '382048650' AND e.title = 'Compétition interne A3M'
ON CONFLICT DO NOTHING;

-- Toutes les catégories Senior (homme et femme) avec pesée à 08h30
INSERT INTO event_categories (event_id, judo_category_id, weigh_in_at, status)
SELECT e.id, jc.id, '2026-06-14 08:30:00+00', 'pending'
FROM events e
JOIN clubs c ON c.id = e.club_id
CROSS JOIN judo_categories jc
WHERE c.siren = '382048650'
  AND e.title = 'Compétition interne A3M'
  AND jc.age_group = 'Senior'
ON CONFLICT (event_id, judo_category_id) DO NOTHING;
