CREATE TABLE IF NOT EXISTS code_sequences (
  name    TEXT PRIMARY KEY,
  last_no INTEGER NOT NULL
);
INSERT OR IGNORE INTO code_sequences(name,last_no) VALUES ('MA2Y',0);




 CREATE TABLE IF NOT EXISTS ma_master (
  MA000    TEXT    PRIMARY KEY,
  MA009    TEXT    UNIQUE NOT NULL DEFAULT '',   -- 発番済 YJ コード
  MA018    TEXT    NOT NULL DEFAULT '',          -- 品名
  MA022    TEXT    NOT NULL DEFAULT '',          -- 品名かな
  MA030    TEXT    NOT NULL DEFAULT '',          -- メーカー
  MA037    TEXT    NOT NULL DEFAULT '',          -- 包装
  MA039    TEXT    NOT NULL DEFAULT '',          -- YJ側単位コード
  MA044    TEXT    NOT NULL DEFAULT '',          -- YJ側数量文字列
  MA061    INTEGER NOT NULL DEFAULT 0,           -- 毒薬
  MA062    INTEGER NOT NULL DEFAULT 0,           -- 劇薬
  MA063    INTEGER NOT NULL DEFAULT 0,           -- 麻薬
  MA064    INTEGER NOT NULL DEFAULT 0,           -- 向精神薬
  MA065    INTEGER NOT NULL DEFAULT 0,           -- 覚せい剤
  MA066    INTEGER NOT NULL DEFAULT 0,           -- 覚醒剤原料
  MA131    TEXT    NOT NULL DEFAULT '',          -- JAN単位名
  MA132    TEXT    NOT NULL DEFAULT '',          -- JAN単位コード
  MA133    TEXT    NOT NULL DEFAULT ''           -- JANあたり数量
 );





CREATE INDEX IF NOT EXISTS idxMaMasterMA009
  ON ma_master(MA009);



CREATE INDEX IF NOT EXISTS idx_master_MA009
  ON ma_master(MA009);


CREATE TABLE IF NOT EXISTS a_records (
  adate                      TEXT    NOT NULL,
  apcode                     TEXT    NOT NULL,
  arpnum                     TEXT    NOT NULL,
  alnum                      TEXT    NOT NULL,
  aflag                      INTEGER NOT NULL,
  ajc                        TEXT    NOT NULL,
  ayj                        TEXT,
  apname                     TEXT,
  akana                      TEXT,
  apkg                       TEXT,
  amaker                     TEXT,
  adatqty                    REAL,
  ajanqty                    REAL,
  ajpu                       REAL,
  ajanunitname               TEXT,
  ajanunitcode               TEXT,
  ayjqty                     REAL,
  ayjpu                      TEXT,
  ayjunitname                TEXT,
  aunitprice                 REAL,
  asubtotal                  REAL,
  ataxamount                 REAL,
  ataxrate                   TEXT,
  aexpdate                   TEXT,
  alot                       TEXT,
  adokuyaku                  INTEGER DEFAULT 0,
  agekiyaku                  INTEGER DEFAULT 0,
  amayaku                    INTEGER DEFAULT 0,
  akouseisinyaku             INTEGER DEFAULT 0,
  akakuseizai                INTEGER DEFAULT 0,
  akakuseizaigenryou         INTEGER DEFAULT 0,
  ama                        TEXT,
  PRIMARY KEY (apcode, adate, arpnum, alnum, aflag)
);

-- ソート／フィルタ用インデックス
CREATE INDEX IF NOT EXISTS idx_ar_apname_kana
  ON a_records(akana);
CREATE INDEX IF NOT EXISTS idx_ar_adokuyaku
  ON a_records(adokuyaku);
-- 必要に応じて他フラグにも INDEX を貼る

