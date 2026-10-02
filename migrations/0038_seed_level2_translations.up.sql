-- LLM-drafted translations (needs_review = TRUE throughout, per
-- migrations 0015-0019's column default) of Level 2's lesson content
-- (seeded by migration 0033), into the 7 non-English locales from
-- learning.SupportedLocales: pcm, yo, ha, ig, fr, sw, ar.
--
-- These are LLM-drafted, NOT reviewed by a native-speaking professional
-- translator or a clinician for this trauma-sensitive subject matter —
-- same posture as migration 0021's Level 1 translations. Do not present
-- this copy to end users as finished.

-- ============================================================
-- Lessons (3 lessons x 7 locales = 21 rows)
-- ============================================================
INSERT INTO lesson_translations (lesson_id, locale, title, content) VALUES
    ('00000000-0000-0000-0000-000000001101', 'pcm', 'The Four Wahala Responses', 'When your body sense say trouble dey come, e no dey wait for permission before e react. Many people don hear about "fight or flight," but sef na four common patterns dey: fight (face the wahala), flight (run comot), freeze (shut down, you no fit move or talk), and fawn (dey beg or try please people make you stay safe).\n\nNone of dis one bad or weak. Each one na survival style wey your body learn say e dey work for one kind situation, many times e come from wetin happen to you before. Person wey grow for house wey no steady fit dey always try please people -- dey check the room for tension make e calm am down -- even after e don comot from dat environment, because the pattern don keep am safe before.\n\nTo know which response you dey normally show no be say make you judge yourself. Na the first step to know say your body dey react to wetin dey happen now or e just dey replay old survival pattern wey no fit the matter again.'),
    ('00000000-0000-0000-0000-000000001101', 'yo', 'Ìdáhùn Ìpalára-Ọkàn Mẹ́rin', 'Nígbà tí ara rẹ bá mọ̀ pé ewu ń bọ̀, kò dúró de ìgbà láti fèsì. Ọ̀pọ̀ènìyàn ti gbọ́ nípa "ìjà tàbí ìsá," ṣùgbọ́n àwọn ìlànà mẹ́rin tó wọ́pọ̀ ló wà ní pàtàkì: ìjà (dojúkọ ewu náà), ìsá (sá fún un), dídi-dúró (dídúró pátápátá, àìlè gbéra tàbí sọ̀rọ̀), àti ìtẹríba (ṣíṣe ìtùnú tàbí ìtẹríba láti wà láàbò).\n\nKò sí ọ̀kan lára wọn tí ó "burú" tàbí tí ó "jẹ́ àìlera." Ọ̀kọ̀ọ̀kan jẹ́ ọgbọ́n ààbò tí ara rẹ ti kọ́ pé ó ń ṣiṣẹ́ nínú ipò kan, tí ó sábà máa ń ní ìbátan pẹ̀lú àwọn ìrírí rẹ ní ìgbà àtijọ́. Ẹni tí a tọ́ dàgbà ní ilé tí kò ní ìdúróṣinṣin lè máa yan ìtẹríba -- yíyẹ̀wò yàrá fún másùnmáwo kí ó sì tù ú lójú -- pípẹ́ lẹ́yìn tí àyíká yẹn ti kọjá, nítorí pé ìlànà náà ti dáàbò bò ó nígbà kan rí.\n\nRírí ìdáhùn tí o máa ń lò kì í ṣe láti ṣe ìdájọ́ ara rẹ. Ó jẹ́ ìgbésẹ̀ àkọ́kọ́ láti mọ ìgbà tí ara rẹ ń fèsì sí ipò tí ó wà lọ́wọ́lọ́wọ́ dípò kí ó máa tún ìlànà ààbò àtijọ́ ṣe tí kò bá ipò yẹn mu mọ́.'),
    ('00000000-0000-0000-0000-000000001101', 'ha', 'Halayen Rauni Guda Hudu', 'Lokacin da jikinka ya ji akwai hatsari, ba ya jira izini kafin ya nuna martani. Mutane da yawa sun ji game da "yaki ko gudu," amma a hakika akwai tsare-tsare guda hudu da aka saba gani: yaki (fuskantar hatsarin), gudu (tserewa daga shi), tsayawa (daskarewa, rashin iya motsi ko magana), da sulhu (yin sulhu ko kokarin faranta wa mutane rai domin tsira).\n\nBabu wani daga cikinsu da ya kasance "ba daidai ba" ko "rauni." Kowanne dabara ce ta tsira da jikinka ya koya cewa tana aiki a wani yanayi, galibi saboda abubuwan da suka faru da kai a baya. Wanda aka girma a gidan da babu kwanciyar hankali zai iya zama mai sulhu a koyaushe -- yana duba dakin domin tashin hankali ya kuma shiryawa -- har bayan ya bar wannan muhallin, saboda wannan dabi''ar ta taba kiyaye shi lafiya.\n\nGane irin martanin da kake yawan nunawa ba yana nufin ka yi wa kanka hukunci ba. Shi ne matakin farko na gane lokacin da jikinka ke mayar da martani ga yanayin yanzu maimakon maimaita tsohuwar dabarar tsira da ba ta dace da al''amarin yanzu ba.'),
    ('00000000-0000-0000-0000-000000001101', 'ig', 'Ụzọ Anọ Ndị Mmerụ Ahụ-Mmụọ Si Eme', 'Mgbe ahụ gị chọpụtara ihe ize ndụ, ọ naghị eche ikike tupu ọ zaghachi. Ọtụtụ ndị mmadụ anụla maka "ọgụ ma ọ bụ ịgbapụ," mana e nwere ụzọ anọ a na-ahụkarị: ọgụ (iguzogide ihe ize ndụ ahụ), ịgbapụ (ịgba ọsọ ya), ịkwụsị (ịkwụsị kpamkpam, enweghị ike ịmegharị ahụ ma ọ bụ ikwu okwu), na nrube isi (ime ka ndị ọzọ nwee afọ ojuju ka ị nọrọ n''udo).\n\nỌ dịghị nke n''ime ha bụ "ihe na-ezighị ezi" ma ọ bụ "adịghị ike." Nke ọ bụla bụ atụmatụ nchekwa ahụ gị mụtara na ọ na-arụ ọrụ n''ọnọdụ ụfọdụ, nke na-abụkarị ihe metụtara ihe ndị mere na ndụ gị na mbụ. Onye a zụlitere n''ụlọ na-enweghị nchekwa nwere ike ịbụ onye na-arụba isi mgbe niile -- na-enyocha ime ụlọ maka ịma jijiji wee dozie ya -- ọbụlagodi mgbe gafesịrị n''ebe ahụ, n''ihi na ụkpụrụ ahụ chebere ya na mbụ.\n\nịmata ụdị azịza ị na-emekarị abụghị ikpe gị onwe gị. Ọ bụ nzọụkwụ mbụ iji mata mgbe ahụ gị na-aza ọnọdụ dị ugbu a karịa ime ka ọ bụrụ naanị omume nchekwa ochie nke na-adịghị agbaso ọnọdụ ahụ ọzọ.'),
    ('00000000-0000-0000-0000-000000001101', 'fr', 'Les Quatre Réactions au Traumatisme', 'Lorsque votre système nerveux perçoit une menace, il n''attend pas la permission pour réagir. La plupart des gens ont entendu parler du « combat ou fuite », mais il existe en réalité quatre schémas courants : le combat (affronter la menace), la fuite (s''en échapper), la sidération (se figer, incapable de bouger ou de parler), et la soumission (apaiser ou chercher à plaire pour rester en sécurité).\n\nAucune de ces réactions n''est « mauvaise » ou « faible ». Chacune est une stratégie de survie que votre corps a apprise comme étant efficace dans une situation donnée, souvent façonnée par des expériences antérieures de votre vie. Une personne élevée dans un foyer imprévisible peut avoir tendance à se soumettre par défaut -- scrutant la pièce à la recherche de tensions pour les apaiser -- longtemps après que cet environnement a disparu, parce que ce schéma l''a protégée autrefois.\n\nReconnaître la réaction vers laquelle vous tendez n''est pas une question de jugement envers vous-même. C''est la première étape pour remarquer quand votre corps réagit à une situation présente plutôt que de rejouer un ancien schéma de survie qui ne convient plus.'),
    ('00000000-0000-0000-0000-000000001101', 'sw', 'Miitikio Minne ya Kiwewe', 'Mfumo wako wa neva unapotambua hatari, haungoji ruhusa kabla ya kuitikia. Watu wengi wamesikia kuhusu "pigana au kimbia," lakini kwa hakika kuna mifumo minne ya kawaida: kupigana (kukabiliana na hatari), kukimbia (kuikwepa), kuganda (kuzimika kabisa, kushindwa kusogea au kuzungumza), na kuridhia (kutuliza au kujipendekeza ili kubaki salama).\n\nHakuna hata mmoja kati ya hii ulio "mbaya" au "udhaifu." Kila mmoja ni mkakati wa kuishi ambao mwili wako umejifunza unafanya kazi katika hali fulani, mara nyingi ukiundwa na matukio ya awali katika maisha yako. Mtu aliyekulia katika nyumba isiyotabirika anaweza kuwa na tabia ya kuridhia kila mara -- akichunguza chumba kutafuta mivutano na kuituliza -- hata muda mrefu baada ya mazingira hayo kutoweka, kwa sababu mkakati huo uliwahi kumlinda.\n\nKutambua ni mwitikio gani unaoelekea nao si kujihukumu wewe mwenyewe. Ni hatua ya kwanza ya kugundua wakati mwili wako unapoitikia hali ya sasa badala ya kurudia mkakati wa zamani wa kuishi ambao hauendani tena na hali hiyo.'),
    ('00000000-0000-0000-0000-000000001101', 'ar', 'استجابات الصدمة الأربع', 'عندما يستشعر جهازك العصبي تهديدًا، فإنه لا ينتظر الإذن قبل أن يتفاعل. سمع معظم الناس عن "القتال أو الهروب"، لكن هناك في الواقع أربعة أنماط شائعة: القتال (مواجهة التهديد)، الهروب (الفرار منه)، التجمد (التوقف التام، عدم القدرة على الحركة أو الكلام)، والاسترضاء (محاولة إرضاء الآخرين للبقاء آمنًا).\n\nلا شيء من هذه الاستجابات "خاطئ" أو "ضعيف". كل واحدة منها استراتيجية بقاء تعلّم جسدك أنها تنجح في موقف معين، غالبًا نتيجة تجارب سابقة في حياتك. قد يميل شخص نشأ في بيئة منزلية غير مستقرة إلى الاسترضاء بشكل دائم -- يفحص الغرفة بحثًا عن التوتر ليخفف منه -- حتى بعد مرور وقت طويل على زوال تلك البيئة، لأن هذا النمط حماه في وقت ما.\n\nإدراك الاستجابة التي تميل إليها ليس حكمًا على نفسك. إنها الخطوة الأولى لملاحظة متى يتفاعل جسدك مع موقف حالي بدلًا من إعادة تمثيل نمط بقاء قديم لم يعد مناسبًا.'),

    ('00000000-0000-0000-0000-000000001102', 'pcm', 'Make You Understand Your Triggers', 'Trigger na anything -- sound, smell, how person voice sound, one kain statement -- wey dey make your body react as if say old wahala dey happen again, even as you dey safe now. Trigger no need to make sense before e dey real; na your body im remember am through connection, e no be say you dey reason am well-well.\n\nWhen trigger catch you, e fit feel too much pass wetin just happen, and e fit confuse you or shame you. E good make you remember say how strong the reaction be, e dey connect to the first experience, e no mean say na because of wetin dey happen now.\n\nOne simple practice wey you fit start na to dey notice, after strong reaction, three things: wetin happen just before am, wetin you feel for your body, and wetin thought come with am. As time dey go, dis one go build map of your own triggers -- no be to run comot for life around dem, but to sabi say "na trigger be dis" instead of "something wrong with me."'),
    ('00000000-0000-0000-0000-000000001102', 'yo', 'Òye Nípa Àwọn Ohun Tó Ń Fa Ìdáhùn Nínú Rẹ', 'Ohun tí ó ń fa ìdáhùn (trigger) jẹ́ ohunkóhun -- ìró, òórùn, ohùn ẹnìkan, gbólóhùn kan pàtó -- tí ó máa ń mú kí ara rẹ fèsì bí ẹni pé ewu àtijọ́ ń ṣẹlẹ̀ lẹ́ẹ̀kan sí i, kódà nígbà tí o bá wà láàbò lọ́wọ́lọ́wọ́. Kò pọndandan kí ohun tí ó ń fa ìdáhùn yìí ní ìtumọ̀ onídàájọ́ kí ó tó jẹ́ òótọ́; ó jẹ́ ìrántí ìsopọ̀ ara rẹ, kì í ṣe àríyànjiyàn tí a rò jinlẹ̀.\n\nNígbà tí ohun kan bá fa ìdáhùn nínú rẹ, ó sábà máa ń dà bí ẹni pé ó pọ̀ jù fún ohun tí ó ṣẹlẹ̀ lọ́wọ́lọ́wọ́, èyí tí ó lè dàrú ọkàn tàbí tí ó lè tini. Ó máa ń ràn ọ́ lọ́wọ́ láti rántí pé bí ìdáhùn náà ṣe lágbára tó ní ìbátan pẹ̀lú ìrírí ìpilẹ̀ṣẹ̀ tí ó so pọ̀, kì í ṣe dandan sí àsìkò lọ́wọ́lọ́wọ́.\n\nÌṣe tí ó rọrùn láti bẹ̀rẹ̀ ni láti máa ṣàkíyèsí, lẹ́yìn ìdáhùn líle, ohun mẹ́ta: ohun tí ó ṣẹlẹ̀ ṣáájú rẹ̀, ohun tí o nímọ̀lára nínú ara rẹ, àti èrò tí ó wá pẹ̀lú rẹ̀. Ní àkókò, èyí máa ń kọ́ àwòrán àwọn ohun tí ń fa ìdáhùn rẹ -- kì í ṣe láti yẹra fún ìgbésí ayé yí wọn ká, bí kò ṣe láti mọ̀ pé "èyí jẹ́ okùnfà" dípò "ohun kan ti bàjẹ́ nínú mi."'),
    ('00000000-0000-0000-0000-000000001102', 'ha', 'Fahimtar Abubuwan Da Ke Tayar Da Martaninka', 'Abin da ke tayarwa (trigger) wani abu ne -- sauti, wari, yanayin murya, wata magana ta musamman -- da ke sanya jikinka ya nuna martani kamar dai tsohon hatsari yana sake faruwa, ko da kana cikin aminci a yanzu. Ba lallai abin tayarwa ya yi ma''ana ta hankali ba kafin ya zama gaskiya; shi tunawar jiki ce ta alaka, ba hujja ce mai tsari ba.\n\nFuskantar abin tayarwa sau da yawa yakan ji kamar ya wuce gwargwado game da abin da kawai ya faru, wanda zai iya rikitarwa ko kunyatarwa. Yana taimaka a tuna cewa girman martanin yana da alaka da ainihin abin da ya faru a baya, ba lallai da lokacin yanzu ba.\n\nAbu mai sauki da za a fara da shi shi ne a lura, bayan wani martani mai karfi, da abubuwa guda uku: me ya faru kafin shi, me ka ji a jikinka, da wace tunani ta zo tare da shi. Da tafiyar lokaci, wannan yana gina taswirar abubuwan tayarwa naka -- ba domin gujewa rayuwa kewaye da su ba, sai dai don gane "wannan abin tayarwa ne" maimakon "wani abu ba daidai ba ne a gare ni."'),
    ('00000000-0000-0000-0000-000000001102', 'ig', 'Ịghọta Ihe Ndị Na-Akpali Gị', 'Ihe na-akpali (trigger) bụ ihe ọ bụla -- ụda, isi, ụdị olu, okwu a kapịrị ọnụ -- nke na-eme ka ahụ gị zaghachi ka à ga-asị na ihe ize ndụ oge gara aga na-emegharị ọzọ, ọbụlagodi mgbe ị nọ n''udo ugbu a. Ihe na-akpali anaghị achọkarị inwe isi ihe tupu ọ bụrụ eziokwu; ọ bụ ncheta jikọrọ ọnụ nke ahụ gị, ọ bụghị arụmụka e ji uche tụlee.\n\nInwe ihe kpaliri gị na-adịkarị ka ihe karịrị ókè gbasara ihe ka naanị mere, nke nwere ike ịgbagwoju anya ma ọ bụ ihere. Ọ na-enyere aka icheta na ike azịza ahụ metụtara ahụmịhe mbụ nke ihe na-akpali ahụ jikọrọ ya, ọ bụghị nke ga-aghaghị ịbụ oge ugbu a.\n\nOmume mfe iji malite bụ ịmara, mgbe azịza siri ike mechara, ihe atọ: ihe merenụ tupu ya, ihe ị nụrụ n''ahụ gị, na echiche sorooro ya. Ka oge na-aga, nke a na-ewulite map nke ihe ndị na-akpali gị -- ọ bụghị izere ndụ gburugburu ha, kama ka ị mata "nke a bụ ihe mkpali" kama "ihe adịghị mma n''ebe m nọ."'),
    ('00000000-0000-0000-0000-000000001102', 'fr', 'Comprendre Vos Déclencheurs', 'Un déclencheur est n''importe quoi -- un son, une odeur, un ton de voix, une phrase particulière -- qui pousse votre système nerveux à réagir comme si une menace passée se reproduisait, même lorsque vous êtes actuellement en sécurité. Les déclencheurs n''ont pas besoin d''être logiques pour être réels ; ils relèvent de la mémoire associative de votre corps, pas d''un raisonnement.\n\nÊtre déclenché semble souvent disproportionné par rapport à ce qui vient de se passer, ce qui peut être déroutant ou gênant. Il est utile de se rappeler que l''intensité de la réaction appartient à l''expérience d''origine à laquelle le déclencheur est lié, pas nécessairement au moment présent.\n\nUne pratique simple pour commencer consiste à remarquer, après une forte réaction, trois choses : ce qui s''est passé juste avant, ce que vous avez ressenti dans votre corps, et quelle pensée l''a accompagnée. Avec le temps, cela construit une carte de vos propres déclencheurs -- non pas pour éviter la vie autour d''eux, mais pour reconnaître « c''est un déclencheur » plutôt que « quelque chose ne va pas chez moi ».'),
    ('00000000-0000-0000-0000-000000001102', 'sw', 'Kuelewa Vichocheo Vyako', 'Kichocheo ni kitu chochote -- sauti, harufu, mtindo wa sauti, msemo maalum -- kinachofanya mfumo wako wa neva uitikie kana kwamba hatari ya zamani inatokea tena, hata unapokuwa salama kwa sasa. Vichocheo havihitaji kuwa na mantiki ili viwe halisi; ni kumbukumbu ya kiunganishi ya mwili wako, si hoja yenye mantiki.\n\nKuchocheka mara nyingi huhisi kuzidi kiwango cha kile kilichotokea, jambo ambalo laweza kuchanganya au kuaibisha. Husaidia kukumbuka kuwa ukali wa mwitikio unahusiana na tukio la asili ambalo kichocheo kimeunganishwa nalo, si lazima na wakati wa sasa.\n\nMazoezi rahisi ya kuanzia ni kutambua, baada ya mwitikio mkali, mambo matatu: nini kilitokea kabla yake, ulihisi nini mwilini mwako, na wazo gani lilikuja nalo. Baada ya muda huu hujenga ramani ya vichocheo vyako binafsi -- si kuepuka maisha yanayovizunguka, bali kutambua "hiki ni kichocheo" badala ya "kuna tatizo na mimi."'),
    ('00000000-0000-0000-0000-000000001102', 'ar', 'فهم المحفزات الخاصة بك', 'المحفز هو أي شيء -- صوت، رائحة، نبرة صوت، عبارة معينة -- يجعل جهازك العصبي يتفاعل كما لو أن تهديدًا من الماضي يحدث مرة أخرى، حتى عندما تكون آمنًا حاليًا. لا يحتاج المحفز إلى أن يكون منطقيًا ليكون حقيقيًا؛ إنه ذاكرة ترابطية لجسدك، وليس حجة مدروسة.\n\nغالبًا ما يبدو الشعور بالتحفّز غير متناسب مع ما حدث للتو، وهو ما قد يكون محيرًا أو محرجًا. من المفيد أن تتذكر أن شدة الاستجابة تعود إلى التجربة الأصلية التي يرتبط بها المحفز، وليس بالضرورة إلى اللحظة الحالية.\n\nممارسة بسيطة للبدء هي ملاحظة ثلاثة أشياء بعد استجابة قوية: ماذا حدث قبلها مباشرة، وماذا شعرت به في جسدك، وما الفكرة التي رافقتها. بمرور الوقت، يبني هذا خريطة لمحفزاتك الخاصة -- ليس لتجنب الحياة من حولها، بل للتعرف على "هذا محفز" بدلًا من "هناك خطأ ما بي".'),

    ('00000000-0000-0000-0000-000000001103', 'pcm', 'How To Build Your Own Awareness Practice', 'Self-awareness about trauma response no be one-time gist -- e dey grow small-small as you dey check yourself well-well instead of just one big "I don see am now." Good habit na to pause small times for day make you ask: wetin I dey feel for my body now, and which state I dey (calm, activated, or shut down)?\n\nJournaling (check dis app Journal tool) and mood tracking na practical ways to build dis awareness as time dey go, because dem dey turn "I just feel off today" wey no clear to specific pattern wey you fit check -- particular time of day, particular person, particular situation.\n\nDis practice no be say make you dey calm forever. Na to make the gap short between when reaction happen and when you recognize am, so that you go get better choice for how you go respond next.'),
    ('00000000-0000-0000-0000-000000001103', 'yo', 'Kíkọ́ Eré Ìmọ̀ Ara-Ẹni Rẹ', 'Ìmọ̀ ara-ẹni nípa ìdáhùn ìpalára-ọkàn kì í ṣe ìmọ̀ ìgbà kan ṣoṣo -- ó máa ń dàgbà nípasẹ̀ ìṣàyẹ̀wò kékèké lemọ́lemọ́ dípò ìmọ̀ ńlá kan ṣoṣo. Àṣà rere ni láti dúró ní ìgbà mélòó kan lójúmọ́ láti bi ara rẹ léèrè: kí ni mo ń nímọ̀lára nínú ara mi báyìí, àti ipò wo ni mo wà (ìdákẹ́jẹ́ẹ́, mímú ṣiṣẹ́, tàbí dídi-dúró)?\n\nKíkọ ìwé-ìrántí (wo ohun èlò Ìwé-Ìrántí app yìí) àti ṣíṣàyẹ̀wò ìmọ̀lára jẹ́ ọ̀nà gidi méjèèjì láti kọ́ ìmọ̀ yìí ní àkókò, nítorí wọ́n máa ń yí "mo nímọ̀lára búburú lónìí" aláìdánimọ̀ padà sí àpẹẹrẹ pàtó tí a lè ṣàyẹ̀wò -- àkókò kan pàtó lọ́jọ́, ẹni kan pàtó, ipò kan pàtó.\n\nÌṣe yìí kì í ṣe láti ní ìdákẹ́jẹ́ẹ́ nígbà gbogbo. Ó jẹ́ láti dín àyè tí ó wà láàrin ìgbà tí ìdáhùn bá ṣẹlẹ̀ àti ìgbà tí o bá mọ̀ ọ́n kù, kí o bàa lè ní àṣàyàn tó dára jù nínú bí o ṣe máa dáhùn lẹ́yìn náà.'),
    ('00000000-0000-0000-0000-000000001103', 'ha', 'Gina Dabi''ar Wayewar Kai', 'Wayewar kai game da martanin rauni ba abu ne na sau daya ba -- yana girma ta hanyar bincike kai karami-karami akai-akai maimakon haske daya babba. Dabi''a mai kyau ita ce ka tsaya sau da yawa a rana ka tambayi kanka: me nake ji a jikina a yanzu, kuma a wane yanayi nake (natsuwa, kunnawa, ko daskarewa)?\n\nRubutun diary (duba kayan aikin Diary na wannan app) da kula da yanayin tunani hanyoyi ne masu amfani guda biyu don ginawa wayewar nan cikin lokaci, domin suna mayar da "ina jin ba daidai ba yau" marar bayyana zuwa tsari takamaimai da za ka iya duba -- wani lokaci na rana, wani mutum, wani yanayi.\n\nWannan dabi''a ba domin samun natsuwa har abada ba ne. Ita ce don rage gibi tsakanin lokacin da martani ke faruwa da lokacin da ka gane shi, domin ka sami zabi mafi kyau a yadda za ka mayar da martani na gaba.'),
    ('00000000-0000-0000-0000-000000001103', 'ig', 'Ịrụ Omume Ịmara Onwe Gị', 'Ịmara onwe gị gbasara azịza mmerụ ahụ-mmụọ abụghị ihe a na-eme otu ugbo -- ọ na-eto site na nyocha nta mgbe niile kama otu nghọta buru ibu. Omume dị mma bụ ikwụsị mgbe ole na ole kwa ụbọchị ka ị jụọ onwe gị: gịnị ka m na-enwe mmetụta n''ahụ m ugbu a, nakwa kedu ọnọdụ m nọ (jụụ, na-arụ ọrụ, ma ọ bụ mkpuchi)?\n\nIde akwụkwọ ndekọ (lelee ngwá ọrụ Akwụkwọ Ndekọ nke app a) na nleba anya mmetụta uche bụ ụzọ abụọ bara uru iji wulite nghọta a n''oge gara aga, n''ihi na ha na-eme ka "ọ dị m ka ihe adịghị mma taa" nke na-edoghị anya ghọọ ụdị a kapịrị ọnụ nke ị nwere ike inyocha -- oge akọwapụtara nke ụbọchị, onye akọwapụtara, ọnọdụ akọwapụtara.\n\nOmume a abụghị ka ị nwee jụụ mgbe niile. Ọ bụ ka oghere dị n''etiti mgbe azịza na-apụta na mgbe ị matara ya dị nta, ka ị nwee nhọrọ ka mma n''otu i si azaghachi ihe ọzọ.'),
    ('00000000-0000-0000-0000-000000001103', 'fr', 'Construire Votre Propre Pratique de Conscience de Soi', 'La conscience de soi face aux réactions au traumatisme n''est pas une prise de conscience ponctuelle -- elle se développe par de petites vérifications répétées plutôt que par une seule révélation majeure. Une bonne habitude consiste à faire une pause plusieurs fois par jour pour se demander : que ressens-je dans mon corps en ce moment, et dans quel état suis-je (calme, activé, ou figé) ?\n\nTenir un journal (voir l''outil Journal de cette application) et suivre son humeur sont deux moyens concrets de développer cette conscience au fil du temps, car ils transforment un vague « je me sens mal aujourd''hui » en schémas précis et observables -- certains moments de la journée, certaines personnes, certaines situations.\n\nCette pratique ne vise pas un calme permanent. Il s''agit de raccourcir l''écart entre le moment où une réaction se produit et celui où vous la reconnaissez, afin d''avoir davantage de choix dans la façon dont vous réagirez ensuite.'),
    ('00000000-0000-0000-0000-000000001103', 'sw', 'Kujenga Mazoezi Yako Binafsi ya Kujitambua', 'Kujitambua kuhusu miitikio ya kiwewe si ufahamu wa mara moja -- hukua kupitia uangalizi mdogo mdogo wa mara kwa mara badala ya mwamko mkubwa mmoja. Tabia nzuri ni kusimama mara kadhaa kwa siku kujiuliza: ninahisi nini mwilini mwangu sasa hivi, na niko katika hali gani (utulivu, kuchochewa, au kuganda)?\n\nKuandika shajara (angalia kifaa cha Shajara cha programu hii) na kufuatilia hisia ni njia mbili za vitendo za kujenga ufahamu huu kwa muda, kwa sababu hubadilisha "nahisi vibaya leo" isiyo wazi kuwa mifumo mahususi unayoweza kuikagua -- nyakati fulani za siku, watu fulani, hali fulani.\n\nMazoezi haya si kuhusu kufikia utulivu wa kudumu. Ni kuhusu kupunguza pengo kati ya wakati mwitikio unapotokea na wakati unapoutambua, ili upate hiari zaidi katika jinsi utakavyoitikia baadaye.'),
    ('00000000-0000-0000-0000-000000001103', 'ar', 'بناء ممارسة الوعي الذاتي الخاصة بك', 'الوعي الذاتي تجاه استجابات الصدمة ليس إدراكًا لمرة واحدة -- بل ينمو من خلال فحوصات صغيرة متكررة بدلًا من اكتشاف كبير واحد. من العادات الجيدة التوقف عدة مرات في اليوم لسؤال نفسك: بماذا أشعر في جسدي الآن، وفي أي حالة أنا (هادئ، متحفز، أو متجمد)؟\n\nكتابة اليوميات (راجع أداة اليوميات في هذا التطبيق) وتتبع المزاج طريقتان عمليتان لبناء هذا الوعي بمرور الوقت، لأنهما تحوّلان شعور "أشعر بالسوء اليوم" الغامض إلى أنماط محددة يمكنك مراجعتها -- وقت معين من اليوم، شخص معين، موقف معين.\n\nهذه الممارسة ليست عن تحقيق هدوء دائم. إنها عن تقصير الفجوة بين لحظة حدوث الاستجابة ولحظة إدراكها، حتى يكون لديك خيار أفضل في كيفية الاستجابة لاحقًا.');

-- ============================================================
-- Quiz questions (3 questions x 7 locales = 21 rows)
-- ============================================================
INSERT INTO quiz_question_translations (question_id, locale, text) VALUES
    ('00000000-0000-0000-0000-000000001301', 'pcm', 'Apart from fight and flight, wetin be the other two common wahala responses?'),
    ('00000000-0000-0000-0000-000000001301', 'yo', 'Yàtọ̀ sí ìjà àti ìsá, kí ni àwọn ìdáhùn méjì mìíràn tí ó wọ́pọ̀?'),
    ('00000000-0000-0000-0000-000000001301', 'ha', 'Baya ga yaki da gudu, mene ne sauran halayen rauni guda biyu da aka saba gani?'),
    ('00000000-0000-0000-0000-000000001301', 'ig', 'Ewezuga ọgụ na ịgbapụ, gịnị bụ azịza mmerụ ahụ-mmụọ ndị ọzọ a na-ahụkarị abụọ?'),
    ('00000000-0000-0000-0000-000000001301', 'fr', 'Outre le combat et la fuite, quelles sont les deux autres réactions courantes au traumatisme ?'),
    ('00000000-0000-0000-0000-000000001301', 'sw', 'Mbali na kupigana na kukimbia, ni ipi miitikio mingine miwili ya kawaida ya kiwewe?'),
    ('00000000-0000-0000-0000-000000001301', 'ar', 'بصرف النظر عن القتال والهروب، ما هما الاستجابتان الشائعتان الأخريان للصدمة؟'),

    ('00000000-0000-0000-0000-000000001302', 'pcm', 'Wetin dey determine how strong trigger reaction go be?'),
    ('00000000-0000-0000-0000-000000001302', 'yo', 'Kí ni ó ń pinnu bí ìdáhùn sí ohun tí ó fa ìdáhùn ṣe lágbára tó?'),
    ('00000000-0000-0000-0000-000000001302', 'ha', 'Mene ne ke tantance girman martanin da aka tayar?'),
    ('00000000-0000-0000-0000-000000001302', 'ig', 'Gịnị na-ekpebi ike nke azịza e ji kpalie?'),
    ('00000000-0000-0000-0000-000000001302', 'fr', 'Qu''est-ce qui détermine l''intensité d''une réaction déclenchée ?'),
    ('00000000-0000-0000-0000-000000001302', 'sw', 'Ni nini kinachoamua ukali wa mwitikio uliochochewa?'),
    ('00000000-0000-0000-0000-000000001302', 'ar', 'ما الذي يحدد شدة الاستجابة المحفَّزة؟'),

    ('00000000-0000-0000-0000-000000001303', 'pcm', 'Wetin be the goal of personal awareness practice, according to dis lesson?'),
    ('00000000-0000-0000-0000-000000001303', 'yo', 'Kí ni ète eré ìmọ̀ ara-ẹni, gẹ́gẹ́ bí ẹ̀kọ́ yìí ṣe sọ?'),
    ('00000000-0000-0000-0000-000000001303', 'ha', 'Mece ce manufar dabi''ar wayewar kai, bisa ga wannan darasi?'),
    ('00000000-0000-0000-0000-000000001303', 'ig', 'Gịnị bụ ebumnuche omume ịmara onwe gị, dịka nkuzi a kwuru?'),
    ('00000000-0000-0000-0000-000000001303', 'fr', 'Quel est l''objectif d''une pratique de conscience de soi personnelle, selon cette leçon ?'),
    ('00000000-0000-0000-0000-000000001303', 'sw', 'Lengo la mazoezi ya kujitambua binafsi ni nini, kulingana na somo hili?'),
    ('00000000-0000-0000-0000-000000001303', 'ar', 'ما هو الهدف من ممارسة الوعي الذاتي الشخصية، وفقًا لهذا الدرس؟');

-- ============================================================
-- Quiz options (12 options x 7 locales = 84 rows)
-- ============================================================
INSERT INTO quiz_option_translations (option_id, locale, text) VALUES
    ('00000000-0000-0000-0000-000000001401', 'pcm', 'Freeze and fawn (dey try please people)'),
    ('00000000-0000-0000-0000-000000001401', 'yo', 'Dídi-dúró àti ìtẹríba'),
    ('00000000-0000-0000-0000-000000001401', 'ha', 'Tsayawa da sulhu'),
    ('00000000-0000-0000-0000-000000001401', 'ig', 'Ịkwụsị na nrube isi'),
    ('00000000-0000-0000-0000-000000001401', 'fr', 'La sidération et la soumission'),
    ('00000000-0000-0000-0000-000000001401', 'sw', 'Kuganda na kuridhia'),
    ('00000000-0000-0000-0000-000000001401', 'ar', 'التجمد والاسترضاء'),

    ('00000000-0000-0000-0000-000000001402', 'pcm', 'Panic and rest'),
    ('00000000-0000-0000-0000-000000001402', 'yo', 'Ìyàgbẹ́ àti ìsinmi'),
    ('00000000-0000-0000-0000-000000001402', 'ha', 'Firgita da hutawa'),
    ('00000000-0000-0000-0000-000000001402', 'ig', 'Ụjọ na izu ike'),
    ('00000000-0000-0000-0000-000000001402', 'fr', 'La panique et le repos'),
    ('00000000-0000-0000-0000-000000001402', 'sw', 'Hofu na kupumzika'),
    ('00000000-0000-0000-0000-000000001402', 'ar', 'الذعر والراحة'),

    ('00000000-0000-0000-0000-000000001403', 'pcm', 'Avoidance and anger'),
    ('00000000-0000-0000-0000-000000001403', 'yo', 'Yíyẹra àti ìbínú'),
    ('00000000-0000-0000-0000-000000001403', 'ha', 'Gujewa da fushi'),
    ('00000000-0000-0000-0000-000000001403', 'ig', 'Izere na iwe'),
    ('00000000-0000-0000-0000-000000001403', 'fr', 'L''évitement et la colère'),
    ('00000000-0000-0000-0000-000000001403', 'sw', 'Kuepuka na hasira'),
    ('00000000-0000-0000-0000-000000001403', 'ar', 'التجنب والغضب'),

    ('00000000-0000-0000-0000-000000001404', 'pcm', 'Sleep and appetite change'),
    ('00000000-0000-0000-0000-000000001404', 'yo', 'Oorun àti àyípadà ìfẹ́-oúnjẹ'),
    ('00000000-0000-0000-0000-000000001404', 'ha', 'Barci da canjin sha''awar abinci'),
    ('00000000-0000-0000-0000-000000001404', 'ig', 'Ụra na mgbanwe agụụ nri'),
    ('00000000-0000-0000-0000-000000001404', 'fr', 'Le sommeil et les changements d''appétit'),
    ('00000000-0000-0000-0000-000000001404', 'sw', 'Usingizi na mabadiliko ya hamu ya kula'),
    ('00000000-0000-0000-0000-000000001404', 'ar', 'النوم وتغير الشهية'),

    ('00000000-0000-0000-0000-000000001411', 'pcm', 'How well the reaction take dey acceptable to people'),
    ('00000000-0000-0000-0000-000000001411', 'yo', 'Bí ìdáhùn náà ṣe ṣe ìtẹ́wọ́gbà láwùjọ'),
    ('00000000-0000-0000-0000-000000001411', 'ha', 'Yadda martanin yake da karbuwa ga jama''a'),
    ('00000000-0000-0000-0000-000000001411', 'ig', 'Otú azịza ahụ si akwado nke ọma n''anya mmadụ'),
    ('00000000-0000-0000-0000-000000001411', 'fr', 'À quel point la réaction paraît socialement acceptable'),
    ('00000000-0000-0000-0000-000000001411', 'sw', 'Jinsi mwitikio unavyokubalika kijamii'),
    ('00000000-0000-0000-0000-000000001411', 'ar', 'مدى قبول الاستجابة اجتماعيًا'),

    ('00000000-0000-0000-0000-000000001412', 'pcm', 'The first experience wey the trigger connect to'),
    ('00000000-0000-0000-0000-000000001412', 'yo', 'Ìrírí ìpilẹ̀ṣẹ̀ tí ohun tí ń fa ìdáhùn so pọ̀ mọ́'),
    ('00000000-0000-0000-0000-000000001412', 'ha', 'Ainihin abin da ya faru da abin tayarwar ya danganta da shi'),
    ('00000000-0000-0000-0000-000000001412', 'ig', 'Ahụmịhe mbụ nke ihe mkpali ahụ jikọrọ ya'),
    ('00000000-0000-0000-0000-000000001412', 'fr', 'L''expérience d''origine à laquelle le déclencheur est lié'),
    ('00000000-0000-0000-0000-000000001412', 'sw', 'Tukio la asili ambalo kichocheo kimeunganishwa nalo'),
    ('00000000-0000-0000-0000-000000001412', 'ar', 'التجربة الأصلية التي يرتبط بها المحفز'),

    ('00000000-0000-0000-0000-000000001413', 'pcm', 'How much sleep you get the night before'),
    ('00000000-0000-0000-0000-000000001413', 'yo', 'Iye oorun tí o sùn ní alẹ́ ṣáájú'),
    ('00000000-0000-0000-0000-000000001413', 'ha', 'Yawan barcin da ka samu a daren da ya gabata'),
    ('00000000-0000-0000-0000-000000001413', 'ig', 'Ole ụra ị hiri n''abalị gara aga'),
    ('00000000-0000-0000-0000-000000001413', 'fr', 'La quantité de sommeil obtenue la nuit précédente'),
    ('00000000-0000-0000-0000-000000001413', 'sw', 'Kiasi cha usingizi ulichopata usiku uliopita'),
    ('00000000-0000-0000-0000-000000001413', 'ar', 'مقدار النوم الذي حصلت عليه الليلة السابقة'),

    ('00000000-0000-0000-0000-000000001414', 'pcm', 'Nothing -- the reaction just dey random'),
    ('00000000-0000-0000-0000-000000001414', 'yo', 'Kò sí ohunkóhun -- ìdáhùn náà máa ń jẹ́ àrìnàkò ni'),
    ('00000000-0000-0000-0000-000000001414', 'ha', 'Babu komai -- martanin bazuwar ne kawai'),
    ('00000000-0000-0000-0000-000000001414', 'ig', 'Ọ dịghị ihe -- azịza ahụ bụ nanị nke mberede'),
    ('00000000-0000-0000-0000-000000001414', 'fr', 'Rien -- la réaction est simplement aléatoire'),
    ('00000000-0000-0000-0000-000000001414', 'sw', 'Hakuna -- mwitikio ni wa nasibu tu'),
    ('00000000-0000-0000-0000-000000001414', 'ar', 'لا شيء -- الاستجابة عشوائية فقط'),

    ('00000000-0000-0000-0000-000000001421', 'pcm', 'To dey always calm'),
    ('00000000-0000-0000-0000-000000001421', 'yo', 'Láti máa wà ní ìdákẹ́jẹ́ẹ́ nígbà gbogbo'),
    ('00000000-0000-0000-0000-000000001421', 'ha', 'Samun natsuwa koyaushe'),
    ('00000000-0000-0000-0000-000000001421', 'ig', 'Ịnọgide na jụụ mgbe niile'),
    ('00000000-0000-0000-0000-000000001421', 'fr', 'Atteindre un calme permanent'),
    ('00000000-0000-0000-0000-000000001421', 'sw', 'Kuwa na utulivu wa kudumu'),
    ('00000000-0000-0000-0000-000000001421', 'ar', 'تحقيق هدوء دائم'),

    ('00000000-0000-0000-0000-000000001422', 'pcm', 'To make the gap short between reaction and when you recognize am'),
    ('00000000-0000-0000-0000-000000001422', 'yo', 'Láti dín àyè kù láàrin ìdáhùn àti ìgbà tí a mọ̀ ọ́n'),
    ('00000000-0000-0000-0000-000000001422', 'ha', 'Rage gibi tsakanin martani da lokacin da ka gane shi'),
    ('00000000-0000-0000-0000-000000001422', 'ig', 'Ibelata oghere dị n''etiti azịza na mgbe amatara ya'),
    ('00000000-0000-0000-0000-000000001422', 'fr', 'Raccourcir l''écart entre une réaction et le moment où on la reconnaît'),
    ('00000000-0000-0000-0000-000000001422', 'sw', 'Kupunguza pengo kati ya mwitikio na wakati unapoutambua'),
    ('00000000-0000-0000-0000-000000001422', 'ar', 'تقصير الفجوة بين الاستجابة ولحظة إدراكها'),

    ('00000000-0000-0000-0000-000000001423', 'pcm', 'To avoid all situation wey fit cause reaction'),
    ('00000000-0000-0000-0000-000000001423', 'yo', 'Láti yẹra fún gbogbo ipò tí ó lè fa ìdáhùn'),
    ('00000000-0000-0000-0000-000000001423', 'ha', 'Guje wa dukkan yanayi da zai iya haifar da martani'),
    ('00000000-0000-0000-0000-000000001423', 'ig', 'Izere ọnọdụ niile nwere ike ịkpata azịza'),
    ('00000000-0000-0000-0000-000000001423', 'fr', 'Éviter toute situation pouvant provoquer une réaction'),
    ('00000000-0000-0000-0000-000000001423', 'sw', 'Kuepuka hali zote zinazoweza kusababisha mwitikio'),
    ('00000000-0000-0000-0000-000000001423', 'ar', 'تجنب كل المواقف التي قد تسبب استجابة'),

    ('00000000-0000-0000-0000-000000001424', 'pcm', 'To prove to people say you no dey over-react'),
    ('00000000-0000-0000-0000-000000001424', 'yo', 'Láti fi hàn fún àwọn ẹlòmíràn pé o kò ṣe àṣejù'),
    ('00000000-0000-0000-0000-000000001424', 'ha', 'Tabbatarwa wasu cewa ba ka wuce gona da iri ba'),
    ('00000000-0000-0000-0000-000000001424', 'ig', 'Igosi ndị ọzọ na ị naghị eme oke azịza'),
    ('00000000-0000-0000-0000-000000001424', 'fr', 'Prouver aux autres que l''on ne réagit pas de façon excessive'),
    ('00000000-0000-0000-0000-000000001424', 'sw', 'Kuwathibitishia wengine kuwa hukuzidisha mwitikio'),
    ('00000000-0000-0000-0000-000000001424', 'ar', 'إثبات للآخرين أنك لا تبالغ في رد الفعل');
