-- Seed data : catégories classiques FFJ
-- age_max NULL = pas de limite supérieure (Veteran)
-- weight_max NULL = catégorie ouverte (+)

INSERT INTO judo_categories (age_group, gender, age_min, age_max, weight_label, weight_max) VALUES
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