CREATE TABLE IF NOT EXISTS jcshms (
JC000 TEXT,
JC001 TEXT,
JC002 TEXT,
JC003 TEXT,
JC004 TEXT,
JC005 TEXT,
JC006 TEXT,
JC007 TEXT,
JC008 TEXT,
JC009 TEXT,
JC010 TEXT,
JC011 TEXT,
JC012 TEXT,
JC013 TEXT,
JC014 TEXT,
JC015 TEXT,
JC016 TEXT,
JC017 TEXT,
JC018 TEXT,
JC019 TEXT,
JC020 TEXT,
JC021 TEXT,
JC022 TEXT,
JC023 TEXT,
JC024 TEXT,
JC025 TEXT,
JC026 TEXT,
JC027 TEXT,
JC028 TEXT,
JC029 TEXT,
JC030 TEXT,
JC031 TEXT,
JC032 TEXT,
JC033 TEXT,
JC034 TEXT,
JC035 TEXT,
JC036 TEXT,
JC037 TEXT,
JC038 TEXT,
JC039 TEXT,
JC040 TEXT,
JC041 TEXT,
JC042 TEXT,
JC043 TEXT,
JC044 TEXT,
JC045 TEXT,
JC046 TEXT,
JC047 TEXT,
JC048 TEXT,
JC049 TEXT,
JC050 TEXT,
JC051 TEXT,
JC052 TEXT,
JC053 TEXT,
JC054 TEXT,
JC055 TEXT,
JC056 TEXT,
JC057 TEXT,
JC058 TEXT,
JC059 TEXT,
JC060 TEXT,
JC061 TEXT,
JC062 TEXT,
JC063 TEXT,
JC064 TEXT,
JC065 TEXT,
JC066 TEXT,
JC067 TEXT,
JC068 TEXT,
JC069 TEXT,
JC070 TEXT,
JC071 TEXT,
JC072 TEXT,
JC073 TEXT,
JC074 TEXT,
JC075 TEXT,
JC076 TEXT,
JC077 TEXT,
JC078 TEXT,
JC079 TEXT,
JC080 TEXT,
JC081 TEXT,
JC082 TEXT,
JC083 TEXT,
JC084 TEXT,
JC085 TEXT,
JC086 TEXT,
JC087 TEXT,
JC088 TEXT,
JC089 TEXT,
JC090 TEXT,
JC091 TEXT,
JC092 TEXT,
JC093 TEXT,
JC094 TEXT,
JC095 TEXT,
JC096 TEXT,
JC097 TEXT,
JC098 TEXT,
JC099 TEXT,
JC100 TEXT,
JC101 TEXT,
JC102 TEXT,
JC103 TEXT,
JC104 TEXT,
JC105 TEXT,
JC106 TEXT,
JC107 TEXT,
JC108 TEXT,
JC109 TEXT,
JC110 TEXT,
JC111 TEXT,
JC112 TEXT,
JC113 TEXT,
JC114 TEXT,
JC115 TEXT,
JC116 TEXT,
JC117 TEXT,
JC118 TEXT,
JC119 TEXT,
JC120 TEXT,
JC121 TEXT,
JC122 TEXT,
JC123 TEXT,
JC124 TEXT,
PRIMARY KEY(JC000)
);

CREATE TABLE IF NOT EXISTS jancode (
JA000 TEXT,
JA001 TEXT,
JA002 TEXT,
JA003 TEXT,
JA004 TEXT,
JA005 TEXT,
JA006 TEXT,
JA007 TEXT,
JA008 TEXT,
JA009 TEXT,
JA010 TEXT,
JA011 TEXT,
JA012 TEXT,
JA013 TEXT,
JA014 TEXT,
JA015 TEXT,
JA016 TEXT,
JA017 TEXT,
JA018 TEXT,
JA019 TEXT,
JA020 TEXT,
JA021 TEXT,
JA022 TEXT,
JA023 TEXT,
JA024 TEXT,
JA025 TEXT,
JA026 TEXT,
JA027 TEXT,
JA028 TEXT,
JA029 TEXT,
PRIMARY KEY(JA001)
);


-- File: Sql/Schema/CreatePartnerMaster.sql

-- パートナーコード → 社名マスター
CREATE TABLE IF NOT EXISTS partner_master (
  id   INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT    NOT NULL UNIQUE,    -- パートナーコード
  name TEXT    NOT NULL            -- 社名
);

-- 初期４社登録
INSERT OR IGNORE INTO partner_master (code, name) VALUES
  ('902020014', 'スズケン'),
  ('901660013', 'メディセオ'),
  ('902690019', '中北薬品'),
  ('902960013', 'アルフレッサ');

