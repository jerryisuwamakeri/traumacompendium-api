-- Draft translations (needs_review = TRUE throughout, per migration
-- 0015-0019's column default) of the content seeded by migration 0014,
-- into the 7 non-English locales from learning.SupportedLocales:
-- pcm (Nigerian Pidgin), yo (Yoruba), ha (Hausa), ig (Igbo), fr (French),
-- sw (Swahili), ar (Arabic).
--
-- These are LLM-drafted, NOT reviewed by a native-speaking professional
-- translator or a clinician for this trauma-sensitive subject matter.
-- Every row here is marked needs_review = TRUE and the README documents
-- this explicitly. Do not present this copy to end users as finished.

-- ============================================================
-- Levels (6 levels x 7 locales = 42 rows)
-- ============================================================
INSERT INTO level_translations (level_id, locale, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'pcm', 'Beginner', 'Wetin trauma be, lies people dey tell, how brain dey work'),
    ('00000000-0000-0000-0000-000000000001', 'yo', 'Ibẹrẹ', 'Ìpilẹ̀ ìpalára-ọkàn, àwọn irọ́ tó ń jẹ́ òtítọ́, ìmọ̀ nípa ọpọlọ'),
    ('00000000-0000-0000-0000-000000000001', 'ha', 'Mafara', 'Asalin rauni, ƙarairayi, ilimin kwakwalwa'),
    ('00000000-0000-0000-0000-000000000001', 'ig', 'Mmalite', 'Ihe bụ mmerụ ahụ-mmụọ, okwu ụgha, sayensị ụbụrụ'),
    ('00000000-0000-0000-0000-000000000001', 'fr', 'Débutant', 'Bases du traumatisme, mythes, neurosciences'),
    ('00000000-0000-0000-0000-000000000001', 'sw', 'Mwanzo', 'Misingi ya majeraha ya kiakili, hadithi za uwongo, sayansi ya ubongo'),
    ('00000000-0000-0000-0000-000000000001', 'ar', 'مبتدئ', 'أساسيات الصدمة، الخرافات، علم الدماغ'),

    ('00000000-0000-0000-0000-000000000002', 'pcm', 'Know Yourself', 'Fight/flight/freeze/fawn, wetin dey trigger you'),
    ('00000000-0000-0000-0000-000000000002', 'yo', 'Ìmọ̀ Ara-ẹni', 'Ìjà/Ìsá/Dídi-dúró/Ìtẹríba, àwọn ohun tó ń mú kí ẹ̀dùn-ọkàn jáde'),
    ('00000000-0000-0000-0000-000000000002', 'ha', 'Sanin Kai', 'Faɗa/Gudu/Tsayawa/Sulhu, abubuwan da ke jawo martani'),
    ('00000000-0000-0000-0000-000000000002', 'ig', 'Ịmara Onwe Gị', 'Ịlụ ọgụ/Ịgba ọsọ/Ịkwụsị/Ịdọrọ, ihe ndị na-akpali mmetụta'),
    ('00000000-0000-0000-0000-000000000002', 'fr', 'Conscience de soi', 'Combat/fuite/figement/soumission, déclencheurs'),
    ('00000000-0000-0000-0000-000000000002', 'sw', 'Kujitambua', 'Kupigana/Kukimbia/Kuganda/Kuridhia, vichocheo'),
    ('00000000-0000-0000-0000-000000000002', 'ar', 'الوعي الذاتي', 'القتال/الهروب/التجمد/الاسترضاء، المحفزات'),

    ('00000000-0000-0000-0000-000000000003', 'pcm', 'Calm Yourself', 'How to ground yourself, sleep well, manage stress'),
    ('00000000-0000-0000-0000-000000000003', 'yo', 'Ìṣàkóso Ara-ẹni', 'Ìdúró gbọingboin, oorun, ìṣàkóso másùnmáwo'),
    ('00000000-0000-0000-0000-000000000003', 'ha', 'Sarrafa Kai', 'Tabbatar da kai, barci, sarrafa damuwa'),
    ('00000000-0000-0000-0000-000000000003', 'ig', 'Njikwa Onwe Gị', 'Ịdọkwasị onwe gị, ụra, njikwa nrụgide'),
    ('00000000-0000-0000-0000-000000000003', 'fr', 'Régulation', 'Ancrage, sommeil, gestion du stress'),
    ('00000000-0000-0000-0000-000000000003', 'sw', 'Udhibiti', 'Kujiweka imara, usingizi, usimamizi wa msongo'),
    ('00000000-0000-0000-0000-000000000003', 'ar', 'التنظيم', 'التأريض، النوم، إدارة التوتر'),

    ('00000000-0000-0000-0000-000000000004', 'pcm', 'Relationship', 'Boundary, trust, how you dey attach to people'),
    ('00000000-0000-0000-0000-000000000004', 'yo', 'Àjọṣepọ̀', 'Àlà, ìgbẹ́kẹ̀lé, onírúurú ọ̀nà ìfaramọ́'),
    ('00000000-0000-0000-0000-000000000004', 'ha', 'Dangantaka', 'Iyaka, amana, nau''ikan haɗin kai'),
    ('00000000-0000-0000-0000-000000000004', 'ig', 'Mmekọrịta', 'Oke, ntụkwasị obi, ụdị njikọ'),
    ('00000000-0000-0000-0000-000000000004', 'fr', 'Relations', 'Limites, confiance, styles d''attachement'),
    ('00000000-0000-0000-0000-000000000004', 'sw', 'Mahusiano', 'Mipaka, imani, mitindo ya kushikamana'),
    ('00000000-0000-0000-0000-000000000004', 'ar', 'العلاقات', 'الحدود، الثقة، أنماط التعلق'),

    ('00000000-0000-0000-0000-000000000005', 'pcm', 'Growth', 'Strong mind, values, how to set goal'),
    ('00000000-0000-0000-0000-000000000005', 'yo', 'Ìdàgbàsókè', 'Agbára láti dúró, àwọn ìlànà, ìgbékalẹ̀ èrèdí'),
    ('00000000-0000-0000-0000-000000000005', 'ha', 'Ci Gaba', 'Juriya, ƙimomi, tsara manufa'),
    ('00000000-0000-0000-0000-000000000005', 'ig', 'Uto', 'Ntachi obi, ụkpụrụ, ịhazi ebumnuche'),
    ('00000000-0000-0000-0000-000000000005', 'fr', 'Croissance', 'Résilience, valeurs, fixation d''objectifs'),
    ('00000000-0000-0000-0000-000000000005', 'sw', 'Ukuaji', 'Uthabiti, maadili, kuweka malengo'),
    ('00000000-0000-0000-0000-000000000005', 'ar', 'النمو', 'المرونة، القيم، تحديد الأهداف'),

    ('00000000-0000-0000-0000-000000000006', 'pcm', 'Advanced', 'Trauma-informed leadership and how to support your padi dem'),
    ('00000000-0000-0000-0000-000000000006', 'yo', 'Onípele Gíga', 'Ìdarí tó mọ̀ nípa ìpalára-ọkàn àti ìrànlọ́wọ́ láàrin ẹgbẹ́'),
    ('00000000-0000-0000-0000-000000000006', 'ha', 'Ci-gaba', 'Jagoranci da sanin rauni da tallafin tsara-tsara'),
    ('00000000-0000-0000-0000-000000000006', 'ig', 'Ọkwa Elu', 'Nduzi gbasara mmerụ ahụ-mmụọ na nkwado ndị ibe'),
    ('00000000-0000-0000-0000-000000000006', 'fr', 'Avancé', 'Leadership tenant compte des traumatismes et soutien par les pairs'),
    ('00000000-0000-0000-0000-000000000006', 'sw', 'Juu Zaidi', 'Uongozi unaozingatia majeraha ya kiakili na msaada wa rika'),
    ('00000000-0000-0000-0000-000000000006', 'ar', 'متقدم', 'القيادة الواعية بالصدمة ودعم الأقران');

-- ============================================================
-- Lessons (3 lessons x 7 locales = 21 rows)
-- ============================================================
INSERT INTO lesson_translations (lesson_id, locale, title, content) VALUES
    (
        '00000000-0000-0000-0000-000000000101', 'pcm',
        'Wetin Trauma Be?',
        'Trauma na how your body and mind dey react to something (or many things) wey too much for you to handle for that time. E no be about how bad the thing look from outside -- na how your body and mind take feel am.\n\nTrauma fit come from one event (accident, attack, loss) or from things wey happen plenty times (stress wey no dey stop, neglect, unsafe place). Two people fit pass through the same thing but their reaction go different well well, because trauma dey shaped by the event itself, your history, the people wey dey support you, and how your body system set up.\n\nThis platform no dey diagnose or treat trauma. Make you see these lessons as starting point make you understand your own experience -- foundation wey you fit build on with correct professional if you decide to.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'yo',
        'Kín Ni Ìpalára-Ọkàn?',
        'Ìpalára-ọkàn jẹ́ ìdáhùn ètò ara rẹ sí ìṣẹ̀lẹ̀ kan (tàbí ọ̀pọ̀lọpọ̀ ìṣẹ̀lẹ̀) tó pọ̀ jù fún ọ láti bójútó nígbà yẹn. Kì í ṣe nípa bí ìṣẹ̀lẹ̀ náà ṣe burú tó láti ojú ìwòye àwọn ẹlòmíràn -- bí kó ṣe jẹ́ pé ara àti ọkàn rẹ ló nírìírí rẹ̀.\n\nÌpalára-ọkàn lè wá láti inú ìṣẹ̀lẹ̀ kan (ìjàmbá, ìkọlù, ìpàdánù) tàbí láti inú àwọn ìrírí tó ń ṣẹlẹ̀ léraléra (másùnmáwo tó ń bá a nìṣó, àìkọbiara, àyíká tí kò ní ààbò). Ènìyàn méjì lè kọjá nínú ìṣẹ̀lẹ̀ kan náà kí wọ́n sì ní ìdáhùn tó yàtọ̀ síra, nítorí pé ìpalára-ọkàn ń gbé àpẹẹrẹ ara ìṣẹ̀lẹ̀ náà, ìtàn ẹ̀mí rẹ, àwọn tó ń ṣètìlẹ́yìn fún ọ, àti ìdásílẹ̀ ètò ara rẹ.\n\nPèpéle yìí kì í ṣe àyẹ̀wò tàbí ìtọ́jú fún ìpalára-ọkàn. Wo àwọn ẹ̀kọ́ wọ̀nyí gẹ́gẹ́ bí ìbẹ̀rẹ̀ ọ̀rọ̀ láti lóye ìrírí tìrẹ -- ìpìlẹ̀ tí o lè kọ́ lé lórí pẹ̀lú akọ́ṣẹ́mọṣẹ́ tó yẹ bí o bá yàn láti ṣe bẹ́ẹ̀.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'ha',
        'Menene Rauni?',
        'Rauni shi ne martanin tsarin jikinka ga wani abu (ko jerin abubuwa) da ya fi ƙarfin jurewarka a lokacin. Ba game da yadda al''amarin ya yi muni a idon waje ba ne -- game da yadda jikinka da hankalinka suka fuskance shi ne.\n\nRauni na iya fitowa daga abu guda (hatsari, farmaki, asara) ko daga abubuwan da ke maimaituwa (damuwa mai ɗorewa, sakaci, mahalli maras tsaro). Mutane biyu na iya fuskantar abu ɗaya amma su sami martani daban-daban, domin rauni yana samuwa ta hanyar abin da ya faru, tarihinka, waɗanda ke tallafa maka, da tsarin jikinka.\n\nWannan dandali ba ya gano ko magance rauni. Ka ɗauki waɗannan darussa a matsayin farkon fahimtar abin da ka fuskanta -- tushen da za ka iya ginawa tare da ƙwararre idan ka zaɓi yin haka.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'ig',
        'Gịnị Bụ Mmerụ Ahụ-Mmụọ?',
        'Mmerụ ahụ-mmụọ bụ mmeghachi omume nke sistemu akwara ozi gị na ihe mere (ma ọ bụ usoro ihe mere) kariri ike gị ịnagide n''oge ahụ. Ọ bụghị gbasara ka ihe ahụ si dị egwu n''anya ndị ọzọ -- kama ọ bụ gbasara ka ahụ na uche gị si nwe mmetụta ya.\n\nMmerụ ahụ-mmụọ pụrụ isi na otu ihe (ihe mberede, mwakpo, mfu) ma ọ bụ site n''ahụmahụ na-eme ọtụtụ ugboro (nrụgide na-adịgide, elegara anya, gburugburu na-adịghị nchebe). Mmadụ abụọ nwere ike ịgafe otu ihe ma nwee mmeghachi omume dị iche iche, n''ihi na mmerụ ahụ-mmụọ na-adaberekwa n''ihe mere, akụkọ ndụ gị, ndị na-akwado gị, na otu sistemu akwara ozi gị si dị.\n\nIkpo okwu a adịghị achọpụta ma ọ bụ gwọọ mmerụ ahụ-mmụọ. Were nkuzi ndị a dịka mmalite ịghọta ahụmahụ gị -- ntọala ị nwere ike iwukwasị ya na enyemaka onye ọkachamara ma ọ bụrụ na ị họrọ ime otú ahụ.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'fr',
        'Qu''est-ce que le traumatisme ?',
        'Le traumatisme est la réaction de votre système nerveux à un événement (ou une série d''événements) qui a dépassé votre capacité à y faire face à ce moment-là. Ce n''est pas une question de gravité apparente de l''événement vu de l''extérieur -- c''est une question de la façon dont votre corps et votre esprit l''ont vécu.\n\nLe traumatisme peut provenir d''un événement unique (accident, agression, perte) ou d''expériences répétées dans le temps (stress continu, négligence, environnement dangereux). Deux personnes peuvent vivre le même événement et avoir des réactions très différentes, car le traumatisme est façonné par l''événement lui-même, votre histoire, votre système de soutien et le fonctionnement de votre système nerveux.\n\nCette plateforme ne diagnostique ni ne traite le traumatisme. Considérez ces leçons comme un vocabulaire de départ pour comprendre votre propre expérience -- une base sur laquelle vous pourrez vous appuyer avec un professionnel qualifié si et quand vous le choisirez.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'sw',
        'Jeraha la Kiakili Ni Nini?',
        'Jeraha la kiakili ni mwitikio wa mfumo wako wa neva kwa tukio (au mfululizo wa matukio) lililozidi uwezo wako wa kukabiliana nalo wakati huo. Si kuhusu jinsi tukio linavyoonekana kuwa kubwa kwa nje -- ni kuhusu jinsi mwili na akili yako vilivyolipitia.\n\nJeraha la kiakili laweza kutokana na tukio moja (ajali, shambulio, hasara) au kutokana na matukio yanayojirudia (msongo wa mara kwa mara, kutojaliwa, mazingira yasiyo salama). Watu wawili wanaweza kupitia tukio lile lile na kuwa na miitikio tofauti sana, kwa sababu jeraha la kiakili huundwa na tukio lenyewe, historia yako, mfumo wako wa msaada, na jinsi mfumo wako wa neva ulivyo.\n\nJukwaa hili halitambui wala kutibu jeraha la kiakili. Chukulia masomo haya kama msamiati wa kuanzia kuelewa uzoefu wako mwenyewe -- msingi unaoweza kuujengea na mtaalamu aliyehitimu endapo utachagua kufanya hivyo.'
    ),
    (
        '00000000-0000-0000-0000-000000000101', 'ar',
        'ما هي الصدمة؟',
        'الصدمة هي استجابة جهازك العصبي لحدث (أو سلسلة من الأحداث) تجاوز قدرتك على التأقلم في ذلك الوقت. الأمر لا يتعلق بمدى شدة الحدث من الخارج -- بل بكيفية تجربة جسدك وعقلك له.\n\nيمكن أن تنشأ الصدمة من حدث واحد (حادث، اعتداء، فقدان) أو من تجارب متكررة بمرور الوقت (ضغط مستمر، إهمال، بيئة غير آمنة). يمكن لشخصين أن يمرا بنفس الحدث ويكون لديهما استجابات مختلفة تمامًا، لأن الصدمة تتشكل بفعل الحدث نفسه، وتاريخك، ونظام دعمك، وتوصيل جهازك العصبي.\n\nهذه المنصة لا تشخص الصدمة ولا تعالجها. اعتبر هذه الدروس مفردات أولية لفهم تجربتك الخاصة -- أساس يمكنك البناء عليه مع متخصص مؤهل إذا وعندما تختار ذلك.'
    ),

    (
        '00000000-0000-0000-0000-000000000102', 'pcm',
        'Lies Wey People Dey Believe About Trauma',
        'Lie: "Na only big big wahala (war, disaster) dey cause trauma."\nReal tin: Trauma fit come from plenty kain experience, even the ones wey look "normal" from outside, like stress wey no dey stop, bullying, or house wey no settled.\n\nLie: "If you no remember am well well, e no bad reach."\nReal tin: Memory wey pass through serious stress dey work different -- broken or blurry memory na normal tin and e no mean say the wahala no serious.\n\nLie: "Strong people no dey affected by trauma."\nReal tin: Trauma reaction na normal thing wey your body system dey do, e no be say person weak or get bad character.\n\nLie: "Time alone go heal trauma."\nReal tin: Time fit help, but understanding, support, and sometimes professional care na wetin dey really carry person from just surviving go healing.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'yo',
        'Àwọn Irọ́ Tó Wọ́pọ̀ Nípa Ìpalára-Ọkàn',
        'Irọ́: "Àwọn ìṣẹ̀lẹ̀ ńlá nìkan (ogun, àjálù) ni ó ń fa ìpalára-ọkàn."\nÒtítọ́: Ìpalára-ọkàn lè wá láti onírúurú ìrírí, títí kan àwọn tó dà bí "lásán" láti ìta, bí másùnmáwo tí kò dáwọ́ dúró, ìpanilára, tàbí ilé tí kò fẹsẹ̀múlẹ̀.\n\nIrọ́: "Bí o kò bá rántí rẹ̀ kedere, kò burú tó bẹ́ẹ̀."\nÒtítọ́: Ìrántí lábẹ́ másùnmáwo ńlá máa ń ṣiṣẹ́ yàtọ̀ -- ìrántí tí ó yapa tàbí tí kò yé ni ó wọ́pọ̀, kò sì túmọ̀ sí pé ìrírí náà kéré.\n\nIrọ́: "Ìpalára-ọkàn kì í kan àwọn tó lágbára."\nÒtítọ́: Ìdáhùn ìpalára-ọkàn jẹ́ ìhùwàsí ètò ara lásán, kì í ṣe àbùkù ìwà tàbí àmì àìlera.\n\nIrọ́: "Àkókò nìkan ló ń wo ìpalára-ọkàn sàn."\nÒtítọ́: Àkókò lè ṣèrànwọ́, ṣùgbọ́n òye, ìtìlẹ́yìn, àti nígbà míràn ìtọ́jú akọ́ṣẹ́mọṣẹ́ ni ó sábà máa ń mú ẹnì kan láti ipò ìwàláàyè dé ipò ìwòsàn.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'ha',
        'Ƙarairayi Na Kowa Game Da Rauni',
        'Ƙarya: "Manyan bala''o''i kaɗai (yaƙi, bala''i) ke haifar da rauni."\nGaskiya: Rauni na iya fitowa daga nau''ikan abubuwa daban-daban, har da waɗanda suke kama da "na yau da kullum" a waje, kamar damuwa mai ɗorewa, cin zarafi, ko gida mara tabbas.\n\nƘarya: "Idan ba ka tuna shi sosai ba, bai yi muni sosai ba."\nGaskiya: Ƙwaƙwalwar ajiya a ƙarƙashin matsananciyar damuwa takan yi aiki daban -- ƙwaƙwalwar da ba cikakke ba ko ba a sarari take ba abu ne na kowa kuma ba yana nufin abin da ya faru bai yi muni ba.\n\nƘarya: "Mutane masu ƙarfi ba sa shafa da rauni."\nGaskiya: Martanin rauni dabi''a ce ta tsarin jiki, ba lahani na hali ko alamar rauni ba.\n\nƘarya: "Lokaci kaɗai yake warkar da rauni."\nGaskiya: Lokaci na iya taimakawa, amma fahimta, tallafi, kuma wani lokacin kulawar ƙwararru ne yawanci ke motsa mutum daga tsira zuwa warkewa.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'ig',
        'Okwu Ụgha A Na-Ekwukarị Gbasara Mmerụ Ahụ-Mmụọ',
        'Okwu ụgha: "Naanị ihe ndakpọ ukwuu (agha, ọdachi) na-akpata mmerụ ahụ-mmụọ."\nEziokwu: Mmerụ ahụ-mmụọ nwere ike isi n''ụdị ahụmahụ dị iche iche, gụnyere ndị dị ka "nkịtị" site n''èzí, dị ka nrụgide na-adịgide, ịkpagbu, ma ọ bụ ụlọ na-enweghị nchekwa.\n\nOkwu ụgha: "Ọ bụrụ na ị nchetaghị ya nke ọma, ọ dịghị njọ."\nEziokwu: Ncheta n''okpuru nrụgide dị ukwuu na-arụkarị ọrụ n''ụzọ dị iche -- ncheta na-adịghị zuru oke ma ọ bụ na-edoghị anya bụ ihe a na-ahụkarị ma ọ bụghị na ọ pụtara na ahụmahụ ahụ dị nta.\n\nOkwu ụgha: "Mmerụ ahụ-mmụọ anaghị emetụta ndị siri ike."\nEziokwu: Mmeghachi omume mmerụ ahụ-mmụọ bụ omume sistemu akwara ozi nkịtị, ọ bụghị ntụpọ agwa ma ọ bụ ihe ngosi adịghị ike.\n\nOkwu ụgha: "Oge naanị na-agwọ mmerụ ahụ-mmụọ."\nEziokwu: Oge nwere ike inye aka, mana nghọta, nkwado, na mgbe ụfọdụ nlekọta ndị ọkachamara ka na-emekarị ka mmadụ si na ịdị ndụ gaa n''ọgwụgwọ.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'fr',
        'Mythes courants sur le traumatisme',
        'Mythe : « Seuls les événements catastrophiques (guerre, catastrophes) causent des traumatismes. »\nRéalité : Le traumatisme peut provenir de nombreux types d''expériences, y compris celles qui semblent « ordinaires » de l''extérieur, comme le stress chronique, le harcèlement ou un foyer instable.\n\nMythe : « Si vous ne vous en souvenez pas clairement, ce n''était pas si grave. »\nRéalité : La mémoire sous un stress extrême fonctionne souvent différemment -- des souvenirs fragmentés ou flous sont courants et ne signifient pas qu''une expérience était mineure.\n\nMythe : « Les personnes fortes ne sont pas affectées par les traumatismes. »\nRéalité : Les réactions traumatiques sont une réaction normale du système nerveux, pas un défaut de caractère ni un signe de faiblesse.\n\nMythe : « Le temps seul guérit les traumatismes. »\nRéalité : Le temps peut aider, mais la compréhension, le soutien et parfois des soins professionnels sont généralement ce qui fait réellement passer une personne de la survie à la guérison.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'sw',
        'Hadithi za Kawaida Kuhusu Majeraha ya Kiakili',
        'Hadithi: "Matukio makubwa tu (vita, maafa) husababisha majeraha ya kiakili."\nUkweli: Jeraha la kiakili laweza kutokana na aina nyingi za uzoefu, ikiwemo zile zinazoonekana "za kawaida" kwa nje, kama msongo wa kudumu, unyanyasaji, au nyumbani pasipo utulivu.\n\nHadithi: "Kama hukumbuki vizuri, haikuwa mbaya sana."\nUkweli: Kumbukumbu chini ya msongo mkubwa mara nyingi hufanya kazi tofauti -- kumbukumbu zilizogawanyika au zisizo wazi ni za kawaida na hazimaanishi kuwa tukio lilikuwa dogo.\n\nHadithi: "Watu wenye nguvu hawaathiriwi na majeraha ya kiakili."\nUkweli: Miitikio ya jeraha la kiakili ni mwitikio wa kawaida wa mfumo wa neva, si kasoro ya tabia au ishara ya udhaifu.\n\nHadithi: "Muda pekee huponya jeraha la kiakili."\nUkweli: Muda waweza kusaidia, lakini uelewa, msaada, na wakati mwingine huduma za kitaalamu ndizo kwa kawaida huhamisha mtu kutoka kuishi tu kwenda kupona.'
    ),
    (
        '00000000-0000-0000-0000-000000000102', 'ar',
        'خرافات شائعة حول الصدمة',
        'خرافة: "الأحداث الكارثية فقط (الحرب، الكوارث) تسبب الصدمة."\nالحقيقة: يمكن أن تنشأ الصدمة من أنواع عديدة من التجارب، بما في ذلك تلك التي تبدو "عادية" من الخارج، مثل التوتر المزمن أو التنمر أو المنزل غير المستقر.\n\nخرافة: "إذا لم تتذكره بوضوح، فلم يكن الأمر بهذا السوء."\nالحقيقة: الذاكرة تحت ضغط شديد غالبًا ما تعمل بشكل مختلف -- الذكريات المجزأة أو غير الواضحة شائعة ولا تعني أن التجربة كانت بسيطة.\n\nخرافة: "الأشخاص الأقوياء لا يتأثرون بالصدمة."\nالحقيقة: استجابات الصدمة هي رد فعل طبيعي للجهاز العصبي، وليست عيبًا في الشخصية أو علامة ضعف.\n\nخرافة: "الوقت وحده يشفي الصدمة."\nالحقيقة: يمكن للوقت أن يساعد، لكن الفهم والدعم وأحيانًا الرعاية المتخصصة هي عادة ما ينقل الشخص فعليًا من النجاة إلى الشفاء.'
    ),

    (
        '00000000-0000-0000-0000-000000000103', 'pcm',
        'How Trauma Dey Affect Brain',
        'Wen your brain sense danger, e go dey prioritize say make you survive pass say make you think am well well. One part of brain wey dem dey call amygdala dey work like smoke alarm, e go trigger stress response before your slow, more thinking part (prefrontal cortex) go fit contribute well well.\n\nNa why, for frightening moment, people dey react sharp sharp instead of "thinking am through" -- na so nervous system suppose work.\n\nIf stress continue happen for long time or e dey repeat, this alarm system fit become more sensitive, e go still dey "alert" even for safe situation. E no be say something wrong with you; na pattern wey you learn wey you fit still unlearn with time, safety, and support. Later levels for this course go dive more enter how to recognize and work with these responses.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'yo',
        'Bí Ìpalára-Ọkàn Ṣe Ń Nípa Lórí Ọpọlọ',
        'Nígbà tí ọpọlọ rẹ bá mọ ewu, yóò kọ́kọ́ fẹ́ kí o yè bó ju kí o ronú jinlẹ̀ lọ. Apá kan tí à ń pè ní amygdala ń ṣiṣẹ́ bí ẹ̀rọ ìkìlọ̀ iná, yóò máa mú ìdáhùn másùnmáwo jáde kí ìrònú rẹ tó lọ́ra, tó jinlẹ̀ (nínú prefrontal cortex) tó lè wọlé dé kíkún.\n\nÌdí nìyí tí, nígbà ìṣẹ̀lẹ̀ amúnibẹ̀rù, àwọn ènìyàn fi máa ń dáhùn kánkán dípò "rírò jinlẹ̀" -- bẹ́ẹ̀ gẹ́lẹ́ ni ètò ara ṣe yẹ kó ṣiṣẹ́.\n\nTí másùnmáwo bá ń bá a nìṣó tàbí tó ń ṣẹlẹ̀ léraléra, ẹ̀rọ ìkìlọ̀ yìí lè di ẹni tó máa ń yára dáhùn jù, yóò sì máa wà ní "ìṣọ́ra" kódà nínú ipò tó ní ààbò. Èyí kì í ṣe àbùkù nínú rẹ; àpẹẹrẹ tí o kọ́ ni, èyí tí o tún lè gbàgbé pẹ̀lú àkókò, ààbò, àti ìtìlẹ́yìn. Àwọn pele tó ń bọ̀ nínú ẹ̀kọ́ yìí yóò jinlẹ̀ síi nípa bí a ṣe ń dá wọn mọ̀ tí a sì ń ṣiṣẹ́ pẹ̀lú wọn.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'ha',
        'Yadda Rauni Ke Shafar Kwakwalwa',
        'Lokacin da kwakwalwarka ta ji haɗari, takan ba da fifiko ga tsira fiye da tunani mai zurfi. Wani sashe mai suna amygdala yana aiki kamar ƙararrawar hayaƙi, yana jawo martanin damuwa kafin sashin tunani mai zurfi, mai jinkiri (prefrontal cortex) ya sami dama ya shiga sosai.\n\nShi ya sa, a lokacin firgita, mutane galibi sukan amsa nan take maimakon "yin tunani sosai" -- haka tsarin jiki ke yi daidai da yadda aka tsara shi.\n\nIdan damuwa ta ci gaba ko ta maimaita akai-akai, wannan tsarin ƙararrawa na iya zama mai saurin amsawa, yana zama a "shirye" ko da a cikin yanayi mai aminci. Wannan ba lahani bane a gare ka; tsari ne da aka koya wanda kuma za a iya sakewa da lokaci, aminci, da tallafi. Matakan gaba a wannan darasi za su zurfafa kan yadda ake gane da aiki da waɗannan martani.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'ig',
        'Otu Mmerụ Ahụ-Mmụọ Si Emetụta Ụbụrụ',
        'Mgbe ụbụrụ gị chọpụtara ihe egwu, ọ na-ebute ụzọ ndụ karịa echiche miri emi. Otu akụkụ a na-akpọ amygdala na-arụ ọrụ dị ka mgbaàmà ọkụ anwụrụ, na-akpali mmeghachi omume nrụgide tupu akparamagwa gị dị nwayọọ, nke dị omimi (prefrontal cortex) enwee ohere itinye ọnụ nke ọma.\n\nỌ bụ ya mere, n''oge ụjọ, ndị mmadụ na-emekarị ngwa ngwa kama "iche echiche miri emi" -- nke ahụ bụ ihe sistemu akwara ozi kwesịrị ime.\n\nỌ bụrụ na nrụgide na-aga n''ihu ma ọ bụ na-ekpughe ọtụtụ ugboro, usoro mgbaàmà a nwere ike ịghọ nke na-adị mfe ịkpali, na-anọgide na "nlebara anya" ọbụna n''ọnọdụ dị nchebe. Nke a abụghị ntụpọ dị n''ime gị; ọ bụ usoro ị mụtara nke a pụrụ ịhapụ ya site n''oge, nchekwa, na nkwado. Ọkwa ndị na-esote na nkuzi a ga-abawanye omimi gbasara ịmata na ịrụkọ ọnụ na mmeghachi omume ndị a.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'fr',
        'Comment le traumatisme affecte le cerveau',
        'Lorsque votre cerveau détecte un danger, il privilégie la survie plutôt que la réflexion. Une région appelée l''amygdale agit comme un détecteur de fumée, déclenchant une réaction de stress avant que votre réflexion plus lente et plus posée (dans le cortex préfrontal) n''ait pleinement l''occasion d''intervenir.\n\nC''est pourquoi, dans un moment effrayant, les gens réagissent souvent instantanément plutôt que de « réfléchir posément » -- c''est exactement ce pour quoi le système nerveux a évolué.\n\nAvec un stress répété ou prolongé, ce système d''alarme peut devenir plus sensible, restant « en alerte » même dans des situations sûres. Ce n''est pas un défaut chez vous ; c''est un schéma appris qui peut aussi être désappris avec du temps, de la sécurité et du soutien. Les niveaux suivants de ce cours approfondissent la reconnaissance et le travail sur ces réactions.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'sw',
        'Jinsi Jeraha la Kiakili Linavyoathiri Ubongo',
        'Ubongo wako unapohisi hatari, hutanguliza kuishi kuliko kutafakari. Sehemu iitwayo amygdala hufanya kazi kama king''ora cha moshi, ikichochea mwitikio wa msongo kabla fikra zako za polepole, za kina zaidi (katika prefrontal cortex) hazijapata nafasi kamili ya kuchangia.\n\nHii ndiyo sababu, katika wakati wa kutisha, watu mara nyingi hujibu papo hapo badala ya "kufikiria kwa kina" -- huo ndio mfumo wa neva ukifanya kile ulichobuniwa kufanya.\n\nKwa msongo unaojirudia au wa muda mrefu, mfumo huu wa king''ora waweza kuwa nyeti zaidi, ukibaki "tayari" hata katika hali salama. Hiyo si kasoro ndani yako; ni mfumo uliojifunza ambao pia waweza kuusahau kwa muda, usalama, na msaada. Viwango vijavyo katika kozi hii vitaenda ndani zaidi kuhusu kutambua na kufanya kazi na miitikio hii.'
    ),
    (
        '00000000-0000-0000-0000-000000000103', 'ar',
        'كيف تؤثر الصدمة على الدماغ',
        'عندما يستشعر دماغك الخطر، يعطي الأولوية للنجاة على التفكير المتأني. منطقة تسمى اللوزة الدماغية تعمل كجهاز إنذار الدخان، فتُطلق استجابة التوتر قبل أن تحصل عملية التفكير الأبطأ والأكثر تأملاً (في القشرة الجبهية) على فرصة كاملة للمشاركة.\n\nلهذا السبب، في اللحظة المخيفة، غالبًا ما يتفاعل الناس فورًا بدلاً من "التفكير مليًا" -- وهذا بالضبط ما تطور الجهاز العصبي ليقوم به.\n\nمع التوتر المتكرر أو المطول، يمكن أن يصبح نظام الإنذار هذا أكثر حساسية، ويبقى "في حالة تأهب" حتى في المواقف الآمنة. هذا ليس عيبًا فيك؛ إنه نمط مكتسب يمكن أيضًا التخلص منه بمرور الوقت والأمان والدعم. المستويات اللاحقة في هذه الدورة تتعمق أكثر في التعرف على هذه الاستجابات والعمل معها.'
    );

-- ============================================================
-- Quiz questions (3 questions x 7 locales = 21 rows)
-- ============================================================
INSERT INTO quiz_question_translations (question_id, locale, text) VALUES
    ('00000000-0000-0000-0000-000000000301', 'pcm', 'Wetin dey determine if event go be trauma for person?'),
    ('00000000-0000-0000-0000-000000000301', 'yo', 'Kín ni ó ń pinnu bóyá ìṣẹ̀lẹ̀ kan jẹ́ ìpalára-ọkàn fún ẹnikẹ́ni?'),
    ('00000000-0000-0000-0000-000000000301', 'ha', 'Me ke ƙayyade ko wani abu ya zama rauni ga mutum?'),
    ('00000000-0000-0000-0000-000000000301', 'ig', 'Gịnị na-ekpebi ma ihe mere bụụrụ mmadụ mmerụ ahụ-mmụọ?'),
    ('00000000-0000-0000-0000-000000000301', 'fr', 'Qu''est-ce qui détermine si un événement est traumatisant pour quelqu''un ?'),
    ('00000000-0000-0000-0000-000000000301', 'sw', 'Ni nini huamua kama tukio ni la kiwewe kwa mtu?'),
    ('00000000-0000-0000-0000-000000000301', 'ar', 'ما الذي يحدد ما إذا كان الحدث صادمًا لشخص ما؟'),

    ('00000000-0000-0000-0000-000000000302', 'pcm', 'Which statement about trauma correct?'),
    ('00000000-0000-0000-0000-000000000302', 'yo', 'Èwo lára àwọn gbólóhùn yìí nípa ìpalára-ọkàn ni ó jẹ́ òtítọ́?'),
    ('00000000-0000-0000-0000-000000000302', 'ha', 'Wane bayani game da rauni ne daidai?'),
    ('00000000-0000-0000-0000-000000000302', 'ig', 'Olee nkwupụta gbasara mmerụ ahụ-mmụọ bụ eziokwu?'),
    ('00000000-0000-0000-0000-000000000302', 'fr', 'Quelle affirmation sur le traumatisme est exacte ?'),
    ('00000000-0000-0000-0000-000000000302', 'sw', 'Ni kauli ipi kuhusu jeraha la kiakili iliyo sahihi?'),
    ('00000000-0000-0000-0000-000000000302', 'ar', 'أي عبارة حول الصدمة صحيحة؟'),

    ('00000000-0000-0000-0000-000000000303', 'pcm', 'Wetin people dey compare amygdala to?'),
    ('00000000-0000-0000-0000-000000000303', 'yo', 'Kín ni wọ́n sábà máa ń fi amygdala wé?'),
    ('00000000-0000-0000-0000-000000000303', 'ha', 'Me ake kwatanta amygdala da shi akai-akai?'),
    ('00000000-0000-0000-0000-000000000303', 'ig', 'Gịnị ka a na-atụnyekarị amygdala?'),
    ('00000000-0000-0000-0000-000000000303', 'fr', 'À quoi l''amygdale est-elle souvent comparée ?'),
    ('00000000-0000-0000-0000-000000000303', 'sw', 'Amygdala hufananishwa na nini mara nyingi?'),
    ('00000000-0000-0000-0000-000000000303', 'ar', 'بم تُقارن اللوزة الدماغية غالبًا؟');

-- ============================================================
-- Quiz options (12 options x 7 locales = 84 rows)
-- ============================================================
INSERT INTO quiz_option_translations (option_id, locale, text) VALUES
    -- Q1 options
    ('00000000-0000-0000-0000-000000000401', 'pcm', 'Na only how bad the event look to other people'),
    ('00000000-0000-0000-0000-000000000401', 'yo', 'Bí ìṣẹ̀lẹ̀ náà ṣe burú tó sí ojú àwọn ẹlòmíràn nìkan'),
    ('00000000-0000-0000-0000-000000000401', 'ha', 'Yadda al''amarin ya yi muni kawai a idon wasu'),
    ('00000000-0000-0000-0000-000000000401', 'ig', 'Naanị ka ihe ahụ si dị egwu n''anya ndị ọzọ'),
    ('00000000-0000-0000-0000-000000000401', 'fr', 'Seulement la gravité apparente de l''événement pour les autres'),
    ('00000000-0000-0000-0000-000000000401', 'sw', 'Jinsi tukio lilivyoonekana kuwa kubwa kwa wengine tu'),
    ('00000000-0000-0000-0000-000000000401', 'ar', 'فقط مدى شدة الحدث في نظر الآخرين'),

    ('00000000-0000-0000-0000-000000000402', 'pcm', 'How your body and mind take feel am that time'),
    ('00000000-0000-0000-0000-000000000402', 'yo', 'Bí ara àti ọkàn rẹ ṣe nírìírí rẹ̀ nígbà yẹn'),
    ('00000000-0000-0000-0000-000000000402', 'ha', 'Yadda jikinka da hankalinka suka fuskance shi a lokacin'),
    ('00000000-0000-0000-0000-000000000402', 'ig', 'Otu ahụ na uche gị si nwee mmetụta ya n''oge ahụ'),
    ('00000000-0000-0000-0000-000000000402', 'fr', 'La façon dont votre corps et votre esprit l''ont vécu sur le moment'),
    ('00000000-0000-0000-0000-000000000402', 'sw', 'Jinsi mwili na akili yako vilivyolipitia wakati huo'),
    ('00000000-0000-0000-0000-000000000402', 'ar', 'كيف عاشه جسدك وعقلك في ذلك الوقت'),

    ('00000000-0000-0000-0000-000000000403', 'pcm', 'Whether e happen more than once'),
    ('00000000-0000-0000-0000-000000000403', 'yo', 'Bóyá ó ṣẹlẹ̀ ju ẹ̀ẹ̀kan lọ'),
    ('00000000-0000-0000-0000-000000000403', 'ha', 'Ko ya faru fiye da sau ɗaya'),
    ('00000000-0000-0000-0000-000000000403', 'ig', 'Ma ọ bụrụ na ọ mere ihe karịrị otu ugboro'),
    ('00000000-0000-0000-0000-000000000403', 'fr', 'S''il s''est produit plus d''une fois'),
    ('00000000-0000-0000-0000-000000000403', 'sw', 'Kama lilitokea zaidi ya mara moja'),
    ('00000000-0000-0000-0000-000000000403', 'ar', 'ما إذا كان قد حدث أكثر من مرة'),

    ('00000000-0000-0000-0000-000000000404', 'pcm', 'How long the thing happen'),
    ('00000000-0000-0000-0000-000000000404', 'yo', 'Bí ó ti pẹ́ tó tí ó ti ṣẹlẹ̀'),
    ('00000000-0000-0000-0000-000000000404', 'ha', 'Tsawon lokacin da ya gabata tun da ya faru'),
    ('00000000-0000-0000-0000-000000000404', 'ig', 'Ogologo oge ọ mesịrị'),
    ('00000000-0000-0000-0000-000000000404', 'fr', 'Depuis combien de temps cela s''est produit'),
    ('00000000-0000-0000-0000-000000000404', 'sw', 'Muda gani uliopita tangu litokee'),
    ('00000000-0000-0000-0000-000000000404', 'ar', 'منذ متى حدث'),

    -- Q2 options
    ('00000000-0000-0000-0000-000000000411', 'pcm', 'Trauma reaction na sign say person weak'),
    ('00000000-0000-0000-0000-000000000411', 'yo', 'Ìdáhùn ìpalára-ọkàn jẹ́ àmì àìlera'),
    ('00000000-0000-0000-0000-000000000411', 'ha', 'Martanin rauni alama ce ta rauni'),
    ('00000000-0000-0000-0000-000000000411', 'ig', 'Mmeghachi omume mmerụ ahụ-mmụọ bụ ihe ngosi adịghị ike'),
    ('00000000-0000-0000-0000-000000000411', 'fr', 'Les réactions traumatiques sont un signe de faiblesse'),
    ('00000000-0000-0000-0000-000000000411', 'sw', 'Miitikio ya kiwewe ni ishara ya udhaifu'),
    ('00000000-0000-0000-0000-000000000411', 'ar', 'استجابات الصدمة علامة على الضعف'),

    ('00000000-0000-0000-0000-000000000412', 'pcm', 'Na only big wahala dey cause trauma'),
    ('00000000-0000-0000-0000-000000000412', 'yo', 'Ìpalára-ọkàn lè wá láti ìṣẹ̀lẹ̀ ńlá nìkan'),
    ('00000000-0000-0000-0000-000000000412', 'ha', 'Rauni na iya fitowa daga manyan bala''o''i kaɗai'),
    ('00000000-0000-0000-0000-000000000412', 'ig', 'Mmerụ ahụ-mmụọ nwere ike isi naanị n''ihe ndakpọ ukwuu'),
    ('00000000-0000-0000-0000-000000000412', 'fr', 'Le traumatisme ne peut provenir que d''événements catastrophiques'),
    ('00000000-0000-0000-0000-000000000412', 'sw', 'Jeraha la kiakili laweza kutokana na matukio makubwa pekee'),
    ('00000000-0000-0000-0000-000000000412', 'ar', 'لا يمكن أن تنشأ الصدمة إلا من أحداث كارثية'),

    ('00000000-0000-0000-0000-000000000413', 'pcm', 'Broken memory na normal reaction to serious stress'),
    ('00000000-0000-0000-0000-000000000413', 'yo', 'Ìrántí tí ó yapa jẹ́ ìdáhùn tó wọ́pọ̀, tó sì jẹ́ déédé sí másùnmáwo ńlá'),
    ('00000000-0000-0000-0000-000000000413', 'ha', 'Ƙwaƙwalwar da ba cikakke ba martani ne na kowa, na al''ada ga matsananciyar damuwa'),
    ('00000000-0000-0000-0000-000000000413', 'ig', 'Ncheta na-adịghị zuru oke bụ mmeghachi omume a na-ahụkarị, nke nkịtị maka nrụgide dị ukwuu'),
    ('00000000-0000-0000-0000-000000000413', 'fr', 'Les souvenirs fragmentés sont une réaction courante et normale à un stress extrême'),
    ('00000000-0000-0000-0000-000000000413', 'sw', 'Kumbukumbu zilizogawanyika ni mwitikio wa kawaida kwa msongo mkubwa'),
    ('00000000-0000-0000-0000-000000000413', 'ar', 'الذكريات المجزأة استجابة شائعة وطبيعية للضغط الشديد'),

    ('00000000-0000-0000-0000-000000000414', 'pcm', 'Time alone always dey enough to heal trauma'),
    ('00000000-0000-0000-0000-000000000414', 'yo', 'Àkókò nìkan ló máa ń tó láti wo ìpalára-ọkàn sàn nígbà gbogbo'),
    ('00000000-0000-0000-0000-000000000414', 'ha', 'Lokaci kaɗai koyaushe ya isa ya warkar da rauni'),
    ('00000000-0000-0000-0000-000000000414', 'ig', 'Oge naanị na-ezuzu mgbe niile iji gwọọ mmerụ ahụ-mmụọ'),
    ('00000000-0000-0000-0000-000000000414', 'fr', 'Le temps seul suffit toujours à guérir un traumatisme'),
    ('00000000-0000-0000-0000-000000000414', 'sw', 'Muda pekee daima hutosha kuponya jeraha la kiakili'),
    ('00000000-0000-0000-0000-000000000414', 'ar', 'الوقت وحده يكفي دائمًا لشفاء الصدمة'),

    -- Q3 options
    ('00000000-0000-0000-0000-000000000421', 'pcm', 'Laibrari'),
    ('00000000-0000-0000-0000-000000000421', 'yo', 'Ilé-ìkàwé'),
    ('00000000-0000-0000-0000-000000000421', 'ha', 'Laburare'),
    ('00000000-0000-0000-0000-000000000421', 'ig', 'Ụlọ akwụkwọ'),
    ('00000000-0000-0000-0000-000000000421', 'fr', 'Une bibliothèque'),
    ('00000000-0000-0000-0000-000000000421', 'sw', 'Maktaba'),
    ('00000000-0000-0000-0000-000000000421', 'ar', 'مكتبة'),

    ('00000000-0000-0000-0000-000000000422', 'pcm', 'Smoke alarm'),
    ('00000000-0000-0000-0000-000000000422', 'yo', 'Ẹ̀rọ ìkìlọ̀ iná'),
    ('00000000-0000-0000-0000-000000000422', 'ha', 'Ƙararrawar hayaƙi'),
    ('00000000-0000-0000-0000-000000000422', 'ig', 'Mgbaàmà ọkụ anwụrụ'),
    ('00000000-0000-0000-0000-000000000422', 'fr', 'Un détecteur de fumée'),
    ('00000000-0000-0000-0000-000000000422', 'sw', 'King''ora cha moshi'),
    ('00000000-0000-0000-0000-000000000422', 'ar', 'جهاز إنذار الدخان'),

    ('00000000-0000-0000-0000-000000000423', 'pcm', 'Thermostat'),
    ('00000000-0000-0000-0000-000000000423', 'yo', 'Ẹ̀rọ amóoru'),
    ('00000000-0000-0000-0000-000000000423', 'ha', 'Na''urar sarrafa zafi'),
    ('00000000-0000-0000-0000-000000000423', 'ig', 'Ngwaọrụ njikwa okpomọkụ'),
    ('00000000-0000-0000-0000-000000000423', 'fr', 'Un thermostat'),
    ('00000000-0000-0000-0000-000000000423', 'sw', 'Kipima joto'),
    ('00000000-0000-0000-0000-000000000423', 'ar', 'منظم حرارة'),

    ('00000000-0000-0000-0000-000000000424', 'pcm', 'Filing cabinet'),
    ('00000000-0000-0000-0000-000000000424', 'yo', 'Pápá ìpamọ́ ìwé'),
    ('00000000-0000-0000-0000-000000000424', 'ha', 'Akwatin ajiye fayiloli'),
    ('00000000-0000-0000-0000-000000000424', 'ig', 'Igbe nchekwa akwụkwọ'),
    ('00000000-0000-0000-0000-000000000424', 'fr', 'Un classeur'),
    ('00000000-0000-0000-0000-000000000424', 'sw', 'Kabati la kuhifadhi faili'),
    ('00000000-0000-0000-0000-000000000424', 'ar', 'خزانة ملفات');

-- ============================================================
-- Badges (6 badges x 7 locales = 42 rows)
-- ============================================================
INSERT INTO badge_translations (badge_id, locale, name, description) VALUES
    ('00000000-0000-0000-0000-000000000501', 'pcm', 'Beginner', 'Finish Beginner level: trauma basics, lies, and brain science.'),
    ('00000000-0000-0000-0000-000000000501', 'yo', 'Olùbẹ̀rẹ̀', 'O parí ipele Ìbẹ̀rẹ̀: ìpilẹ̀ ìpalára-ọkàn, irọ́, àti ìmọ̀ ọpọlọ.'),
    ('00000000-0000-0000-0000-000000000501', 'ha', 'Mafara', 'Ka kammala matakin Mafara: asalin rauni, ƙarairayi, da ilimin kwakwalwa.'),
    ('00000000-0000-0000-0000-000000000501', 'ig', 'Mmalite', 'Ị rụchara ọkwa Mmalite: ihe bụ mmerụ ahụ-mmụọ, okwu ụgha, na sayensị ụbụrụ.'),
    ('00000000-0000-0000-0000-000000000501', 'fr', 'Débutant', 'A terminé le niveau Débutant : bases du traumatisme, mythes et neurosciences.'),
    ('00000000-0000-0000-0000-000000000501', 'sw', 'Mwanzo', 'Amekamilisha kiwango cha Mwanzo: misingi ya majeraha ya kiakili, hadithi za uwongo, na sayansi ya ubongo.'),
    ('00000000-0000-0000-0000-000000000501', 'ar', 'مبتدئ', 'أكمل مستوى المبتدئ: أساسيات الصدمة والخرافات وعلم الدماغ.'),

    ('00000000-0000-0000-0000-000000000502', 'pcm', 'Know Yourself', 'Finish Self-Awareness level.'),
    ('00000000-0000-0000-0000-000000000502', 'yo', 'Onímọ̀ Ara-ẹni', 'O parí ipele Ìmọ̀ Ara-ẹni.'),
    ('00000000-0000-0000-0000-000000000502', 'ha', 'Sanin Kai', 'Ka kammala matakin Sanin Kai.'),
    ('00000000-0000-0000-0000-000000000502', 'ig', 'Onye Ma Onwe Ya', 'Ị rụchara ọkwa Ịmara Onwe Gị.'),
    ('00000000-0000-0000-0000-000000000502', 'fr', 'Conscient de soi', 'A terminé le niveau Conscience de soi.'),
    ('00000000-0000-0000-0000-000000000502', 'sw', 'Anayejitambua', 'Amekamilisha kiwango cha Kujitambua.'),
    ('00000000-0000-0000-0000-000000000502', 'ar', 'واعٍ ذاتيًا', 'أكمل مستوى الوعي الذاتي.'),

    ('00000000-0000-0000-0000-000000000503', 'pcm', 'Calm Body', 'Finish Regulation level.'),
    ('00000000-0000-0000-0000-000000000503', 'yo', 'Olùṣàkóso', 'O parí ipele Ìṣàkóso.'),
    ('00000000-0000-0000-0000-000000000503', 'ha', 'Sarrafaffe', 'Ka kammala matakin Sarrafa Kai.'),
    ('00000000-0000-0000-0000-000000000503', 'ig', 'Onye Jikwara Onwe Ya', 'Ị rụchara ọkwa Njikwa Onwe Gị.'),
    ('00000000-0000-0000-0000-000000000503', 'fr', 'Régulé', 'A terminé le niveau Régulation.'),
    ('00000000-0000-0000-0000-000000000503', 'sw', 'Mwenye Udhibiti', 'Amekamilisha kiwango cha Udhibiti.'),
    ('00000000-0000-0000-0000-000000000503', 'ar', 'منظَّم', 'أكمل مستوى التنظيم.'),

    ('00000000-0000-0000-0000-000000000504', 'pcm', 'Connected', 'Finish Relationships level.'),
    ('00000000-0000-0000-0000-000000000504', 'yo', 'Onídàpọ̀mọ́ra', 'O parí ipele Àjọṣepọ̀.'),
    ('00000000-0000-0000-0000-000000000504', 'ha', 'Mai Haɗi', 'Ka kammala matakin Dangantaka.'),
    ('00000000-0000-0000-0000-000000000504', 'ig', 'Onye Jikọrọ Ọnụ', 'Ị rụchara ọkwa Mmekọrịta.'),
    ('00000000-0000-0000-0000-000000000504', 'fr', 'Connecté', 'A terminé le niveau Relations.'),
    ('00000000-0000-0000-0000-000000000504', 'sw', 'Aliyeunganika', 'Amekamilisha kiwango cha Mahusiano.'),
    ('00000000-0000-0000-0000-000000000504', 'ar', 'مرتبِط', 'أكمل مستوى العلاقات.'),

    ('00000000-0000-0000-0000-000000000505', 'pcm', 'Growing', 'Finish Growth level.'),
    ('00000000-0000-0000-0000-000000000505', 'yo', 'Olùdàgbà', 'O parí ipele Ìdàgbàsókè.'),
    ('00000000-0000-0000-0000-000000000505', 'ha', 'Mai Ci Gaba', 'Ka kammala matakin Ci Gaba.'),
    ('00000000-0000-0000-0000-000000000505', 'ig', 'Onye Na-eto Eto', 'Ị rụchara ọkwa Uto.'),
    ('00000000-0000-0000-0000-000000000505', 'fr', 'En croissance', 'A terminé le niveau Croissance.'),
    ('00000000-0000-0000-0000-000000000505', 'sw', 'Anayekua', 'Amekamilisha kiwango cha Ukuaji.'),
    ('00000000-0000-0000-0000-000000000505', 'ar', 'نامٍ', 'أكمل مستوى النمو.'),

    ('00000000-0000-0000-0000-000000000506', 'pcm', 'Trauma-Informed Leader', 'Finish Advanced level.'),
    ('00000000-0000-0000-0000-000000000506', 'yo', 'Aṣáájú Onímọ̀ Ìpalára-Ọkàn', 'O parí ipele Onípele Gíga.'),
    ('00000000-0000-0000-0000-000000000506', 'ha', 'Jagora Mai Sanin Rauni', 'Ka kammala matakin Ci-gaba.'),
    ('00000000-0000-0000-0000-000000000506', 'ig', 'Onye Ndu Maara Mmerụ Ahụ-Mmụọ', 'Ị rụchara ọkwa Ọkwa Elu.'),
    ('00000000-0000-0000-0000-000000000506', 'fr', 'Leader tenant compte des traumatismes', 'A terminé le niveau Avancé.'),
    ('00000000-0000-0000-0000-000000000506', 'sw', 'Kiongozi Anayezingatia Majeraha ya Kiakili', 'Amekamilisha kiwango cha Juu Zaidi.'),
    ('00000000-0000-0000-0000-000000000506', 'ar', 'قائد واعٍ بالصدمة', 'أكمل مستوى متقدم.');
