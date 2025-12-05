-- +goose Up
CREATE TABLE sections (
	code TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	manager_id UUID REFERENCES users (id)
)
;

INSERT INTO
	sections (code, name)
VALUES
	('101', 'Administrerende direktør'),
	('102', 'Stab Administrasjonsavdeling'),
	(
		'111',
		'Seksjon for økonomi og virksomhetsstyring'
	),
	('120', 'Internasjonalt sekretæriat'),
	('150', 'Seksjon HR'),
	(
		'160',
		'Seksjon for eiendom, arkiv og administrative systemer'
	),
	('201', 'Stab økonomisk statistikk'),
	('210', 'Seksjon for nasjonalregnskap'),
	('211', 'Seksjon for finansregnskap'),
	('212', 'Seksjon for offentlige finanser'),
	('213', 'Seksjon for finansmarkedsstatistikk'),
	('214', 'Seksjon for utenrikshandelsstatistikk'),
	(
		'216',
		'Seksjon for internasjonalt utviklingsarbeid'
	),
	('240', 'Seksjon for prisstatistikk'),
	('301', 'Stab person- og sosialstatistikk'),
	(
		'312',
		'Seksjon for arbeidsmarkeds- og lønnsstatistikk'
	),
	('320', 'Seksjon for befolkningsstatistikk'),
	('330', 'Seksjon for helsestatistikk'),
	(
		'350',
		'Seksjon for inntekts- og levekårsstatistikk'
	),
	('360', 'Seksjon for utdanningsstatistikk'),
	('380', 'Seksjon for mikrodata'),
	('401', 'Stab - nærings- og miljøstatistikk'),
	(
		'421',
		'Seksjon for FoU, teknologi og næringslivets utvikling.'
	),
	('422', 'Seksjon for næringslivets konjunkturer'),
	('423', 'Seksjon for næringslivets strukturer'),
	('424', 'Seksjon for regnskapsstatistikk og BoF'),
	(
		'425',
		'Seksjon for energi-, miljø- og transportstatistikk'
	),
	(
		'426',
		'Seksjon for eiendoms-, areal- og primærnæringsstatistikk'
	),
	('501', 'Stab Forskningsavdelingen'),
	(
		'510',
		'Gruppe for befolkning og offentlig økonomi '
	),
	(
		'520',
		'Gruppe for miljø-, ressurs- og innovasjonsøkonomi'
	),
	('530', 'Gruppe for makroøkonomi'),
	('550', 'Gruppe for arbeidsmarked og skatt'),
	('601', 'Stab Kommunikasjon og brukerkontakt'),
	('610', 'Seksjon for redaksjon og publisering'),
	('611', 'Seksjon for brukerkontakt'),
	('630', 'Seksjon for virksomhetskommunikasjon'),
	(
		'660',
		'Seksjon for brukerinnsikt og webutvikling'
	),
	('701', 'Stab IT'),
	('702', 'Seksjon for IT-arkitektur'),
	('703', 'Seksjon for IT-partner'),
	('722', 'Seksjon for datafangstplattform'),
	('723', 'Seksjon for formidlingsplattform'),
	('724', 'Seksjon for dataplattform'),
	('782', 'Seksjon for drift og infrastruktur'),
	('801', 'Stab metodeutvikling og datainnsamling'),
	('811', 'Seksjon for metode'),
	('821', 'Seksjon for næringslivsundersøkelser'),
	(
		'831',
		'Seksjon for operasjonell forretningsstøtte'
	),
	('851', 'Seksjon for personundersøkelser')
;

ALTER TABLE teams
ADD COLUMN section_code TEXT NOT NULL DEFAULT '724' REFERENCES sections (code)
;

ALTER TABLE teams
ALTER COLUMN section_code
DROP DEFAULT
;
