// Analyst annotations only; original export: source/ida/0x180165140.c
// qword comments derive module constant slot + original file range; not runtime values.
// write access to const memory has been detected, the output may be wrong!
int *__fastcall sub_180165140(__int64 a1, __int64 a2, int **a3)
{
  __int64 v4; // r12
  int *v5; // rbx
  _QWORD *v6; // r13
  _QWORD *v7; // rdi
  int *v8; // rsi
  bool v9; // zf
  unsigned int v10; // r14d
  __int64 v11; // rax
  int *v12; // rbx
  int v13; // edi
  __int64 v14; // rax
  int *v15; // rbx
  __int64 v16; // rdi
  int *v17; // rbx
  __int64 v18; // rax
  int *v19; // rdi
  int *v20; // rsi
  __int64 v21; // rax
  bool v22; // sf
  int *v23; // r13
  __int64 v24; // rcx
  int v25; // eax
  int *v26; // rsi
  __int64 v27; // rax
  int *v28; // rax
  int *v29; // rbx
  __int64 v30; // rax
  int *v31; // rax
  int *v32; // rdx
  __int64 v33; // rax
  int *v34; // rdx
  __int64 v35; // rax
  int *v36; // rbx
  int v37; // edi
  __int64 v38; // rax
  __int64 v39; // rax
  __int64 v40; // rax
  int *v41; // rbx
  __int64 v42; // rdi
  __int64 v43; // rax
  int *v44; // rax
  int *v45; // rdi
  __int64 v46; // rax
  int *v47; // rbx
  int v48; // eax
  __int64 v49; // rdx
  int *v50; // r12
  int *v51; // rax
  int *v52; // rbx
  __int64 v53; // rax
  __int64 v54; // rax
  __int64 v55; // r13
  __int64 v56; // rax
  int *v57; // rcx
  __int64 v58; // rax
  int *v59; // r12
  __int64 v60; // rdi
  __int64 v61; // rbx
  __int64 v62; // rsi
  int *v63; // r13
  __int64 v64; // rdi
  __int64 v65; // rbx
  int *v66; // rsi
  __int64 v67; // rdi
  __int64 v68; // rax
  int *v69; // r8
  __int64 v70; // rax
  int *v71; // rbx
  __int64 v72; // rax
  __int64 v73; // rax
  __int64 v74; // rax
  __int64 v75; // rax
  __int64 v76; // rdi
  __int64 v77; // rax
  __int64 v78; // rax
  __int64 v79; // rax
  int *v80; // rcx
  __int64 v81; // rsi
  int *v82; // rbx
  __int64 v83; // rax
  __int64 v84; // rax
  __int64 v85; // rax
  _QWORD *v86; // rcx
  __int64 v87; // rax
  int *v88; // rdi
  __int64 v89; // rax
  __int64 v90; // rax
  __int64 v91; // rax
  __int64 v92; // rcx
  __int64 v93; // rdx
  int *v94; // rbx
  __int64 v95; // rax
  __int64 v96; // r12
  __int64 v97; // rax
  __int64 v98; // rax
  int v99; // eax
  __int64 v100; // rax
  int *v101; // rbx
  __int64 v102; // rax
  int *v103; // rsi
  __int64 v104; // rax
  int *v105; // rdi
  __int64 v106; // rax
  __int64 v107; // rax
  __int64 v108; // rax
  __int64 v109; // rax
  __int64 v110; // rax
  __int64 v111; // r13
  __int64 v112; // rax
  int *v113; // rsi
  __int64 v114; // rax
  __int64 v115; // rax
  int *v116; // r12
  __int64 v117; // rax
  __int64 v118; // rax
  __int64 v119; // rax
  __int64 v120; // rcx
  __int64 v121; // rax
  __int64 v122; // rdi
  __int64 v123; // rax
  __int64 v124; // rax
  int *v125; // rbx
  int *v126; // rdi
  __int64 v127; // rax
  __int64 v128; // rax
  int *v129; // rsi
  __int64 v130; // rax
  int *v131; // rcx
  int *v132; // rdx
  int *v133; // r8
  int *v134; // rbx
  int *v135; // rdi
  __int64 v136; // rax
  int v137; // eax
  int *v138; // rbx
  int v139; // eax
  __int64 v140; // rax
  int v141; // eax
  int *v142; // rsi
  _DWORD *v143; // rdi
  __int64 v144; // rax
  int *v145; // rbx
  __int64 v146; // r13
  __int64 v147; // rax
  int *v148; // rdi
  int *v149; // r12
  __int64 v150; // rax
  __int64 v151; // rdi
  __int64 v152; // rax
  __int64 v153; // r8
  __int64 v154; // rax
  __int64 v155; // r8
  int *v156; // rdi
  __int64 v157; // rax
  _QWORD *v158; // rax
  int *v159; // rbx
  __int64 v160; // rax
  __int64 v161; // rax
  __int64 v162; // rax
  __int64 v163; // rdi
  _QWORD *v164; // rsi
  __int64 v165; // rax
  _QWORD *v166; // r8
  int *v167; // rbx
  int *v168; // rax
  __int64 v169; // rax
  __int64 v170; // rax
  int *v171; // rbx
  __int64 v172; // rax
  int *v173; // r12
  __int64 v174; // rdi
  __int64 v175; // rbx
  __int64 v176; // rsi
  int *v177; // r13
  __int64 v178; // rdi
  __int64 v179; // rbx
  int *v180; // rsi
  __int64 v181; // rdi
  __int64 v182; // rax
  int *v183; // r8
  __int64 v184; // rax
  int *v185; // rbx
  __int64 v186; // rax
  __int64 v187; // rsi
  __int64 v188; // rdi
  int *v189; // rbx
  __int64 v190; // rcx
  int *v191; // rdi
  __int64 v192; // rax
  __int64 v193; // rax
  __int64 v194; // rsi
  __int64 v195; // rax
  __int64 v196; // rax
  __int64 v197; // rax
  int *v198; // rbx
  __int64 v199; // rax
  __int64 v200; // rax
  int *v201; // rbx
  __int64 v202; // rcx
  int *v203; // rdi
  __int64 v204; // rax
  int *v205; // rbx
  __int64 v206; // rax
  __int64 v207; // rcx
  __int64 v208; // rdi
  __int64 v209; // rax
  int *v210; // rbx
  __int64 v211; // r8
  __int64 v212; // rax
  int *v213; // rdi
  int *v214; // rsi
  __int64 v215; // rax
  __int64 v216; // rdi
  __int64 v217; // rax
  __int64 v218; // rdi
  __int64 v219; // rax
  int *v220; // rbx
  __int64 v221; // rax
  __int64 v222; // rax
  int *v223; // rcx
  __int64 v224; // rax
  __int64 v225; // rdi
  _QWORD *v226; // rbx
  __int64 v227; // rax
  __int64 v228; // rax
  __int64 v229; // rax
  _QWORD *v230; // rbx
  __int64 v231; // rax
  __int64 v232; // rax
  __int64 v233; // rax
  __int64 v234; // rax
  __int64 v235; // rax
  __int64 v236; // rax
  __int64 v237; // rax
  __int64 v238; // rbx
  __int64 v239; // rax
  _QWORD *v240; // rdi
  __int64 v241; // rax
  __int64 v242; // rax
  __int64 v243; // rax
  __int64 v244; // rax
  int *v245; // rbx
  __int64 v246; // rax
  __int64 v247; // rax
  __int64 v248; // rdx
  __int64 v249; // r8
  int *v250; // rax
  int *v251; // rbx
  __int64 v252; // rax
  __int64 v253; // rax
  int *v254; // rbx
  __int64 v255; // rax
  int *v256; // rsi
  __int64 v257; // rax
  int *v258; // rdi
  __int64 v259; // rax
  __int64 v260; // rax
  __int64 v261; // rax
  __int64 v262; // rax
  __int64 v263; // rax
  __int64 v264; // rax
  int *v265; // rsi
  __int64 v266; // rax
  __int64 v267; // rax
  int *v268; // r12
  __int64 v269; // rax
  __int64 v270; // rax
  __int64 v271; // rax
  __int64 v272; // rax
  int *v273; // rdi
  int *v274; // rdx
  int *v275; // rax
  int *v276; // rbx
  __int64 v277; // rax
  int *v278; // rax
  int *v279; // rbx
  __int64 v280; // rax
  __int64 v281; // rax
  __int64 v282; // rax
  __int64 v283; // rbx
  __int64 v284; // rax
  int *v285; // r8
  _QWORD *v286; // rcx
  int *v287; // rdx
  __int64 v288; // rax
  int *v289; // r12
  _DWORD *v290; // rbx
  __int64 v291; // rax
  int *v292; // rsi
  __int64 v293; // rdi
  __int64 v294; // rbx
  int *v295; // r13
  __int64 v296; // rdi
  __int64 v297; // rax
  __int64 v298; // rax
  int *v299; // rbx
  __int64 v300; // rax
  int *v301; // rcx
  __int64 v302; // rax
  __int64 v303; // rdi
  __int64 v304; // rbx
  int *v305; // r13
  __int64 v306; // rdi
  __int64 v307; // rax
  _QWORD *v308; // r8
  __int64 v309; // rax
  int *v310; // rbx
  __int64 v311; // rax
  int *v312; // rbx
  __int64 v313; // rax
  __int64 v314; // rax
  __int64 v315; // rax
  int *v316; // rbx
  __int64 v317; // rax
  int *v318; // r13
  __int64 v319; // rax
  __int64 v320; // rax
  __int64 v321; // rax
  __int64 v322; // rax
  int *v323; // rbx
  __int64 v324; // rax
  int *v325; // rdi
  __int64 v326; // rcx
  __int64 v327; // rax
  int *v328; // rsi
  __int64 v329; // rax
  int *v330; // rdi
  __int64 v331; // rax
  __int64 v332; // rax
  __int64 v333; // rax
  int *v334; // rdi
  int *v335; // rax
  int *v336; // rbx
  __int64 v337; // rax
  int *v338; // rax
  int *v339; // rbx
  __int64 v340; // rax
  int *v341; // rcx
  int *v342; // r13
  __int64 v343; // rax
  int *v344; // rcx
  __int64 v345; // rax
  __int64 v346; // rcx
  int *v347; // r14
  unsigned int v348; // ebx
  __int64 v349; // rdi
  __int64 v350; // rbx
  __int64 v351; // rsi
  int *v352; // r12
  __int64 v353; // rdi
  __int64 v354; // rbx
  int *v355; // rsi
  __int64 v356; // r13
  __int64 v357; // rax
  int *v358; // r8
  __int64 v359; // rax
  int *v360; // rdi
  __int64 v361; // rax
  __int64 v362; // rdi
  __int64 v363; // rbx
  int *v364; // r13
  __int64 v365; // rax
  int *v366; // rbx
  int *v367; // rsi
  __int64 v368; // rax
  int *v369; // rdi
  __int64 v370; // rax
  int v371; // eax
  __int64 v372; // rax
  __int64 v373; // rax
  int *v374; // rcx
  __int64 v375; // rax
  int *v376; // rcx
  __int64 v377; // rax
  __int64 v378; // rdi
  __int64 v379; // rbx
  __int64 v380; // r13
  __int64 v381; // rax
  int *v382; // r8
  __int64 v383; // rax
  int *v384; // rdi
  __int64 v385; // rax
  __int64 v386; // rax
  __int64 v387; // rax
  __int64 v388; // rsi
  __int64 v389; // rax
  __int64 v390; // rax
  __int64 v391; // rax
  int *v392; // rcx
  __int64 v393; // r13
  int *v394; // rdi
  __int64 v395; // rax
  __int64 v396; // rcx
  __int64 v397; // rsi
  __int64 v398; // rax
  __int64 v399; // rax
  __int64 v400; // rax
  __int64 v401; // rcx
  __int64 v402; // rdx
  int *v403; // rdi
  __int64 v404; // rax
  _DWORD *v405; // r13
  __int64 v406; // rdx
  int *v407; // rax
  int *v408; // rdi
  __int64 v409; // rax
  __int64 v410; // rax
  int *v411; // rdi
  int *v412; // r14
  __int64 v413; // rax
  int *v414; // rbx
  __int64 v415; // rcx
  __int64 v416; // rax
  int *v417; // rsi
  __int64 v418; // rax
  __int64 v419; // rax
  __int64 v420; // rax
  __int64 v421; // rax
  __int64 v422; // rax
  int *v423; // r14
  __int64 v424; // rax
  __int64 v425; // rax
  int *v426; // r12
  __int64 v427; // rax
  __int64 v428; // rax
  __int64 v429; // rax
  __int64 v430; // rax
  __int64 v431; // rax
  __int64 v432; // rsi
  _DWORD *v433; // rdi
  __int64 v434; // rax
  int *v435; // rbx
  int v436; // edi
  __int64 v437; // rax
  _DWORD *v438; // r12
  __int64 v439; // rax
  __int64 v440; // rax
  __int64 v441; // rax
  int *v442; // rbx
  __int64 v443; // rax
  int *v444; // rsi
  __int64 v445; // rcx
  __int64 v446; // rax
  int *v447; // rdi
  __int64 v448; // rax
  __int64 v449; // r13
  _DWORD *v450; // r14
  __int64 v451; // rax
  int *v452; // rsi
  __int64 v453; // r8
  __int64 v454; // r8
  __int64 v455; // r8
  __int64 v456; // r8
  __int64 v457; // r8
  _DWORD *v458; // rcx
  int *v459; // r13
  __int64 v460; // rax
  __int64 v461; // rax
  __int64 v462; // rax
  int *v463; // rsi
  __int64 v464; // rax
  __int64 v465; // rax
  __int64 v466; // rax
  __int64 v467; // rax
  __int64 v468; // rdx
  __int64 v469; // rax
  int *v470; // rax
  int *v471; // rbx
  __int64 v472; // rbx
  __int64 v473; // rax
  __int64 v474; // rax
  __int64 v475; // rax
  int *v476; // rsi
  __int64 v477; // rax
  __int64 v478; // rbx
  __int64 v479; // rax
  int v480; // eax
  int v481; // eax
  __int64 v482; // rax
  int *v483; // rdi
  int *v484; // rdx
  int *v485; // rax
  int *v486; // rax
  int *v487; // rbx
  __int64 v488; // rax
  int *v489; // rax
  int *v490; // rbx
  __int64 v491; // rax
  __int64 v492; // rax
  __int64 v493; // rax
  __int64 v494; // r12
  __int64 v495; // rax
  int *v496; // rbx
  __int64 v497; // rax
  int *v498; // rdi
  __int64 v499; // rcx
  __int64 v500; // rax
  int *v501; // rsi
  __int64 v502; // rax
  int *v503; // rdi
  __int64 v504; // rax
  __int64 v505; // rax
  __int64 v506; // rax
  __int64 v507; // rdi
  int *v508; // rbx
  int *v509; // rax
  __int64 v510; // rax
  __int64 v511; // rax
  int *v512; // r12
  _DWORD *v513; // rbx
  __int64 v514; // rax
  int *v515; // r13
  __int64 v516; // rdi
  __int64 v517; // rbx
  int *v518; // rsi
  __int64 v519; // rdi
  __int64 v520; // rax
  __int64 v521; // rax
  int *v522; // rbx
  __int64 v523; // rax
  int *v524; // rcx
  __int64 v525; // rax
  __int64 v526; // rdi
  __int64 v527; // rbx
  int *v528; // rsi
  __int64 v529; // rdi
  __int64 v530; // rax
  _DWORD *v531; // r8
  __int64 v532; // rax
  int *v533; // rbx
  __int64 v534; // rax
  int *v535; // rbx
  __int64 v536; // rax
  __int64 v537; // rax
  int *v538; // rbx
  __int64 v539; // rax
  int *v540; // rsi
  __int64 v541; // rax
  __int64 v542; // rax
  __int64 v543; // rax
  __int64 v544; // rdi
  __int64 v545; // rax
  int *v546; // rbx
  __int64 v547; // rax
  int *v548; // r13
  __int64 v549; // rax
  __int64 v550; // rax
  int *v551; // rdi
  __int64 v552; // rax
  int *v553; // rbx
  __int64 v554; // rax
  int *v555; // r12
  __int64 v556; // rax
  __int64 v557; // rax
  _DWORD *v558; // rsi
  __int64 v559; // rax
  int *v560; // rbx
  __int64 v561; // rax
  int *v562; // rsi
  __int64 v563; // r14
  __int64 v564; // rax
  __int64 v565; // rax
  __int64 v566; // rax
  __int64 v567; // rax
  __int64 v568; // rsi
  __int64 v569; // rax
  __int64 v570; // rax
  __int64 v571; // rax
  __int64 v572; // rbx
  __int64 v573; // r8
  __int64 v574; // r8
  __int64 v575; // r8
  __int64 v576; // rsi
  __int64 v577; // rbx
  __int64 v578; // r8
  __int64 v579; // r8
  __int64 v580; // r8
  int *v581; // rsi
  __int64 v582; // rax
  int *v583; // rbx
  __int64 v584; // rbx
  int *v585; // r13
  __int64 v586; // rax
  _DWORD *v587; // rbx
  int *v588; // r14
  int *v589; // rdx
  int *v590; // rcx
  int *v591; // rbx
  int *v592; // rsi
  __int64 v593; // rax
  int v594; // eax
  int *v595; // rbx
  int v596; // eax
  __int64 v597; // rax
  _DWORD *v598; // rsi
  __int64 v599; // rax
  int *v600; // rbx
  __int64 v601; // rax
  __int64 v602; // rsi
  __int64 v603; // rax
  char v604; // al
  __int64 v605; // rcx
  __int64 v606; // rax
  __int64 v607; // rax
  int *v608; // r14
  __int64 v609; // rax
  __int64 v610; // rax
  __int64 v611; // rax
  __int64 v612; // r8
  __int64 v613; // r8
  __int64 v614; // r8
  __int64 v615; // rsi
  __int64 v616; // rax
  __int64 v617; // rbx
  __int64 v618; // rax
  int *v619; // rbx
  __int64 v620; // rcx
  __int64 v621; // rax
  __int64 v622; // rax
  int *v623; // rsi
  __int64 v624; // rax
  int *v625; // rbx
  __int64 v626; // rax
  __int64 v627; // rax
  __int64 v628; // rax
  __int64 v629; // rax
  __int64 v630; // rax
  int *v631; // rbx
  __int64 v632; // rax
  int *v633; // rsi
  __int64 v634; // rcx
  __int64 v635; // rax
  int *v636; // rcx
  int *v637; // rdi
  __int64 v638; // rax
  __int64 v639; // rax
  int *v640; // rsi
  __int64 v641; // rax
  int v642; // eax
  __int64 v643; // rax
  __int64 v644; // rax
  __int64 v645; // rax
  int *v646; // r12
  __int64 v647; // rax
  __int64 v648; // rax
  __int64 v649; // rax
  __int64 v650; // rax
  __int64 v651; // rdi
  __int64 v652; // rbx
  __int64 v653; // rdi
  __int64 v654; // rax
  int *v655; // rbx
  __int64 v656; // rax
  int *v657; // rsi
  int v658; // eax
  _DWORD *v659; // rdi
  __int64 v660; // rax
  __int64 v661; // rax
  int *v662; // rdi
  __int64 v663; // rax
  __int64 v664; // rax
  __int64 v665; // rax
  int *v666; // rbx
  __int64 v667; // rax
  int *v668; // rdi
  __int64 v669; // rax
  int *v670; // rsi
  __int64 v671; // rax
  __int64 v672; // rax
  __int64 v673; // rax
  int *v674; // rsi
  int *v675; // rax
  int *v676; // rbx
  __int64 v677; // rax
  __int64 v678; // rax
  int *v679; // rbx
  int v680; // edi
  __int64 v681; // rax
  int *v682; // rbx
  __int64 v683; // rsi
  int *v684; // rdi
  __int64 v685; // rax
  int *v686; // rsi
  __int64 v687; // rax
  __int64 v688; // rax
  __int64 v689; // rax
  int *v690; // rsi
  __int64 v691; // rax
  int *v692; // rbx
  int *v693; // rdi
  __int64 v694; // rax
  int *v695; // rbx
  __int64 v696; // rax
  __int64 v697; // r14
  __int64 v698; // rax
  __int64 v699; // rbx
  __int64 v700; // rdi
  __int64 v701; // rax
  int *v702; // rbx
  __int64 v703; // rax
  __int64 v704; // rsi
  __int64 v705; // rax
  int *v706; // rbx
  int *v707; // rdi
  __int64 v708; // rax
  int *v709; // rbx
  __int64 v710; // rax
  __int64 v711; // rax
  __int64 v712; // rsi
  __int64 v713; // rax
  int *v714; // rbx
  int *v715; // rdi
  __int64 v716; // rax
  int *v717; // rbx
  __int64 v718; // rax
  __int64 v719; // rax
  __int64 v720; // rax
  __int64 v721; // rdx
  int *v722; // r15
  int *v723; // rsi
  int *v724; // rdi
  int *v725; // rbx
  int *v726; // rcx
  __int64 v727; // rax
  __int64 v728; // rax
  __int64 v729; // rax
  __int64 v730; // rax
  __int64 v731; // rax
  int *v732; // rbx
  __int64 v733; // rax
  int *v734; // rbx
  __int64 v735; // rax
  int *v736; // rbx
  __int64 v737; // rax
  _QWORD *v738; // rbx
  __int64 v739; // rax
  _QWORD *v740; // rbx
  __int64 v741; // rax
  int *v742; // rbx
  __int64 v743; // rax
  _QWORD *v744; // rbx
  __int64 v745; // rax
  _QWORD *v746; // rbx
  __int64 v747; // rax
  _QWORD *v748; // rbx
  __int64 v749; // rax
  int *v750; // rbx
  __int64 v751; // rax
  int *v752; // rbx
  __int64 v753; // rax
  int *v754; // rbx
  __int64 v755; // rax
  int *v756; // rbx
  __int64 v757; // rax
  int *v758; // rbx
  __int64 v759; // rax
  int *v760; // rbx
  __int64 v761; // rax
  int *v762; // rbx
  __int64 v763; // rax
  int *v764; // rbx
  __int64 v765; // rax
  int *v766; // rbx
  __int64 v767; // rax
  int *v768; // rbx
  __int64 v769; // rax
  int *v770; // rbx
  __int64 v771; // rax
  int *v772; // rbx
  __int64 v773; // rax
  __int64 v774; // rax
  int *v775; // rbx
  __int64 v776; // rax
  int *v777; // rbx
  __int64 v778; // rax
  int *v779; // rbx
  __int64 v780; // rax
  int *v781; // rbx
  __int64 v782; // rax
  int *v783; // rbx
  __int64 v784; // rax
  int *v785; // rbx
  __int64 v786; // rax
  int *v787; // rbx
  __int64 v788; // rax
  __int64 v790; // rax
  _QWORD *v791; // rsi
  __int64 v792; // rax
  int *v793; // rsi
  __int64 v794; // rax
  __int64 v795; // rax
  __int64 v796; // rax
  __int64 v797; // rax
  __int64 v798; // rax
  __int64 v799; // rbx
  __int64 v800; // rax
  int *v801; // r14
  __int64 v802; // rdx
  __int64 v803; // r8
  __int64 v804; // r9
  __int64 v805; // rcx
  int *v806; // rcx
  __int64 v807; // rbx
  int *v808; // rcx
  int *v809; // rcx
  _QWORD *v810; // rcx
  int *v811; // rcx
  int *v812; // rcx
  int *v813; // rcx
  int *v814; // rcx
  _QWORD *v815; // rcx
  _QWORD *v816; // rcx
  _QWORD *v817; // rcx
  int *v818; // rcx
  int *v819; // rcx
  int *v820; // rcx
  int *v821; // rcx
  int *v822; // rcx
  int *v823; // rcx
  int *v824; // rcx
  int *v825; // rcx
  int *v826; // rcx
  int *v827; // rcx
  int *v828; // rcx
  int *v829; // rcx
  int *v830; // rcx
  int *v831; // rcx
  int *v832; // rcx
  int *v833; // rcx
  int *v834; // rcx
  int *v835; // rcx
  int *v836; // rcx
  int *v837; // [rsp+20h] [rbp-190h]
  int *v838; // [rsp+28h] [rbp-188h]
  int *v839; // [rsp+30h] [rbp-180h]
  int *v840; // [rsp+38h] [rbp-178h]
  int *v841; // [rsp+40h] [rbp-170h]
  int *v842; // [rsp+48h] [rbp-168h]
  int *v843; // [rsp+50h] [rbp-160h]
  int *v844; // [rsp+58h] [rbp-158h]
  int *v845; // [rsp+60h] [rbp-150h]
  _QWORD *v846; // [rsp+68h] [rbp-148h]
  int *v847; // [rsp+70h] [rbp-140h]
  int *v848; // [rsp+78h] [rbp-138h]
  int *v849; // [rsp+80h] [rbp-130h]
  int *v850; // [rsp+88h] [rbp-128h]
  _QWORD *v851; // [rsp+A8h] [rbp-108h]
  _QWORD *v852; // [rsp+B0h] [rbp-100h]
  _QWORD *v853; // [rsp+B8h] [rbp-F8h]
  int *v854; // [rsp+C0h] [rbp-F0h]
  int *v855; // [rsp+C8h] [rbp-E8h]
  int *v856; // [rsp+D0h] [rbp-E0h]
  int *v857; // [rsp+D8h] [rbp-D8h]
  int *v858; // [rsp+E0h] [rbp-D0h]
  int *v859; // [rsp+E8h] [rbp-C8h]
  int *v860; // [rsp+F0h] [rbp-C0h]
  int *v861; // [rsp+100h] [rbp-B0h]
  __int64 v862; // [rsp+130h] [rbp-80h] BYREF
  int *v863; // [rsp+138h] [rbp-78h]
  _QWORD *v864; // [rsp+140h] [rbp-70h]
  int *v865; // [rsp+148h] [rbp-68h]
  __int64 v866; // [rsp+150h] [rbp-60h]
  _QWORD *v867; // [rsp+158h] [rbp-58h]
  int *v868; // [rsp+160h] [rbp-50h]
  _QWORD *v869; // [rsp+168h] [rbp-48h]
  __int64 v870; // [rsp+170h] [rbp-40h]
  int *v871; // [rsp+178h] [rbp-38h]
  __int64 v872; // [rsp+180h] [rbp-30h]
  int *v873; // [rsp+188h] [rbp-28h]
  int *v874; // [rsp+190h] [rbp-20h]
  int *v875; // [rsp+198h] [rbp-18h]
  int *v876; // [rsp+1A0h] [rbp-10h]
  int *v877; // [rsp+1A8h] [rbp-8h]
  int *v878; // [rsp+1B0h] [rbp+0h]
  _DWORD *v879; // [rsp+1B8h] [rbp+8h]
  _QWORD *v880; // [rsp+1C0h] [rbp+10h]
  int *v881; // [rsp+1C8h] [rbp+18h]
  _QWORD *v882; // [rsp+1D0h] [rbp+20h]
  _QWORD *v883; // [rsp+1D8h] [rbp+28h]
  _QWORD *v884; // [rsp+1E0h] [rbp+30h]
  int *v885; // [rsp+1F0h] [rbp+40h]
  int *v886; // [rsp+1F8h] [rbp+48h]
  int *v887; // [rsp+200h] [rbp+50h]
  int *v888; // [rsp+208h] [rbp+58h]
  int *v889; // [rsp+210h] [rbp+60h]
  int *v890; // [rsp+218h] [rbp+68h]
  int *v891; // [rsp+220h] [rbp+70h]
  int *v892; // [rsp+228h] [rbp+78h]
  int *v893; // [rsp+230h] [rbp+80h]
  int *v894; // [rsp+238h] [rbp+88h]
  int *v895; // [rsp+240h] [rbp+90h]
  int *v896; // [rsp+248h] [rbp+98h]
  int *v897; // [rsp+250h] [rbp+A0h]
  int *v898; // [rsp+258h] [rbp+A8h]
  int *v899; // [rsp+260h] [rbp+B0h]
  int *v900; // [rsp+268h] [rbp+B8h]
  int *v901; // [rsp+270h] [rbp+C0h]
  __int64 v902; // [rsp+278h] [rbp+C8h]
  _QWORD *v903; // [rsp+280h] [rbp+D0h]
  int *v904; // [rsp+288h] [rbp+D8h]
  __int64 v905; // [rsp+290h] [rbp+E0h]
  int *v906; // [rsp+298h] [rbp+E8h]
  __int64 v907; // [rsp+2A0h] [rbp+F0h] BYREF
  int *v908; // [rsp+2A8h] [rbp+F8h]
  int *v909; // [rsp+2B0h] [rbp+100h]
  int *v910; // [rsp+2B8h] [rbp+108h]
  int *v911; // [rsp+2C0h] [rbp+110h]
  int *v912; // [rsp+2C8h] [rbp+118h]
  int *v913; // [rsp+2D0h] [rbp+120h]
  int *v914; // [rsp+2D8h] [rbp+128h]
  int *v915; // [rsp+2E0h] [rbp+130h]
  _QWORD *v916; // [rsp+2E8h] [rbp+138h]
  _QWORD *v917; // [rsp+2F0h] [rbp+140h]
  _QWORD *v918; // [rsp+2F8h] [rbp+148h]
  int *v919; // [rsp+300h] [rbp+150h]
  int *v920; // [rsp+308h] [rbp+158h]
  int *v921; // [rsp+310h] [rbp+160h]
  _QWORD *v922; // [rsp+318h] [rbp+168h]
  _QWORD *v923; // [rsp+320h] [rbp+170h]
  int *v924; // [rsp+328h] [rbp+178h]
  __int64 v925; // [rsp+330h] [rbp+180h] BYREF
  int *v926; // [rsp+338h] [rbp+188h] BYREF
  int *v927; // [rsp+340h] [rbp+190h] BYREF
  __int64 v928; // [rsp+348h] [rbp+198h] BYREF
  int *v929; // [rsp+350h] [rbp+1A0h] BYREF
  __int64 v930; // [rsp+358h] [rbp+1A8h] BYREF
  int *v931; // [rsp+360h] [rbp+1B0h] BYREF
  _QWORD *v932; // [rsp+368h] [rbp+1B8h] BYREF
  int *v933; // [rsp+370h] [rbp+1C0h] BYREF
  __int64 v934; // [rsp+378h] [rbp+1C8h] BYREF
  int *v935; // [rsp+380h] [rbp+1D0h] BYREF
  _DWORD *v936; // [rsp+388h] [rbp+1D8h] BYREF
  int *v937; // [rsp+390h] [rbp+1E0h] BYREF
  __int64 v938; // [rsp+398h] [rbp+1E8h] BYREF
  int *v939; // [rsp+3A0h] [rbp+1F0h]
  _QWORD v940[2]; // [rsp+3A8h] [rbp+1F8h] BYREF
  _QWORD v941[2]; // [rsp+3B8h] [rbp+208h] BYREF
  _QWORD v942[2]; // [rsp+3C8h] [rbp+218h] BYREF
  _QWORD v943[2]; // [rsp+3D8h] [rbp+228h] BYREF
  _QWORD v944[2]; // [rsp+3E8h] [rbp+238h] BYREF
  _QWORD v945[2]; // [rsp+3F8h] [rbp+248h] BYREF
  _QWORD v946[2]; // [rsp+408h] [rbp+258h] BYREF
  _QWORD v947[2]; // [rsp+418h] [rbp+268h] BYREF
  _QWORD v948[2]; // [rsp+428h] [rbp+278h] BYREF
  _QWORD v949[2]; // [rsp+438h] [rbp+288h] BYREF
  _QWORD v950[2]; // [rsp+448h] [rbp+298h] BYREF
  _QWORD v951[2]; // [rsp+458h] [rbp+2A8h] BYREF
  _QWORD v952[2]; // [rsp+468h] [rbp+2B8h] BYREF
  _QWORD v953[2]; // [rsp+478h] [rbp+2C8h] BYREF
  _QWORD v954[2]; // [rsp+488h] [rbp+2D8h] BYREF
  _QWORD v955[2]; // [rsp+498h] [rbp+2E8h] BYREF
  _QWORD v956[2]; // [rsp+4A8h] [rbp+2F8h] BYREF
  _QWORD v957[3]; // [rsp+4B8h] [rbp+308h] BYREF
  _QWORD v958[3]; // [rsp+4D0h] [rbp+320h] BYREF
  _QWORD v959[3]; // [rsp+4E8h] [rbp+338h] BYREF
  _QWORD v960[3]; // [rsp+500h] [rbp+350h] BYREF
  _QWORD v961[4]; // [rsp+518h] [rbp+368h] BYREF
  _QWORD v962[4]; // [rsp+538h] [rbp+388h] BYREF
  _QWORD v963[4]; // [rsp+558h] [rbp+3A8h] BYREF
  _QWORD v964[4]; // [rsp+578h] [rbp+3C8h] BYREF
  _QWORD v965[5]; // [rsp+598h] [rbp+3E8h] BYREF
  _QWORD v966[5]; // [rsp+5C0h] [rbp+410h] BYREF
  _QWORD v967[5]; // [rsp+5E8h] [rbp+438h] BYREF
  _QWORD v968[6]; // [rsp+610h] [rbp+460h] BYREF

  v4 = qword_1803B7450;
  v5 = *a3;
  v881 = a3[1];
  v876 = a3[2];
  v878 = a3[3];
  v912 = a3[4];
  v888 = a3[5];
  v871 = a3[6];
  v913 = a3[7];
  v886 = nullptr;
  v6 = nullptr;
  v920 = nullptr;
  v7 = nullptr;
  v910 = nullptr;
  v8 = nullptr;
  v880 = nullptr;
  v911 = nullptr;
  v897 = nullptr;
  v896 = nullptr;
  v893 = nullptr;
  v867 = nullptr;
  v864 = nullptr;
  v865 = nullptr;
  v882 = nullptr;
  v883 = nullptr;
  v884 = nullptr;
  v900 = nullptr;
  v894 = nullptr;
  v898 = nullptr;
  v877 = nullptr;
  v899 = nullptr;
  v887 = nullptr;
  v885 = nullptr;
  v901 = nullptr;
  v889 = nullptr;
  v890 = nullptr;
  v891 = nullptr;
  v892 = nullptr;
  v874 = nullptr;
  v904 = nullptr;
  v906 = nullptr;
  v862 = 0;
  v875 = v5;
  v866 = qword_1803B7450;
  if ( qword_1803B7450 == 0 )
  {
LABEL_8:
    v4 = sub_18036CD10(a1, *(_QWORD *)&qword_1803B7448, qword_1803CA750, 280);
    v866 = v4;
    qword_1803B7450 = v4;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803B7450 > 1 || *(_QWORD *)(qword_1803B7450 + 16) != 0 )
  {
    if ( qword_1803B7450 != 0 && *(int *)qword_1803B7450 >= 0 )
    {
      v9 = (*(_QWORD *)qword_1803B7450)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v4 + 8) + 48LL))(v4);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v4);
  v10 = 216;
  *(_DWORD *)(v4 + 40) = 216;
// Original IDA line 1035 | qword_1803B7528 -> novels.pipeline.short_fiction_runner#5 bytes [4003365, 4003377]: "_cancelled"
  v11 = sub_18037E560(a1, v5, qword_1803B7528);
  v12 = (int *)v11;
  if ( v11 == 0 )
  {
    v862 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_1392;
  }
  v13 = sub_180001510(v11);
  if ( v13 == -1 )
  {
    v862 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v12 >= 0 )
    {
      v9 = (*(_QWORD *)v12)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v12 + 1) + 48LL))(v12);
    }
    v7 = nullptr;
    goto LABEL_1392;
  }
  if ( *v12 >= 0 )
  {
    v9 = (*(_QWORD *)v12)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v12 + 1) + 48LL))(v12);
  }
  if ( v13 != 0 )
  {
    v14 = sub_18015D410(a1);
    if ( v14 == 0 )
    {
// Original IDA line 1068 | qword_1803B77F0 -> novels.pipeline.short_fiction_runner#94 bytes [4004460, 4004485]: "OperationCancelledError"
      sub_180373340(a1, &v862, qword_1803B77F0);
      v10 = 217;
      v7 = nullptr;
      goto LABEL_1392;
    }
    v10 = 217;
    *(_DWORD *)(v4 + 40) = 217;
// Original IDA line 1075 | qword_1803B77F8 -> novels.pipeline.short_fiction_runner#95 bytes [4004485, 4004504]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4004487, 4004504], "value": "操作已取消"}]}
    v15 = (int *)sub_1803770D0(a1, v14, qword_1803B77F8);
    if ( v15 != nullptr )
    {
      v16 = sub_180373490(a1);
      if ( *v15 >= 0 )
      {
        v9 = (*(_QWORD *)v15)-- == 1;
        if ( v9 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v15 + 1) + 48LL))(v15);
      }
      if ( v16 != 0 )
      {
        v862 = v16;
        sub_18000D1D0(a1, v16);
        v7 = nullptr;
        goto LABEL_1392;
      }
    }
LABEL_23:
    v862 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    v7 = nullptr;
    goto LABEL_1392;
  }
// Original IDA line 1099 | qword_1803B7758 -> novels.pipeline.short_fiction_runner#75 bytes [4004234, 4004245]: "_try_read"
  v17 = (int *)sub_18037F0C0(a1, v875, *(_QWORD *)&qword_1803B7758);
  if ( v17 == nullptr )
  {
    v10 = 222;
    goto LABEL_23;
  }
// Original IDA line 1105 | qword_1803B7718 -> novels.pipeline.short_fiction_runner#67 bytes [4004155, 4004164]: "outline"
  v18 = sub_180382440(v871, qword_1803B7718);
  v19 = (int *)v18;
  if ( v18 == 0 )
  {
    v862 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v17 >= 0 )
    {
      v9 = (*(_QWORD *)v17)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v17 + 1) + 48LL))(v17);
    }
    v10 = 222;
    v7 = nullptr;
    goto LABEL_1392;
  }
// Original IDA line 1121 | qword_1803B7788 -> novels.pipeline.short_fiction_runner#81 bytes [4004291, 4004300]: "v002.md"
  v20 = (int *)sub_180382440(v18, qword_1803B7788);
  if ( *v19 >= 0 )
  {
    v9 = (*(_QWORD *)v19)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v19 + 1) + 48LL))(v19);
  }
  if ( v20 == nullptr )
  {
    v862 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v17 >= 0 )
    {
      v9 = (*(_QWORD *)v17)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v17 + 1) + 48LL))(v17);
    }
    v10 = 222;
    v7 = nullptr;
    v8 = nullptr;
    goto LABEL_1392;
  }
  v10 = 222;
  *(_DWORD *)(v4 + 40) = 222;
  v21 = sub_180376770(a1, v17, v20);
  v22 = *v17 < 0;
  v23 = (int *)v21;
  v921 = (int *)v21;
  if ( !v22 )
  {
    v9 = (*(_QWORD *)v17)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v17 + 1) + 48LL))(v17);
  }
  if ( *v20 >= 0 )
  {
    v9 = (*(_QWORD *)v20)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v20 + 1) + 48LL))(v20);
  }
  if ( v23 == nullptr )
    goto LABEL_51;
  v910 = v23;
  v25 = sub_180001510(v23);
  if ( v25 == -1 )
  {
    sub_18000CE30(a1, &v862);
    v6 = v867;
    v10 = 223;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1392;
  }
  if ( v25 != 0 )
  {
    if ( *v23 != -1 )
      ++*v23;
    v880 = v23;
    v26 = v23;
    v27 = sub_18015ED20(a1);
    if ( v27 == 0 )
    {
// Original IDA line 1183 | qword_1803B7800 -> novels.pipeline.short_fiction_runner#96 bytes [4004504, 4004512]: "logger"
      sub_180373340(a1, &v862, qword_1803B7800);
      v6 = v867;
      v10 = 225;
      v7 = v867;
      v8 = (int *)v867;
      goto LABEL_1392;
    }
    v10 = 225;
    *(_DWORD *)(v4 + 40) = 225;
// Original IDA line 1192 | qword_1803B7808 -> novels.pipeline.short_fiction_runner#97 bytes [4004512, 4004518]: "info"; qword_1803B7810 -> novels.pipeline.short_fiction_runner#98 bytes [4004518, 4004574]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4004520, 4004574], "value": "Resuming from existing outline (skipping stages 0-2)"}]}
    v28 = (int *)sub_18037E6C0(a1, v27, qword_1803B7808, *(_QWORD *)(qword_1803B7810 + 24));
    v29 = v28;
    if ( v28 == nullptr )
      goto LABEL_51;
    if ( *v28 >= 0 )
    {
      v9 = (*(_QWORD *)v28)-- == 1;
      if ( v9 )
      {
        v30 = sub_18000CC40(v28);
        (*(void (__fastcall **)(int *))(v30 + 48))(v29);
      }
    }
  }
  else
  {
    v31 = v881;
    v32 = v875;
    v10 = 227;
    *(_DWORD *)(v4 + 40) = 227;
    v965[0] = v31;
    v965[1] = v876;
    v965[2] = v878;
    v965[3] = v912;
    v965[4] = v871;
// Original IDA line 1217 | qword_1803B7818 -> novels.pipeline.short_fiction_runner#99 bytes [4004574, 4004595]: "_run_outline_stages"
    v33 = sub_18037EC40(a1, v32, *(_QWORD *)&qword_1803B7818, v965);
    if ( v33 == 0 )
      goto LABEL_51;
    v26 = (int *)v33;
    v880 = (_QWORD *)v33;
  }
  v34 = v875;
  v10 = 231;
  *(_DWORD *)(v4 + 40) = 231;
// Original IDA line 1226 | qword_1803B7528 -> novels.pipeline.short_fiction_runner#5 bytes [4003365, 4003377]: "_cancelled"
  v35 = sub_18037E560(a1, v34, qword_1803B7528);
  v36 = (int *)v35;
  if ( v35 == 0 )
    goto LABEL_51;
  v37 = sub_180001510(v35);
  if ( v37 == -1 )
  {
    sub_18000CE30(a1, &v862);
    if ( *v36 >= 0 )
    {
      v9 = (*(_QWORD *)v36)-- == 1;
      if ( v9 )
      {
        v38 = sub_18000CC40(v36);
        (*(void (__fastcall **)(int *))(v38 + 48))(v36);
        v6 = v867;
        v7 = v867;
        v8 = (int *)v867;
        goto LABEL_1392;
      }
    }
LABEL_53:
    v6 = v867;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1392;
  }
  if ( *v36 >= 0 )
  {
    v9 = (*(_QWORD *)v36)-- == 1;
    if ( v9 )
    {
      v39 = sub_18000CC40(v36);
      (*(void (__fastcall **)(int *))(v39 + 48))(v36);
    }
  }
  if ( v37 != 0 )
  {
    v40 = sub_18015D410(a1);
    if ( v40 == 0 )
    {
// Original IDA line 1267 | qword_1803B77F0 -> novels.pipeline.short_fiction_runner#94 bytes [4004460, 4004485]: "OperationCancelledError"
      sub_180373340(a1, &v862, qword_1803B77F0);
      v6 = v867;
      v10 = 232;
      v7 = v867;
      v8 = (int *)v867;
      goto LABEL_1392;
    }
    v10 = 232;
    *(_DWORD *)(v4 + 40) = 232;
// Original IDA line 1276 | qword_1803B77F8 -> novels.pipeline.short_fiction_runner#95 bytes [4004485, 4004504]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4004487, 4004504], "value": "操作已取消"}]}
    v41 = (int *)sub_1803770D0(a1, v40, qword_1803B77F8);
    v24 = a1;
    if ( v41 != nullptr )
    {
      v42 = sub_180373490(a1);
      if ( *v41 >= 0 )
      {
        v9 = (*(_QWORD *)v41)-- == 1;
        if ( v9 )
        {
          v43 = sub_18000CC40(v41);
          (*(void (__fastcall **)(int *))(v43 + 48))(v41);
        }
      }
      v24 = a1;
      if ( v42 != 0 )
      {
        v862 = v42;
        sub_180373100(a1, &v862);
        v6 = v867;
        v7 = v867;
        v8 = (int *)v867;
        goto LABEL_1392;
      }
    }
    goto LABEL_52;
  }
  v44 = v871;
  v45 = v875;
  v10 = 234;
  *(_DWORD *)(v4 + 40) = 234;
  v961[0] = v44;
  v961[1] = v888;
  v961[2] = v913;
  v961[3] = v26;
// Original IDA line 1311 | qword_1803B7820 -> novels.pipeline.short_fiction_runner#100 bytes [4004595, 4004623]: "_maybe_pause_after_outline"
  v46 = sub_18037EAE0(a1, v45, *(_QWORD *)&qword_1803B7820, v961);
  v895 = (int *)v46;
  v47 = (int *)v46;
  if ( v46 == 0 )
    goto LABEL_51;
  v911 = (int *)v46;
  v48 = sub_180001510(v46);
  if ( v48 == -1 )
  {
    sub_18000CE30(a1, &v862);
    v6 = v867;
    v10 = 235;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1392;
  }
  if ( v48 != 0 )
  {
    v50 = v47;
    if ( *v47 != -1 )
      ++*v47;
    sub_180001D80(a1, v49);
LABEL_1242:
    if ( *v23 >= 0 )
    {
      v9 = (*(_QWORD *)v23)-- == 1;
      if ( v9 )
      {
        v729 = sub_18000CC40(v910);
        (*(void (__fastcall **)(int *))(v729 + 48))(v910);
      }
    }
    if ( v26 != nullptr && *v26 >= 0 )
    {
      v9 = (*(_QWORD *)v26)-- == 1;
      if ( v9 )
      {
        v730 = sub_18000CC40(v26);
        (*(void (__fastcall **)(int *))(v730 + 48))(v26);
      }
    }
    if ( *v47 >= 0 )
    {
      v9 = (*(_QWORD *)v47)-- == 1;
      if ( v9 )
      {
        v731 = sub_18000CC40(v911);
        (*(void (__fastcall **)(int *))(v731 + 48))(v911);
      }
    }
    v732 = v897;
    if ( v897 != nullptr && *v897 >= 0 )
    {
      v9 = (*(_QWORD *)v897)-- == 1;
      if ( v9 )
      {
        v733 = sub_18000CC40(v732);
        (*(void (__fastcall **)(int *))(v733 + 48))(v732);
      }
    }
    v734 = v896;
    if ( v896 != nullptr && *v896 >= 0 )
    {
      v9 = (*(_QWORD *)v896)-- == 1;
      if ( v9 )
      {
        v735 = sub_18000CC40(v734);
        (*(void (__fastcall **)(int *))(v735 + 48))(v734);
      }
    }
    v736 = v893;
    if ( v893 != nullptr && *v893 >= 0 )
    {
      v9 = (*(_QWORD *)v893)-- == 1;
      if ( v9 )
      {
        v737 = sub_18000CC40(v736);
        (*(void (__fastcall **)(int *))(v737 + 48))(v736);
      }
    }
    v738 = v867;
    if ( v867 != nullptr && *(int *)v867 >= 0 )
    {
      v9 = (*v867)-- == 1;
      if ( v9 )
      {
        v739 = sub_18000CC40(v738);
        (*(void (__fastcall **)(_QWORD *))(v739 + 48))(v738);
      }
    }
    v740 = v864;
    if ( v864 != nullptr && *(int *)v864 >= 0 )
    {
      v9 = (*v864)-- == 1;
      if ( v9 )
      {
        v741 = sub_18000CC40(v740);
        (*(void (__fastcall **)(_QWORD *))(v741 + 48))(v740);
      }
    }
    v742 = v865;
    if ( v865 != nullptr && *v865 >= 0 )
    {
      v9 = (*(_QWORD *)v865)-- == 1;
      if ( v9 )
      {
        v743 = sub_18000CC40(v742);
        (*(void (__fastcall **)(int *))(v743 + 48))(v742);
      }
    }
    v744 = v882;
    if ( v882 != nullptr && *(int *)v882 >= 0 )
    {
      v9 = (*v882)-- == 1;
      if ( v9 )
      {
        v745 = sub_18000CC40(v744);
        (*(void (__fastcall **)(_QWORD *))(v745 + 48))(v744);
      }
    }
    v746 = v883;
    if ( v883 != nullptr && *(int *)v883 >= 0 )
    {
      v9 = (*v883)-- == 1;
      if ( v9 )
      {
        v747 = sub_18000CC40(v746);
        (*(void (__fastcall **)(_QWORD *))(v747 + 48))(v746);
      }
    }
    v748 = v884;
    if ( v884 != nullptr && *(int *)v884 >= 0 )
    {
      v9 = (*v884)-- == 1;
      if ( v9 )
      {
        v749 = sub_18000CC40(v748);
        (*(void (__fastcall **)(_QWORD *))(v749 + 48))(v748);
      }
    }
    v750 = v900;
    if ( v900 != nullptr && *v900 >= 0 )
    {
      v9 = (*(_QWORD *)v900)-- == 1;
      if ( v9 )
      {
        v751 = sub_18000CC40(v750);
        (*(void (__fastcall **)(int *))(v751 + 48))(v750);
      }
    }
    v752 = v894;
    if ( v894 != nullptr && *v894 >= 0 )
    {
      v9 = (*(_QWORD *)v894)-- == 1;
      if ( v9 )
      {
        v753 = sub_18000CC40(v752);
        (*(void (__fastcall **)(int *))(v753 + 48))(v752);
      }
    }
    v754 = v898;
    if ( v898 != nullptr && *v898 >= 0 )
    {
      v9 = (*(_QWORD *)v898)-- == 1;
      if ( v9 )
      {
        v755 = sub_18000CC40(v754);
        (*(void (__fastcall **)(int *))(v755 + 48))(v754);
      }
    }
    v756 = v877;
    if ( v877 != nullptr && *v877 >= 0 )
    {
      v9 = (*(_QWORD *)v877)-- == 1;
      if ( v9 )
      {
        v757 = sub_18000CC40(v756);
        (*(void (__fastcall **)(int *))(v757 + 48))(v756);
      }
    }
    v758 = v899;
    if ( v899 != nullptr && *v899 >= 0 )
    {
      v9 = (*(_QWORD *)v899)-- == 1;
      if ( v9 )
      {
        v759 = sub_18000CC40(v758);
        (*(void (__fastcall **)(int *))(v759 + 48))(v758);
      }
    }
    v760 = v887;
    if ( v887 != nullptr && *v887 >= 0 )
    {
      v9 = (*(_QWORD *)v887)-- == 1;
      if ( v9 )
      {
        v761 = sub_18000CC40(v760);
        (*(void (__fastcall **)(int *))(v761 + 48))(v760);
      }
    }
    v762 = v885;
    if ( v885 != nullptr && *v885 >= 0 )
    {
      v9 = (*(_QWORD *)v885)-- == 1;
      if ( v9 )
      {
        v763 = sub_18000CC40(v762);
        (*(void (__fastcall **)(int *))(v763 + 48))(v762);
      }
    }
    v764 = v901;
    if ( v901 != nullptr && *v901 >= 0 )
    {
      v9 = (*(_QWORD *)v901)-- == 1;
      if ( v9 )
      {
        v765 = sub_18000CC40(v764);
        (*(void (__fastcall **)(int *))(v765 + 48))(v764);
      }
    }
    v766 = v889;
    if ( v889 != nullptr && *v889 >= 0 )
    {
      v9 = (*(_QWORD *)v889)-- == 1;
      if ( v9 )
      {
        v767 = sub_18000CC40(v766);
        (*(void (__fastcall **)(int *))(v767 + 48))(v766);
      }
    }
    v768 = v890;
    if ( v890 != nullptr && *v890 >= 0 )
    {
      v9 = (*(_QWORD *)v890)-- == 1;
      if ( v9 )
      {
        v769 = sub_18000CC40(v768);
        (*(void (__fastcall **)(int *))(v769 + 48))(v768);
      }
    }
    v770 = v891;
    if ( v891 != nullptr && *v891 >= 0 )
    {
      v9 = (*(_QWORD *)v891)-- == 1;
      if ( v9 )
      {
        v771 = sub_18000CC40(v770);
        (*(void (__fastcall **)(int *))(v771 + 48))(v770);
      }
    }
    v772 = v892;
    if ( v892 != nullptr && *v892 >= 0 )
    {
      v9 = (*(_QWORD *)v892)-- == 1;
      if ( v9 )
      {
        v773 = sub_18000CC40(v772);
        (*(void (__fastcall **)(int *))(v773 + 48))(v772);
      }
    }
    if ( *v45 >= 0 )
    {
      v9 = (*(_QWORD *)v45)-- == 1;
      if ( v9 )
      {
        v774 = sub_18000CC40(v45);
        (*(void (__fastcall **)(int *))(v774 + 48))(v45);
      }
    }
    v775 = v881;
    if ( *v881 >= 0 )
    {
      v9 = (*(_QWORD *)v881)-- == 1;
      if ( v9 )
      {
        v776 = sub_18000CC40(v775);
        (*(void (__fastcall **)(int *))(v776 + 48))(v775);
      }
    }
    v777 = v876;
    if ( *v876 >= 0 )
    {
      v9 = (*(_QWORD *)v876)-- == 1;
      if ( v9 )
      {
        v778 = sub_18000CC40(v777);
        (*(void (__fastcall **)(int *))(v778 + 48))(v777);
      }
    }
    v779 = v878;
    if ( *v878 >= 0 )
    {
      v9 = (*(_QWORD *)v878)-- == 1;
      if ( v9 )
      {
        v780 = sub_18000CC40(v779);
        (*(void (__fastcall **)(int *))(v780 + 48))(v779);
      }
    }
    v781 = v912;
    if ( *v912 >= 0 )
    {
      v9 = (*(_QWORD *)v912)-- == 1;
      if ( v9 )
      {
        v782 = sub_18000CC40(v781);
        (*(void (__fastcall **)(int *))(v782 + 48))(v781);
      }
    }
    v783 = v888;
    if ( *v888 >= 0 )
    {
      v9 = (*(_QWORD *)v888)-- == 1;
      if ( v9 )
      {
        v784 = sub_18000CC40(v783);
        (*(void (__fastcall **)(int *))(v784 + 48))(v783);
      }
    }
    v785 = v871;
    if ( *v871 >= 0 )
    {
      v9 = (*(_QWORD *)v871)-- == 1;
      if ( v9 )
      {
        v786 = sub_18000CC40(v785);
        (*(void (__fastcall **)(int *))(v786 + 48))(v785);
      }
    }
    v787 = v913;
    if ( *v913 >= 0 )
    {
      v9 = (*(_QWORD *)v913)-- == 1;
      if ( v9 )
      {
        v788 = sub_18000CC40(v787);
        (*(void (__fastcall **)(int *))(v788 + 48))(v787);
      }
    }
    return v50;
  }
  v10 = 241;
  *(_DWORD *)(v4 + 40) = 241;
// Original IDA line 1654 | qword_1803B7830 -> novels.pipeline.short_fiction_runner#102 bytes [4004631, 4004657]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4004633, 4004635], "value": 3}, {"tag": "a", "fileRange": [4004635, 4004643], "value": "active"}, {"tag": "u", "fileRange": [4004643, 4004657], "value": "撰写正文"}]}; qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"
  v51 = (int *)sub_18037E980(a1, v45, *(_QWORD *)&qword_1803B7828, qword_1803B7830 + 24);
  v52 = v51;
  if ( v51 == nullptr )
  {
LABEL_51:
    v24 = a1;
LABEL_52:
    sub_18000CE30(v24, &v862);
    goto LABEL_53;
  }
  if ( *v51 >= 0 )
  {
    v9 = (*(_QWORD *)v51)-- == 1;
    if ( v9 )
    {
      v53 = sub_18000CC40(v51);
      (*(void (__fastcall **)(int *))(v53 + 48))(v52);
    }
  }
  v10 = 242;
  *(_DWORD *)(v4 + 40) = 242;
// Original IDA line 1675 | qword_1803B7838 -> novels.pipeline.short_fiction_runner#103 bytes [4004657, 4004677]: "_writer_sys_prompt"
  v54 = sub_18037E560(a1, v45, *(_QWORD *)&qword_1803B7838);
  v908 = (int *)v54;
  v55 = v54;
  v24 = a1;
  if ( v54 == 0 )
    goto LABEL_52;
  v897 = (int *)v54;
  v56 = sub_18015D850(a1);
  if ( v56 == 0 )
  {
// Original IDA line 1685 | qword_1803B7840 -> novels.pipeline.short_fiction_runner#104 bytes [4004677, 4004707]: "ShortFictionDraftPromptInput"
    sub_180373340(a1, &v862, qword_1803B7840);
    v6 = v867;
    v10 = 243;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1392;
  }
  v57 = v881;
  v10 = 243;
  *(_DWORD *)(v4 + 40) = 243;
  v962[0] = v57;
  v962[2] = v876;
  v962[3] = v878;
  v962[1] = v26;
// Original IDA line 1699 | qword_1803B7848 -> novels.pipeline.short_fiction_runner#105 bytes [4004707, 4004772]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004709, 4004720], "value": "direction"}, {"tag": "a", "fileRange": [4004720, 4004738], "value": "outline_markdown"}, {"tag": "a", "fileRange": [4004738, 4004753], "value": "chapter_cou... (truncated literal)
  v58 = sub_18037CCB0(a1, v56, v962, qword_1803B7848);
  v909 = (int *)v58;
  v24 = a1;
  if ( v58 == 0 )
    goto LABEL_52;
  v896 = (int *)v58;
// Original IDA line 1705 | qword_1803B7850 -> novels.pipeline.short_fiction_runner#106 bytes [4004772, 4004779]: "_chat"
  v59 = (int *)sub_18037F0C0(a1, v45, *(_QWORD *)&qword_1803B7850);
  if ( v59 == nullptr )
  {
    sub_18000CE30(a1, &v862);
    v6 = v867;
    v10 = 249;
    v7 = v867;
    v8 = (int *)v867;
LABEL_1391:
    v4 = v866;
    goto LABEL_1392;
  }
// Original IDA line 1717 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
  v60 = qword_1803B7860;
// Original IDA line 1718 | qword_1803B7868 -> novels.pipeline.short_fiction_runner#109 bytes [4004793, 4004801]: "system"
  v61 = qword_1803B7868;
// Original IDA line 1719 | qword_1803B7858 -> novels.pipeline.short_fiction_runner#107 bytes [4004779, 4004787]: "writer"
  v869 = (_QWORD *)qword_1803B7858;
  v62 = PyDict_NewPresized(2);
  PyDict_SetItem(v62, v60, v61);
// Original IDA line 1722 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
  PyDict_SetItem(v62, qword_1803B7870, v55);
  v63 = (int *)sub_180372160(a1, 2);
  sub_180024750(v63, 0, v62);
// Original IDA line 1725 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
  v64 = qword_1803B7860;
// Original IDA line 1726 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
  v65 = qword_1803B7878;
  v66 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v66, v64, v65);
// Original IDA line 1729 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
  v67 = qword_1803B7870;
  v68 = sub_18015E9F0(a1);
  if ( v68 == 0 )
  {
// Original IDA line 1733 | qword_1803B7880 -> novels.pipeline.short_fiction_runner#112 bytes [4004816, 4004856]: "build_short_fiction_writer_user_prompt"
    sub_180373340(a1, &v862, qword_1803B7880);
    v10 = 253;
LABEL_105:
    if ( *v66 >= 0 )
    {
      v9 = (*(_QWORD *)v66)-- == 1;
      if ( v9 )
      {
        v72 = sub_18000CC40(v66);
        (*(void (__fastcall **)(int *))(v72 + 48))(v66);
      }
    }
    if ( *v59 >= 0 )
    {
      v9 = (*(_QWORD *)v59)-- == 1;
      if ( v9 )
      {
        v73 = sub_18000CC40(v59);
        (*(void (__fastcall **)(int *))(v73 + 48))(v59);
      }
    }
    if ( *v63 >= 0 )
    {
      v9 = (*(_QWORD *)v63)-- == 1;
      if ( v9 )
      {
        v74 = sub_18000CC40(v63);
        (*(void (__fastcall **)(int *))(v74 + 48))(v63);
        v6 = v867;
        v7 = v867;
        v8 = (int *)v867;
        goto LABEL_1391;
      }
    }
    goto LABEL_1388;
  }
  v10 = 253;
  v69 = v909;
  *(_DWORD *)(v866 + 40) = 253;
  v70 = sub_180376770(a1, v68, v69);
  v71 = (int *)v70;
  if ( v70 == 0 )
  {
    sub_18000CE30(a1, &v862);
    goto LABEL_105;
  }
  PyDict_SetItem(v66, v67, v70);
  if ( *v71 >= 0 )
  {
    v9 = (*(_QWORD *)v71)-- == 1;
    if ( v9 )
    {
      v75 = sub_18000CC40(v71);
      (*(void (__fastcall **)(int *))(v75 + 48))(v71);
    }
  }
  sub_180024750(v63, 1, v66);
// Original IDA line 1790 | qword_1803B7888 -> novels.pipeline.short_fiction_runner#113 bytes [4004856, 4004870]: "撰写正文"
  v76 = qword_1803B7888;
  v77 = sub_18015EB00(a1);
  if ( v77 == 0 )
  {
// Original IDA line 1794 | qword_1803B7890 -> novels.pipeline.short_fiction_runner#114 bytes [4004870, 4004905]: "estimate_short_fiction_max_tokens"
    sub_180373340(a1, &v862, qword_1803B7890);
    if ( *v59 >= 0 )
    {
      v9 = (*(_QWORD *)v59)-- == 1;
      if ( v9 )
      {
        v78 = sub_18000CC40(v59);
        (*(void (__fastcall **)(int *))(v78 + 48))(v59);
      }
    }
    if ( *v63 >= 0 )
    {
      v9 = (*(_QWORD *)v63)-- == 1;
      if ( v9 )
      {
        v79 = sub_18000CC40(v63);
        (*(void (__fastcall **)(int *))(v79 + 48))(v63);
      }
    }
    v6 = v867;
    v10 = 256;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1391;
  }
  v80 = v876;
  v81 = v866;
  v10 = 256;
  *(_DWORD *)(v866 + 40) = 256;
  v940[0] = v80;
  v940[1] = v878;
  v82 = (int *)sub_180377750(a1, v77, v940);
  if ( v82 == nullptr )
  {
    sub_18000CE30(a1, &v862);
    if ( *v59 >= 0 )
    {
      v9 = (*(_QWORD *)v59)-- == 1;
      if ( v9 )
      {
        v83 = sub_18000CC40(v59);
        (*(void (__fastcall **)(int *))(v83 + 48))(v59);
      }
    }
    if ( *v63 >= 0 )
    {
      v9 = (*(_QWORD *)v63)-- == 1;
      if ( v9 )
      {
        v84 = sub_18000CC40(v63);
        (*(void (__fastcall **)(int *))(v84 + 48))(v63);
      }
    }
    v6 = v867;
    v4 = v81;
    v7 = v867;
    v8 = (int *)v867;
    goto LABEL_1392;
  }
// Original IDA line 1853 | qword_1803B7898 -> novels.pipeline.short_fiction_runner#115 bytes [4004905, 4004914]: 0.58
  v85 = qword_1803B7898;
  v86 = v869;
  v10 = 249;
  *(_DWORD *)(v81 + 40) = 249;
  v957[2] = v85;
  v941[0] = v86;
  v941[1] = v63;
  v957[0] = v76;
  v957[1] = v82;
// Original IDA line 1862 | qword_1803B78A0 -> novels.pipeline.short_fiction_runner#116 bytes [4004914, 4004954]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004916, 4004929], "value": "stage_label"}, {"tag": "a", "fileRange": [4004929, 4004941], "value": "max_tokens"}, {"tag": "a", "fileRange": [4004941, 4004954], "value": "temperature"}]}
  v87 = sub_18037D650(a1, v59, v941, v957, qword_1803B78A0);
  v22 = *v59 < 0;
  v88 = (int *)v87;
  v924 = (int *)v87;
  if ( !v22 )
  {
    v9 = (*(_QWORD *)v59)-- == 1;
    if ( v9 )
    {
      v89 = sub_18000CC40(v59);
      (*(void (__fastcall **)(int *))(v89 + 48))(v59);
    }
  }
  if ( *v63 >= 0 )
  {
    v9 = (*(_QWORD *)v63)-- == 1;
    if ( v9 )
    {
      v90 = sub_18000CC40(v63);
      (*(void (__fastcall **)(int *))(v90 + 48))(v63);
    }
  }
  if ( *v82 >= 0 )
  {
    v9 = (*(_QWORD *)v82)-- == 1;
    if ( v9 )
    {
      v91 = sub_18000CC40(v82);
      (*(void (__fastcall **)(int *))(v91 + 48))(v82);
    }
  }
  v92 = a1;
  if ( v88 != nullptr )
  {
    v893 = v88;
    v93 = sub_18015EE30(a1);
    if ( v93 == 0 )
    {
// Original IDA line 1900 | qword_1803B78A8 -> novels.pipeline.short_fiction_runner#117 bytes [4004954, 4004987]: "parse_short_fiction_batch_draft"
      sub_180373340(a1, &v862, qword_1803B78A8);
      v6 = v867;
      v4 = v81;
      v7 = v867;
      v8 = (int *)v867;
      v10 = 259;
      goto LABEL_1392;
    }
    v94 = v876;
    v10 = 259;
    *(_DWORD *)(v81 + 40) = 259;
    v927 = v88;
    v926 = v94;
// Original IDA line 1913 | qword_1803B78B0 -> novels.pipeline.short_fiction_runner#118 bytes [4004987, 4005008]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004989, 4005008], "value": "expected_chapters"}]}
    v95 = sub_18037D030(a1, v93, &v927, &v926, qword_1803B78B0);
    v96 = v95;
    v92 = a1;
    if ( v95 != 0 )
    {
      v6 = (_QWORD *)v95;
      v867 = (_QWORD *)v95;
      v97 = sub_18015EC10(a1);
      if ( v97 == 0 )
      {
// Original IDA line 1923 | qword_1803B78B8 -> novels.pipeline.short_fiction_runner#119 bytes [4005008, 4005043]: "find_empty_short_fiction_chapters"
        sub_180373340(a1, &v862, qword_1803B78B8);
        v7 = v864;
        v4 = v81;
        v8 = (int *)v864;
        v10 = 262;
        goto LABEL_1392;
      }
      v10 = 262;
      *(_DWORD *)(v81 + 40) = 262;
      v98 = sub_180376770(a1, v97, v96);
      if ( v98 == 0 )
      {
        sub_18000CE30(a1, &v862);
        v7 = v864;
        v4 = v81;
        v8 = (int *)v864;
        goto LABEL_1392;
      }
      v864 = (_QWORD *)v98;
      v7 = (_QWORD *)v98;
      v99 = sub_180001510(v98);
      if ( v99 == -1 )
      {
        sub_18000CE30(a1, &v862);
        v4 = v81;
        v10 = 263;
        v8 = v865;
        goto LABEL_1392;
      }
      if ( v99 == 0 )
        goto LABEL_426;
      v100 = sub_18015D300(a1);
      if ( v100 == 0 )
      {
// Original IDA line 1957 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
        sub_180373340(a1, &v862, qword_1803B7610);
        v4 = v81;
        v10 = 264;
        v8 = v865;
        goto LABEL_1392;
      }
// Original IDA line 1963 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
      v101 = (int *)sub_18037F0C0(a1, v100, qword_1803B78C0);
      if ( v101 == nullptr )
      {
        sub_18000CE30(a1, &v862);
        v4 = v81;
        v10 = 264;
        v8 = v865;
        goto LABEL_1392;
      }
// Original IDA line 1972 | qword_1803B7728 -> novels.pipeline.short_fiction_runner#69 bytes [4004173, 4004181]: "drafts"
      v102 = sub_180382540(v871, qword_1803B7728);
      v103 = (int *)v102;
      if ( v102 == 0 )
      {
        sub_18000CE30(a1, &v862);
        if ( *v101 >= 0 )
        {
          v9 = (*(_QWORD *)v101)-- == 1;
          if ( v9 )
          {
            v104 = sub_18000CC40(v101);
            (*(void (__fastcall **)(int *))(v104 + 48))(v101);
          }
        }
        v10 = 265;
        goto LABEL_1390;
      }
// Original IDA line 1989 | qword_1803B78C8 -> novels.pipeline.short_fiction_runner#121 bytes [4005055, 4005072]: "v001-partial.md"
      v105 = (int *)sub_180382540(v102, qword_1803B78C8);
      if ( *v103 >= 0 )
      {
        v9 = (*(_QWORD *)v103)-- == 1;
        if ( v9 )
        {
          v106 = sub_18000CC40(v103);
          (*(void (__fastcall **)(int *))(v106 + 48))(v103);
        }
      }
      if ( v105 == nullptr )
      {
        sub_18000CE30(a1, &v862);
        if ( *v101 >= 0 )
        {
          v9 = (*(_QWORD *)v101)-- == 1;
          if ( v9 )
          {
            v107 = sub_18000CC40(v101);
            (*(void (__fastcall **)(int *))(v107 + 48))(v101);
          }
        }
        v10 = 265;
        goto LABEL_1389;
      }
      v108 = sub_18015F160(a1);
      if ( v108 == 0 )
      {
// Original IDA line 2017 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
        sub_180373340(a1, &v862, qword_1803B78D0);
        if ( *v101 >= 0 )
        {
          v9 = (*(_QWORD *)v101)-- == 1;
          if ( v9 )
          {
            v109 = sub_18000CC40(v101);
            (*(void (__fastcall **)(int *))(v109 + 48))(v101);
          }
        }
        if ( *v105 >= 0 )
        {
          v9 = (*(_QWORD *)v105)-- == 1;
          if ( v9 )
          {
            v110 = sub_18000CC40(v105);
            (*(void (__fastcall **)(int *))(v110 + 48))(v105);
          }
        }
        v10 = 266;
        goto LABEL_1389;
      }
      v111 = v866;
      v10 = 266;
      *(_DWORD *)(v866 + 40) = 266;
      v112 = sub_180376770(a1, v108, v96);
      v113 = (int *)v112;
      if ( v112 == 0 )
      {
        sub_18000CE30(a1, &v862);
        if ( *v101 >= 0 )
        {
          v9 = (*(_QWORD *)v101)-- == 1;
          if ( v9 )
          {
            v114 = sub_18000CC40(v101);
            (*(void (__fastcall **)(int *))(v114 + 48))(v101);
          }
        }
        if ( *v105 >= 0 )
        {
          v9 = (*(_QWORD *)v105)-- == 1;
          if ( v9 )
          {
            v115 = sub_18000CC40(v105);
            (*(void (__fastcall **)(int *))(v115 + 48))(v105);
          }
        }
        v7 = v864;
        v4 = v111;
        v6 = v867;
        v8 = v865;
        goto LABEL_1392;
      }
      v10 = 264;
      *(_DWORD *)(v111 + 40) = 264;
      v942[0] = v105;
      v942[1] = v112;
      v116 = (int *)sub_180377750(a1, v101, v942);
      if ( *v101 >= 0 )
      {
        v9 = (*(_QWORD *)v101)-- == 1;
        if ( v9 )
        {
          v117 = sub_18000CC40(v101);
          (*(void (__fastcall **)(int *))(v117 + 48))(v101);
        }
      }
      if ( *v105 >= 0 )
      {
        v9 = (*(_QWORD *)v105)-- == 1;
        if ( v9 )
        {
          v118 = sub_18000CC40(v105);
          (*(void (__fastcall **)(int *))(v118 + 48))(v105);
        }
      }
      if ( *v113 >= 0 )
      {
        v9 = (*(_QWORD *)v113)-- == 1;
        if ( v9 )
        {
          v119 = sub_18000CC40(v113);
          (*(void (__fastcall **)(int *))(v119 + 48))(v113);
        }
      }
      if ( v116 == nullptr )
        goto LABEL_197;
      if ( *v116 >= 0 )
      {
        v9 = (*(_QWORD *)v116)-- == 1;
        if ( v9 )
        {
          v121 = sub_18000CC40(v116);
          (*(void (__fastcall **)(int *))(v121 + 48))(v116);
        }
      }
      v122 = qword_1803AB258;
      v123 = sub_18015D520(a1);
      if ( v123 == 0 )
      {
// Original IDA line 2118 | qword_1803B78D8 -> novels.pipeline.short_fiction_runner#123 bytes [4005109, 4005150]: "SHORT_FICTION_DRAFT_COMPLETION_ATTEMPTS"
        sub_180373340(a1, &v862, qword_1803B78D8);
LABEL_204:
        v4 = v111;
        v10 = 268;
        v6 = v867;
        goto LABEL_205;
      }
      v124 = sub_180380B80(v123, qword_1803AB258);
      v125 = (int *)v124;
      if ( v124 != 0 )
      {
        v126 = (int *)sub_180375230(a1, v122, v124);
        if ( *v125 >= 0 )
        {
          v9 = (*(_QWORD *)v125)-- == 1;
          if ( v9 )
          {
            v127 = sub_18000CC40(v125);
            (*(void (__fastcall **)(int *))(v127 + 48))(v125);
          }
        }
        if ( v126 != nullptr )
        {
          v128 = sub_18000D280(a1, v126);
          v22 = *v126 < 0;
          v129 = (int *)v128;
          v903 = (_QWORD *)v128;
          if ( !v22 )
          {
            v9 = (*(_QWORD *)v126)-- == 1;
            if ( v9 )
            {
              v130 = sub_18000CC40(v126);
              (*(void (__fastcall **)(int *))(v130 + 48))(v126);
            }
          }
          if ( v129 != nullptr )
          {
            v7 = v864;
            v4 = v111;
            v6 = v867;
            v131 = (int *)sub_180038D80(v129);
            v868 = v131;
            if ( v131 == nullptr )
            {
LABEL_349:
              if ( (unsigned __int8)sub_18000CDA0(a1) == 0 )
              {
LABEL_350:
                sub_18000CE30(a1, &v862);
                v10 = 268;
                goto LABEL_385;
              }
LABEL_418:
              v245 = v865;
              if ( v865 != nullptr && *v865 >= 0 )
              {
                v9 = (*(_QWORD *)v865)-- == 1;
                if ( v9 )
                {
                  v246 = sub_18000CC40(v245);
                  (*(void (__fastcall **)(int *))(v246 + 48))(v245);
                }
              }
              if ( *v129 >= 0 )
              {
                v9 = (*(_QWORD *)v129)-- == 1;
                if ( v9 )
                {
                  v247 = sub_18000CC40(v129);
                  (*(void (__fastcall **)(int *))(v247 + 48))(v129);
                }
              }
              v94 = v876;
LABEL_426:
              v8 = v865;
              v918 = v884;
              v917 = v883;
              v916 = v882;
              v915 = v865;
              v922 = v7;
              v923 = v6;
              v248 = sub_18015F270(a1);
              if ( v248 == 0 )
              {
// Original IDA line 2203 | qword_1803B7980 -> novels.pipeline.short_fiction_runner#144 bytes [4005539, 4005579]: "validate_short_fiction_draft_for_final"
                sub_180373340(a1, &v862, qword_1803B7980);
                v10 = 295;
                goto LABEL_1391;
              }
              if ( v6 == nullptr )
              {
// Original IDA line 2209 | qword_1803B7948 -> novels.pipeline.short_fiction_runner#137 bytes [4005331, 4005341]: "draft_v1"
                v249 = qword_1803B7948;
LABEL_430:
                sub_180370FE0(a1, &v862, v249);
                sub_18000D1D0(a1, v862);
                v10 = 295;
                goto LABEL_1391;
              }
              if ( v94 == nullptr )
              {
// Original IDA line 2218 | qword_1803B7938 -> novels.pipeline.short_fiction_runner#135 bytes [4005310, 4005320]: "chapters"
                v249 = qword_1803B7938;
                goto LABEL_430;
              }
              v4 = v866;
              v10 = 295;
              *(_DWORD *)(v866 + 40) = 295;
              v932 = v6;
              v931 = v94;
// Original IDA line 2226 | qword_1803B78B0 -> novels.pipeline.short_fiction_runner#118 bytes [4004987, 4005008]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004989, 4005008], "value": "expected_chapters"}]}
              v250 = (int *)sub_18037D030(a1, v248, &v932, &v931, qword_1803B78B0);
              v251 = v250;
              if ( v250 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                goto LABEL_1392;
              }
              if ( *v250 >= 0 )
              {
                v9 = (*(_QWORD *)v250)-- == 1;
                if ( v9 )
                {
                  v252 = sub_18000CC40(v250);
                  (*(void (__fastcall **)(int *))(v252 + 48))(v251);
                }
              }
              v253 = sub_18015D300(a1);
              if ( v253 == 0 )
              {
// Original IDA line 2245 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                sub_180373340(a1, &v862, qword_1803B7610);
                v10 = 296;
                goto LABEL_1392;
              }
// Original IDA line 2249 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
              v254 = (int *)sub_18037F0C0(a1, v253, qword_1803B78C0);
              if ( v254 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 296;
                goto LABEL_1392;
              }
// Original IDA line 2256 | qword_1803B7728 -> novels.pipeline.short_fiction_runner#69 bytes [4004173, 4004181]: "drafts"
              v255 = sub_180382540(v871, qword_1803B7728);
              v256 = (int *)v255;
              if ( v255 == 0 )
              {
                sub_18000CE30(a1, &v862);
                if ( *v254 >= 0 )
                {
                  v9 = (*(_QWORD *)v254)-- == 1;
                  if ( v9 )
                  {
                    v257 = sub_18000CC40(v254);
                    (*(void (__fastcall **)(int *))(v257 + 48))(v254);
                  }
                }
                v8 = v865;
                v10 = 296;
                goto LABEL_1392;
              }
// Original IDA line 2274 | qword_1803B7988 -> novels.pipeline.short_fiction_runner#145 bytes [4005579, 4005588]: "v001.md"
              v258 = (int *)sub_180382540(v255, qword_1803B7988);
              if ( *v256 >= 0 )
              {
                v9 = (*(_QWORD *)v256)-- == 1;
                if ( v9 )
                {
                  v259 = sub_18000CC40(v256);
                  (*(void (__fastcall **)(int *))(v259 + 48))(v256);
                }
              }
              if ( v258 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                if ( *v254 >= 0 )
                {
                  v9 = (*(_QWORD *)v254)-- == 1;
                  if ( v9 )
                  {
                    v260 = sub_18000CC40(v254);
                    (*(void (__fastcall **)(int *))(v260 + 48))(v254);
                  }
                }
                goto LABEL_454;
              }
              v261 = sub_18015F160(a1);
              if ( v261 == 0 )
              {
// Original IDA line 2301 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
                sub_180373340(a1, &v862, qword_1803B78D0);
                if ( *v254 >= 0 )
                {
                  v9 = (*(_QWORD *)v254)-- == 1;
                  if ( v9 )
                  {
                    v262 = sub_18000CC40(v254);
                    (*(void (__fastcall **)(int *))(v262 + 48))(v254);
                  }
                }
                if ( *v258 >= 0 )
                {
                  v9 = (*(_QWORD *)v258)-- == 1;
                  if ( v9 )
                  {
                    v263 = sub_18000CC40(v258);
                    (*(void (__fastcall **)(int *))(v263 + 48))(v258);
                    v7 = v864;
                    v10 = 296;
                    v8 = v865;
                    goto LABEL_1392;
                  }
                }
LABEL_454:
                v7 = v864;
                v10 = 296;
                v8 = v865;
                goto LABEL_1392;
              }
              v10 = 296;
              *(_DWORD *)(v4 + 40) = 296;
              v264 = sub_180376770(a1, v261, v6);
              v265 = (int *)v264;
              if ( v264 == 0 )
              {
                sub_18000CE30(a1, &v862);
                if ( *v254 >= 0 )
                {
                  v9 = (*(_QWORD *)v254)-- == 1;
                  if ( v9 )
                  {
                    v266 = sub_18000CC40(v254);
                    (*(void (__fastcall **)(int *))(v266 + 48))(v254);
                  }
                }
                if ( *v258 >= 0 )
                {
                  v9 = (*(_QWORD *)v258)-- == 1;
                  if ( v9 )
                  {
                    v267 = sub_18000CC40(v258);
                    (*(void (__fastcall **)(int *))(v267 + 48))(v258);
                    v7 = v864;
                    v8 = v865;
                    goto LABEL_1392;
                  }
                }
LABEL_205:
                v7 = v864;
                v8 = v865;
                goto LABEL_1392;
              }
              *(_DWORD *)(v4 + 40) = 296;
              v945[0] = v258;
              v945[1] = v264;
              v268 = (int *)sub_180377750(a1, v254, v945);
              if ( *v254 >= 0 )
              {
                v9 = (*(_QWORD *)v254)-- == 1;
                if ( v9 )
                {
                  v269 = sub_18000CC40(v254);
                  (*(void (__fastcall **)(int *))(v269 + 48))(v254);
                }
              }
              if ( *v258 >= 0 )
              {
                v9 = (*(_QWORD *)v258)-- == 1;
                if ( v9 )
                {
                  v270 = sub_18000CC40(v258);
                  (*(void (__fastcall **)(int *))(v270 + 48))(v258);
                }
              }
              if ( *v265 >= 0 )
              {
                v9 = (*(_QWORD *)v265)-- == 1;
                if ( v9 )
                {
                  v271 = sub_18000CC40(v265);
                  (*(void (__fastcall **)(int *))(v271 + 48))(v265);
                }
              }
              if ( v268 == nullptr )
                goto LABEL_479;
              if ( *v268 >= 0 )
              {
                v9 = (*(_QWORD *)v268)-- == 1;
                if ( v9 )
                {
                  v272 = sub_18000CC40(v268);
                  (*(void (__fastcall **)(int *))(v272 + 48))(v268);
                }
              }
              v273 = v875;
              if ( v875 == nullptr )
              {
// Original IDA line 2408 | qword_1803B75A8 -> novels.pipeline.short_fiction_runner#21 bytes [4003611, 4003617]: "self"
                sub_180370FE0(a1, &v862, qword_1803B75A8);
                sub_18000D1D0(a1, v862);
                v10 = 297;
                goto LABEL_1389;
              }
              v4 = v866;
              v10 = 297;
              v274 = v875;
              *(_DWORD *)(v866 + 40) = 297;
// Original IDA line 2417 | qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"; qword_1803B7990 -> novels.pipeline.short_fiction_runner#146 bytes [4005588, 4005603]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4005590, 4005592], "value": 3}, {"tag": "a", "fileRange": [4005592, 4005603], "value": "completed"}]}
              v275 = (int *)sub_18037E820(a1, v274, *(_QWORD *)&qword_1803B7828, qword_1803B7990 + 24);
              v276 = v275;
              if ( v275 == nullptr )
                goto LABEL_829;
              if ( *v275 >= 0 )
              {
                v9 = (*(_QWORD *)v275)-- == 1;
                if ( v9 )
                {
                  v277 = sub_18000CC40(v275);
                  (*(void (__fastcall **)(int *))(v277 + 48))(v276);
                }
              }
              v10 = 302;
              *(_DWORD *)(v4 + 40) = 302;
// Original IDA line 2432 | qword_1803B7998 -> novels.pipeline.short_fiction_runner#147 bytes [4005603, 4005629]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4005605, 4005607], "value": 4}, {"tag": "a", "fileRange": [4005607, 4005615], "value": "active"}, {"tag": "u", "fileRange": [4005615, 4005629], "value": "审阅正文"}]}; qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"
              v278 = (int *)sub_18037E980(a1, v273, *(_QWORD *)&qword_1803B7828, qword_1803B7998 + 24);
              v279 = v278;
              if ( v278 == nullptr )
                goto LABEL_829;
              if ( *v278 >= 0 )
              {
                v9 = (*(_QWORD *)v278)-- == 1;
                if ( v9 )
                {
                  v280 = sub_18000CC40(v278);
                  (*(void (__fastcall **)(int *))(v280 + 48))(v279);
                }
              }
              v281 = sub_18015F160(a1);
              if ( v281 == 0 )
              {
// Original IDA line 2448 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
                sub_180373340(a1, &v862, qword_1803B78D0);
                v7 = v864;
                v10 = 303;
                v8 = v865;
                goto LABEL_1392;
              }
              v10 = 303;
              *(_DWORD *)(v4 + 40) = 303;
              v282 = sub_180376770(a1, v281, v6);
              v919 = (int *)v282;
              v283 = v282;
              if ( v282 == 0 )
                goto LABEL_829;
              v900 = (int *)v282;
              v284 = sub_18015D960(a1);
              if ( v284 == 0 )
              {
// Original IDA line 2465 | qword_1803B79A0 -> novels.pipeline.short_fiction_runner#148 bytes [4005629, 4005665]: "ShortFictionDraftReviewPromptInput"
                sub_180373340(a1, &v862, qword_1803B79A0);
                v7 = v864;
                v10 = 304;
                v8 = v865;
                goto LABEL_1392;
              }
              v285 = v881;
              if ( v881 == nullptr )
              {
// Original IDA line 2474 | qword_1803B7928 -> novels.pipeline.short_fiction_runner#133 bytes [4005285, 4005296]: "direction"
                sub_180370FE0(a1, &v862, qword_1803B7928);
                sub_18000D1D0(a1, v862);
                v7 = v864;
                v10 = 305;
                v8 = v865;
                goto LABEL_1392;
              }
              v286 = v880;
              if ( v880 == nullptr )
              {
// Original IDA line 2484 | qword_1803B7930 -> novels.pipeline.short_fiction_runner#134 bytes [4005296, 4005310]: "outline_text"
                sub_180370FE0(a1, &v862, qword_1803B7930);
                sub_18000D1D0(a1, v862);
                v7 = v864;
                v10 = 306;
                v8 = v865;
                goto LABEL_1392;
              }
              v287 = v878;
              if ( v878 == nullptr )
              {
// Original IDA line 2494 | qword_1803B7940 -> novels.pipeline.short_fiction_runner#136 bytes [4005320, 4005331]: "chars_per"
                sub_180370FE0(a1, &v862, qword_1803B7940);
                sub_18000D1D0(a1, v862);
                v7 = v864;
                v10 = 308;
                v8 = v865;
                goto LABEL_1392;
              }
              v10 = 304;
              *(_DWORD *)(v4 + 40) = 304;
              v966[1] = v286;
              v966[2] = v876;
              v966[0] = v285;
              v966[3] = v287;
              v966[4] = v283;
// Original IDA line 2508 | qword_1803B79A8 -> novels.pipeline.short_fiction_runner#149 bytes [4005665, 4005746]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4005667, 4005678], "value": "direction"}, {"tag": "a", "fileRange": [4005678, 4005696], "value": "outline_markdown"}, {"tag": "a", "fileRange": [4005696, 4005711], "value": "chapter_cou... (truncated literal)
              v288 = sub_18037CCB0(a1, v284, v966, qword_1803B79A8);
              v903 = (_QWORD *)v288;
              if ( v288 == 0 )
              {
LABEL_829:
                sub_18000CE30(a1, &v862);
                v7 = v864;
                v8 = v865;
                goto LABEL_1392;
              }
              v894 = (int *)v288;
// Original IDA line 2519 | qword_1803B7850 -> novels.pipeline.short_fiction_runner#106 bytes [4004772, 4004779]: "_chat"
              v289 = (int *)sub_18037F0C0(a1, v273, *(_QWORD *)&qword_1803B7850);
              if ( v289 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 311;
                goto LABEL_1389;
              }
// Original IDA line 2526 | qword_1803B79B0 -> novels.pipeline.short_fiction_runner#150 bytes [4005746, 4005755]: "auditor"
              v290 = (_DWORD *)qword_1803B79B0;
              v291 = sub_180374A40(a1, 2);
              v292 = (int *)v291;
              if ( *v290 != -1 )
                ++*v290;
              sub_18000CC80(v291, 0, v290);
// Original IDA line 2532 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v293 = qword_1803B7860;
// Original IDA line 2533 | qword_1803B7868 -> novels.pipeline.short_fiction_runner#109 bytes [4004793, 4004801]: "system"
              v294 = qword_1803B7868;
              v295 = (int *)PyDict_NewPresized(2);
              PyDict_SetItem(v295, v293, v294);
// Original IDA line 2536 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              v296 = qword_1803B7870;
              v297 = sub_18015E170(a1);
              if ( v297 == 0 )
              {
// Original IDA line 2540 | qword_1803B79B8 -> novels.pipeline.short_fiction_runner#151 bytes [4005755, 4005803]: "build_short_fiction_draft_review_system_prompt"
                sub_180373340(a1, &v862, qword_1803B79B8);
                v10 = 314;
                goto LABEL_515;
              }
              v10 = 314;
              *(_DWORD *)(v866 + 40) = 314;
              v298 = sub_180376210(a1, v297);
              v299 = (int *)v298;
              if ( v298 == 0 )
              {
                sub_18000CE30(a1, &v862);
LABEL_515:
                if ( *v295 >= 0 )
                {
                  v9 = (*(_QWORD *)v295)-- == 1;
                  if ( !v9 )
                    goto LABEL_532;
                  v300 = sub_18000CC40(v295);
                  v301 = v295;
LABEL_531:
                  (*(void (__fastcall **)(int *))(v300 + 48))(v301);
                }
LABEL_532:
                if ( *v289 >= 0 )
                {
                  v9 = (*(_QWORD *)v289)-- == 1;
                  if ( v9 )
                  {
                    v313 = sub_18000CC40(v289);
                    (*(void (__fastcall **)(int *))(v313 + 48))(v289);
                  }
                }
                if ( *v292 >= 0 )
                {
                  v9 = (*(_QWORD *)v292)-- == 1;
                  if ( v9 )
                  {
                    v314 = sub_18000CC40(v292);
                    (*(void (__fastcall **)(int *))(v314 + 48))(v292);
                  }
                }
                goto LABEL_1388;
              }
              PyDict_SetItem(v295, v296, v298);
              if ( *v299 >= 0 )
              {
                v9 = (*(_QWORD *)v299)-- == 1;
                if ( v9 )
                {
                  v302 = sub_18000CC40(v299);
                  (*(void (__fastcall **)(int *))(v302 + 48))(v299);
                }
              }
              v873 = (int *)sub_180372160(a1, 2);
              sub_180024750(v873, 0, v295);
// Original IDA line 2595 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v303 = qword_1803B7860;
// Original IDA line 2596 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
              v304 = qword_1803B7878;
              v305 = (int *)PyDict_NewPresized(2);
              PyDict_SetItem(v305, v303, v304);
// Original IDA line 2599 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              v306 = qword_1803B7870;
              v307 = sub_18015E280(a1);
              if ( v307 == 0 )
              {
// Original IDA line 2603 | qword_1803B79C0 -> novels.pipeline.short_fiction_runner#152 bytes [4005803, 4005849]: "build_short_fiction_draft_review_user_prompt"
                sub_180373340(a1, &v862, qword_1803B79C0);
                v10 = 315;
                goto LABEL_525;
              }
              v10 = 315;
              v308 = v903;
              *(_DWORD *)(v866 + 40) = 315;
              v309 = sub_180376770(a1, v307, v308);
              v310 = (int *)v309;
              if ( v309 == 0 )
              {
                sub_18000CE30(a1, &v862);
LABEL_525:
                if ( *v305 >= 0 )
                {
                  v9 = (*(_QWORD *)v305)-- == 1;
                  if ( v9 )
                  {
                    v311 = sub_18000CC40(v305);
                    (*(void (__fastcall **)(int *))(v311 + 48))(v305);
                  }
                }
                v312 = v873;
                if ( *v873 < 0 )
                  goto LABEL_532;
                v9 = (*(_QWORD *)v873)-- == 1;
                if ( !v9 )
                  goto LABEL_532;
                v300 = sub_18000CC40(v312);
                v301 = v312;
                goto LABEL_531;
              }
              PyDict_SetItem(v305, v306, v309);
              if ( *v310 >= 0 )
              {
                v9 = (*(_QWORD *)v310)-- == 1;
                if ( v9 )
                {
                  v315 = sub_18000CC40(v310);
                  (*(void (__fastcall **)(int *))(v315 + 48))(v310);
                }
              }
              sub_180024750(v873, 1, v305);
              sub_18000CC80(v292, 1, v873);
// Original IDA line 2647 | qword_1803B79C8 -> novels.pipeline.short_fiction_runner#153 bytes [4005849, 4005915]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4005851, 4005864], "value": "stage_label"}, {"tag": "a", "fileRange": [4005864, 4005876], "value": "max_tokens"}, {"tag": "a", "fileRange": [4005876, 4005889], "value": "temperature"}], "... (truncated literal)
              v316 = (int *)sub_18036FB80(a1, qword_1803B79C8);
              v10 = 311;
              *(_DWORD *)(v866 + 40) = 311;
              v317 = sub_180001790(a1, v289, v292, v316);
              v22 = *v289 < 0;
              v318 = (int *)v317;
              v904 = (int *)v317;
              if ( !v22 )
              {
                v9 = (*(_QWORD *)v289)-- == 1;
                if ( v9 )
                {
                  v319 = sub_18000CC40(v289);
                  (*(void (__fastcall **)(int *))(v319 + 48))(v289);
                }
              }
              if ( *v292 >= 0 )
              {
                v9 = (*(_QWORD *)v292)-- == 1;
                if ( v9 )
                {
                  v320 = sub_18000CC40(v292);
                  (*(void (__fastcall **)(int *))(v320 + 48))(v292);
                }
              }
              if ( *v316 >= 0 )
              {
                v9 = (*(_QWORD *)v316)-- == 1;
                if ( v9 )
                {
                  v321 = sub_18000CC40(v316);
                  (*(void (__fastcall **)(int *))(v321 + 48))(v316);
                }
              }
              if ( v318 == nullptr )
                goto LABEL_1178;
              v898 = v318;
              v322 = sub_18015D300(a1);
              if ( v322 == 0 )
              {
// Original IDA line 2687 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                sub_180373340(a1, &v862, qword_1803B7610);
                v10 = 321;
LABEL_1388:
                v6 = v867;
                goto LABEL_1389;
              }
// Original IDA line 2693 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
              v323 = (int *)sub_18037F0C0(a1, v322, qword_1803B78C0);
              if ( v323 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 321;
                goto LABEL_1388;
              }
// Original IDA line 2700 | qword_1803B7720 -> novels.pipeline.short_fiction_runner#68 bytes [4004164, 4004173]: "reviews"
              v324 = sub_180382540(v871, qword_1803B7720);
              v325 = (int *)v324;
              if ( v324 == 0 )
              {
                v326 = a1;
LABEL_558:
                sub_18000CE30(v326, &v862);
                if ( *v323 >= 0 )
                {
                  v9 = (*(_QWORD *)v323)-- == 1;
                  if ( v9 )
                  {
                    v327 = sub_18000CC40(v323);
                    (*(void (__fastcall **)(int *))(v327 + 48))(v323);
                  }
                }
                v10 = 321;
                goto LABEL_1388;
              }
// Original IDA line 2719 | qword_1803B79D0 -> novels.pipeline.short_fiction_runner#154 bytes [4005915, 4005930]: "draft-v001.md"
              v328 = (int *)sub_180382540(v324, qword_1803B79D0);
              if ( *v325 >= 0 )
              {
                v9 = (*(_QWORD *)v325)-- == 1;
                if ( v9 )
                {
                  v329 = sub_18000CC40(v325);
                  (*(void (__fastcall **)(int *))(v329 + 48))(v325);
                }
              }
              v326 = a1;
              if ( v328 == nullptr )
                goto LABEL_558;
              v4 = v866;
              v10 = 321;
              *(_DWORD *)(v866 + 40) = 321;
              v946[0] = v328;
              v946[1] = v318;
              v330 = (int *)sub_180377750(a1, v323, v946);
              if ( *v323 >= 0 )
              {
                v9 = (*(_QWORD *)v323)-- == 1;
                if ( v9 )
                {
                  v331 = sub_18000CC40(v323);
                  (*(void (__fastcall **)(int *))(v331 + 48))(v323);
                }
              }
              if ( *v328 >= 0 )
              {
                v9 = (*(_QWORD *)v328)-- == 1;
                if ( v9 )
                {
                  v332 = sub_18000CC40(v328);
                  (*(void (__fastcall **)(int *))(v332 + 48))(v328);
                }
              }
              if ( v330 == nullptr )
                goto LABEL_1153;
              if ( *v330 >= 0 )
              {
                v9 = (*(_QWORD *)v330)-- == 1;
                if ( v9 )
                {
                  v333 = sub_18000CC40(v330);
                  (*(void (__fastcall **)(int *))(v333 + 48))(v330);
                }
              }
              v334 = v875;
              v10 = 322;
              *(_DWORD *)(v4 + 40) = 322;
// Original IDA line 2770 | qword_1803B79D8 -> novels.pipeline.short_fiction_runner#155 bytes [4005930, 4005945]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4005932, 4005934], "value": 4}, {"tag": "a", "fileRange": [4005934, 4005945], "value": "completed"}]}; qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"
              v335 = (int *)sub_18037E820(a1, v334, *(_QWORD *)&qword_1803B7828, qword_1803B79D8 + 24);
              v336 = v335;
              if ( v335 == nullptr )
                goto LABEL_1153;
              if ( *v335 >= 0 )
              {
                v9 = (*(_QWORD *)v335)-- == 1;
                if ( v9 )
                {
                  v337 = sub_18000CC40(v335);
                  (*(void (__fastcall **)(int *))(v337 + 48))(v336);
                }
              }
              v10 = 327;
              *(_DWORD *)(v4 + 40) = 327;
// Original IDA line 2785 | qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"; qword_1803B79E0 -> novels.pipeline.short_fiction_runner#156 bytes [4005945, 4005971]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4005947, 4005949], "value": 5}, {"tag": "a", "fileRange": [4005949, 4005957], "value": "active"}, {"tag": "u", "fileRange": [4005957, 4005971], "value": "修订正文"}]}
              v338 = (int *)sub_18037E980(a1, v334, *(_QWORD *)&qword_1803B7828, qword_1803B79E0 + 24);
              v339 = v338;
              if ( v338 == nullptr )
              {
LABEL_1153:
                sub_18000CE30(a1, &v862);
                v6 = v867;
                v7 = v864;
                v8 = v865;
                goto LABEL_1392;
              }
              if ( *v338 >= 0 )
              {
                v9 = (*(_QWORD *)v338)-- == 1;
                if ( v9 )
                {
                  v340 = sub_18000CC40(v338);
                  (*(void (__fastcall **)(int *))(v340 + 48))(v339);
                }
              }
              v341 = (int *)v867;
              if ( *(_DWORD *)v867 != -1 )
                ++*(_DWORD *)v867;
              v342 = (int *)Py_NoneStruct;
              v877 = v341;
              v872 = Py_NoneStruct;
              if ( Py_NoneStruct != -1 )
                ++Py_NoneStruct;
              v886 = (int *)Py_NoneStruct;
              v343 = sub_18015DA70(a1);
              if ( v343 == 0 )
              {
// Original IDA line 2817 | qword_1803B79E8 -> novels.pipeline.short_fiction_runner#157 bytes [4005971, 4006009]: "ShortFictionDraftRevisionPromptInput"
                sub_180373340(a1, &v862, qword_1803B79E8);
                v6 = v867;
                v10 = 331;
                v7 = v864;
                v8 = v865;
                goto LABEL_1392;
              }
              v344 = v881;
              v10 = 331;
              *(_DWORD *)(v4 + 40) = 331;
              v967[0] = v344;
              v967[1] = v880;
              v967[2] = v876;
              v967[3] = v878;
              v967[4] = v904;
// Original IDA line 2832 | qword_1803B79F0 -> novels.pipeline.short_fiction_runner#158 bytes [4006009, 4006082]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4006011, 4006022], "value": "direction"}, {"tag": "a", "fileRange": [4006022, 4006040], "value": "outline_markdown"}, {"tag": "a", "fileRange": [4006040, 4006055], "value": "chapter_cou... (truncated literal)
              v345 = sub_18037CCB0(a1, v343, v967, qword_1803B79F0);
              v873 = (int *)v345;
              v346 = a1;
              if ( v345 == 0 )
              {
LABEL_593:
                sub_18000CE30(v346, &v862);
                v6 = v867;
                v7 = v864;
                v8 = v865;
                goto LABEL_1392;
              }
              v899 = (int *)v345;
// Original IDA line 2845 | qword_1803B7850 -> novels.pipeline.short_fiction_runner#106 bytes [4004772, 4004779]: "_chat"
              v347 = (int *)sub_18037F0C0(a1, v334, *(_QWORD *)&qword_1803B7850);
              if ( v347 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v348 = 339;
LABEL_719:
                v412 = v871;
                goto LABEL_720;
              }
// Original IDA line 2854 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v349 = qword_1803B7860;
// Original IDA line 2855 | qword_1803B7868 -> novels.pipeline.short_fiction_runner#109 bytes [4004793, 4004801]: "system"
              v350 = qword_1803B7868;
// Original IDA line 2856 | qword_1803B7858 -> novels.pipeline.short_fiction_runner#107 bytes [4004779, 4004787]: "writer"
              v902 = qword_1803B7858;
              v351 = PyDict_NewPresized(2);
              PyDict_SetItem(v351, v349, v350);
// Original IDA line 2859 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              PyDict_SetItem(v351, qword_1803B7870, v908);
              v352 = (int *)sub_180372160(a1, 4);
              sub_180024750(v352, 0, v351);
// Original IDA line 2862 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v353 = qword_1803B7860;
// Original IDA line 2863 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
              v354 = qword_1803B7878;
              v355 = (int *)PyDict_NewPresized(2);
              PyDict_SetItem(v355, v353, v354);
// Original IDA line 2866 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              v356 = qword_1803B7870;
              v357 = sub_18015E9F0(a1);
              if ( v357 == 0 )
              {
// Original IDA line 2870 | qword_1803B7880 -> novels.pipeline.short_fiction_runner#112 bytes [4004816, 4004856]: "build_short_fiction_writer_user_prompt"
                sub_180373340(a1, &v862, qword_1803B7880);
                v348 = 343;
                goto LABEL_634;
              }
              v348 = 343;
              v358 = v909;
              *(_DWORD *)(v866 + 40) = 343;
              v359 = sub_180376770(a1, v357, v358);
              v360 = (int *)v359;
              if ( v359 != 0 )
              {
                PyDict_SetItem(v355, v356, v359);
                if ( *v360 >= 0 )
                {
                  v9 = (*(_QWORD *)v360)-- == 1;
                  if ( v9 )
                  {
                    v361 = sub_18000CC40(v360);
                    (*(void (__fastcall **)(int *))(v361 + 48))(v360);
                  }
                }
                sub_180024750(v352, 1, v355);
// Original IDA line 2892 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
                v362 = qword_1803B7860;
// Original IDA line 2893 | qword_1803B79F8 -> novels.pipeline.short_fiction_runner#159 bytes [4006082, 4006093]: "assistant"
                v363 = qword_1803B79F8;
                v364 = (int *)PyDict_NewPresized(2);
                PyDict_SetItem(v364, v362, v363);
// Original IDA line 2896 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
                v905 = qword_1803B7870;
// Original IDA line 2897 | qword_1803B7970 -> novels.pipeline.short_fiction_runner#142 bytes [4005522, 4005535]: "raw_content"
                v365 = sub_18037F0C0(a1, v867, qword_1803B7970);
                v366 = (int *)v365;
                if ( v365 == 0 )
                  goto LABEL_603;
// Original IDA line 2901 | qword_1803B7620 -> novels.pipeline.short_fiction_runner#36 bytes [4003841, 4003848]: "strip"
                v367 = (int *)sub_18037F0C0(a1, v365, qword_1803B7620);
                if ( *v366 >= 0 )
                {
                  v9 = (*(_QWORD *)v366)-- == 1;
                  if ( v9 )
                  {
                    v368 = sub_18000CC40(v366);
                    (*(void (__fastcall **)(int *))(v368 + 48))(v366);
                  }
                }
                if ( v367 == nullptr )
                {
LABEL_603:
                  sub_18000CE30(a1, &v862);
                  v348 = 344;
                  goto LABEL_617;
                }
                v348 = 344;
                *(_DWORD *)(v866 + 40) = 344;
                v369 = (int *)sub_180376210(a1, v367);
                if ( *v367 >= 0 )
                {
                  v9 = (*(_QWORD *)v367)-- == 1;
                  if ( v9 )
                  {
                    v370 = sub_18000CC40(v367);
                    (*(void (__fastcall **)(int *))(v370 + 48))(v367);
                  }
                }
                if ( v369 == nullptr )
                {
                  sub_18000CE30(a1, &v862);
LABEL_617:
                  if ( *v364 < 0 )
                    goto LABEL_638;
                  v9 = (*(_QWORD *)v364)-- == 1;
                  if ( !v9 )
                    goto LABEL_638;
                  v373 = sub_18000CC40(v364);
                  v374 = v364;
LABEL_637:
                  (*(void (__fastcall **)(int *))(v373 + 48))(v374);
                  goto LABEL_638;
                }
                v371 = sub_180001510(v369);
                if ( v371 == -1 )
                {
                  sub_18000CE30(a1, &v862);
                  if ( *v369 >= 0 )
                  {
                    v9 = (*(_QWORD *)v369)-- == 1;
                    if ( v9 )
                    {
                      v372 = sub_18000CC40(v369);
                      (*(void (__fastcall **)(int *))(v372 + 48))(v369);
                    }
                  }
                  goto LABEL_617;
                }
                if ( v371 != 1 )
                {
                  if ( *v369 >= 0 )
                  {
                    v9 = (*(_QWORD *)v369)-- == 1;
                    if ( v9 )
                    {
                      v375 = sub_18000CC40(v369);
                      (*(void (__fastcall **)(int *))(v375 + 48))(v369);
                    }
                  }
                  v376 = v919;
                  if ( *v919 != -1 )
                    ++*v919;
                  v369 = v376;
                }
                PyDict_SetItem(v364, v905, v369);
                if ( *v369 >= 0 )
                {
                  v9 = (*(_QWORD *)v369)-- == 1;
                  if ( v9 )
                  {
                    v377 = sub_18000CC40(v369);
                    (*(void (__fastcall **)(int *))(v377 + 48))(v369);
                  }
                }
                sub_180024750(v352, 2, v364);
// Original IDA line 2987 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
                v378 = qword_1803B7860;
// Original IDA line 2988 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
                v379 = qword_1803B7878;
                v355 = (int *)PyDict_NewPresized(2);
                PyDict_SetItem(v355, v378, v379);
// Original IDA line 2991 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
                v380 = qword_1803B7870;
                v381 = sub_18015E390(a1);
                if ( v381 == 0 )
                {
// Original IDA line 2995 | qword_1803B7A00 -> novels.pipeline.short_fiction_runner#160 bytes [4006093, 4006138]: "build_short_fiction_draft_revision_followup"
                  sub_180373340(a1, &v862, qword_1803B7A00);
                  v348 = 345;
                  goto LABEL_634;
                }
                v348 = 345;
                v382 = v873;
                *(_DWORD *)(v866 + 40) = 345;
                v383 = sub_180376770(a1, v381, v382);
                v384 = (int *)v383;
                if ( v383 != 0 )
                {
                  PyDict_SetItem(v355, v380, v383);
                  if ( *v384 >= 0 )
                  {
                    v9 = (*(_QWORD *)v384)-- == 1;
                    if ( v9 )
                    {
                      v387 = sub_18000CC40(v384);
                      (*(void (__fastcall **)(int *))(v387 + 48))(v384);
                    }
                  }
                  sub_180024750(v352, 3, v355);
// Original IDA line 3017 | qword_1803B7A08 -> novels.pipeline.short_fiction_runner#161 bytes [4006138, 4006152]: "修订正文"
                  v388 = qword_1803B7A08;
                  v389 = sub_18015EB00(a1);
                  if ( v389 == 0 )
                  {
// Original IDA line 3021 | qword_1803B7890 -> novels.pipeline.short_fiction_runner#114 bytes [4004870, 4004905]: "estimate_short_fiction_max_tokens"
                    sub_180373340(a1, &v862, qword_1803B7890);
                    if ( *v347 >= 0 )
                    {
                      v9 = (*(_QWORD *)v347)-- == 1;
                      if ( v9 )
                      {
                        v390 = sub_18000CC40(v347);
                        (*(void (__fastcall **)(int *))(v390 + 48))(v347);
                      }
                    }
                    if ( *v352 >= 0 )
                    {
                      v9 = (*(_QWORD *)v352)-- == 1;
                      if ( v9 )
                      {
                        v391 = sub_18000CC40(v352);
                        (*(void (__fastcall **)(int *))(v391 + 48))(v352);
                      }
                    }
                    v348 = 348;
                    goto LABEL_717;
                  }
                  v392 = v876;
                  v393 = v866;
                  v348 = 348;
                  *(_DWORD *)(v866 + 40) = 348;
                  v947[0] = v392;
                  v947[1] = v878;
                  v394 = (int *)sub_180377750(a1, v389, v947);
                  if ( v394 == nullptr )
                  {
                    sub_18000CE30(a1, &v862);
LABEL_638:
                    if ( *v347 >= 0 )
                    {
                      v9 = (*(_QWORD *)v347)-- == 1;
                      if ( v9 )
                      {
                        v385 = sub_18000CC40(v347);
                        (*(void (__fastcall **)(int *))(v385 + 48))(v347);
                      }
                    }
                    if ( *v352 >= 0 )
                    {
                      v9 = (*(_QWORD *)v352)-- == 1;
                      if ( v9 )
                      {
                        v386 = sub_18000CC40(v352);
                        (*(void (__fastcall **)(int *))(v386 + 48))(v352);
                      }
                    }
                    goto LABEL_717;
                  }
// Original IDA line 3074 | qword_1803B7A10 -> novels.pipeline.short_fiction_runner#162 bytes [4006152, 4006161]: 0.45
                  v395 = qword_1803B7A10;
                  v396 = v902;
                  v348 = 339;
                  *(_DWORD *)(v393 + 40) = 339;
                  v960[2] = v395;
                  v948[0] = v396;
                  v948[1] = v352;
                  v960[0] = v388;
                  v960[1] = v394;
// Original IDA line 3083 | qword_1803B78A0 -> novels.pipeline.short_fiction_runner#116 bytes [4004914, 4004954]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004916, 4004929], "value": "stage_label"}, {"tag": "a", "fileRange": [4004929, 4004941], "value": "max_tokens"}, {"tag": "a", "fileRange": [4004941, 4004954], "value": "temperature"}]}
                  v397 = sub_18037D650(a1, v347, v948, v960, qword_1803B78A0);
                  if ( *v347 >= 0 )
                  {
                    v9 = (*(_QWORD *)v347)-- == 1;
                    if ( v9 )
                    {
                      v398 = sub_18000CC40(v347);
                      (*(void (__fastcall **)(int *))(v398 + 48))(v347);
                    }
                  }
                  if ( *v352 >= 0 )
                  {
                    v9 = (*(_QWORD *)v352)-- == 1;
                    if ( v9 )
                    {
                      v399 = sub_18000CC40(v352);
                      (*(void (__fastcall **)(int *))(v399 + 48))(v352);
                    }
                  }
                  if ( *v394 >= 0 )
                  {
                    v9 = (*(_QWORD *)v394)-- == 1;
                    if ( v9 )
                    {
                      v400 = sub_18000CC40(v394);
                      (*(void (__fastcall **)(int *))(v400 + 48))(v394);
                    }
                  }
                  v401 = a1;
                  if ( v397 != 0 )
                  {
                    v887 = (int *)v397;
                    v402 = sub_18015EE30(a1);
                    if ( v402 == 0 )
                    {
// Original IDA line 3118 | qword_1803B78A8 -> novels.pipeline.short_fiction_runner#117 bytes [4004954, 4004987]: "parse_short_fiction_batch_draft"
                      sub_180373340(a1, &v862, qword_1803B78A8);
                      v348 = 351;
LABEL_717:
                      v4 = v866;
                      goto LABEL_718;
                    }
                    v403 = v876;
                    v348 = 351;
                    *(_DWORD *)(v393 + 40) = 351;
                    v4 = v393;
                    v934 = v397;
                    v933 = v403;
// Original IDA line 3130 | qword_1803B78B0 -> novels.pipeline.short_fiction_runner#118 bytes [4004987, 4005008]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004989, 4005008], "value": "expected_chapters"}]}
                    v404 = sub_18037D030(a1, v402, &v934, &v933, qword_1803B78B0);
                    v405 = (_DWORD *)v404;
                    if ( v404 == 0 )
                      goto LABEL_670;
                    v885 = (int *)v404;
                    v406 = sub_18015F270(a1);
                    if ( v406 == 0 )
                    {
// Original IDA line 3138 | qword_1803B7980 -> novels.pipeline.short_fiction_runner#144 bytes [4005539, 4005579]: "validate_short_fiction_draft_for_final"
                      sub_180373340(a1, &v862, qword_1803B7980);
                      v348 = 352;
                      goto LABEL_718;
                    }
                    v348 = 352;
                    *(_DWORD *)(v4 + 40) = 352;
                    v936 = v405;
                    v935 = v403;
// Original IDA line 3146 | qword_1803B78B0 -> novels.pipeline.short_fiction_runner#118 bytes [4004987, 4005008]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004989, 4005008], "value": "expected_chapters"}]}
                    v407 = (int *)sub_18037D030(a1, v406, &v936, &v935, qword_1803B78B0);
                    v408 = v407;
                    if ( v407 == nullptr )
                    {
LABEL_670:
                      sub_18000CE30(a1, &v862);
LABEL_718:
                      v342 = (int *)v872;
                      goto LABEL_719;
                    }
                    if ( *v407 >= 0 )
                    {
                      v9 = (*(_QWORD *)v407)-- == 1;
                      if ( v9 )
                      {
                        v409 = sub_18000CC40(v407);
                        (*(void (__fastcall **)(int *))(v409 + 48))(v408);
                      }
                    }
                    v410 = sub_18015D300(a1);
                    if ( v410 == 0 )
                    {
// Original IDA line 3168 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                      sub_180373340(a1, &v862, qword_1803B7610);
                      v348 = 353;
                      goto LABEL_718;
                    }
// Original IDA line 3172 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
                    v411 = (int *)sub_18037F0C0(a1, v410, qword_1803B78C0);
                    if ( v411 == nullptr )
                    {
                      sub_18000CE30(a1, &v862);
                      v348 = 353;
                      goto LABEL_718;
                    }
                    v412 = v871;
// Original IDA line 3180 | qword_1803B7728 -> novels.pipeline.short_fiction_runner#69 bytes [4004173, 4004181]: "drafts"
                    v413 = sub_180382540(v871, qword_1803B7728);
                    v414 = (int *)v413;
                    if ( v413 == 0 )
                    {
                      v415 = a1;
LABEL_683:
                      sub_18000CE30(v415, &v862);
                      if ( *v411 >= 0 )
                      {
                        v9 = (*(_QWORD *)v411)-- == 1;
                        if ( v9 )
                        {
                          v416 = sub_18000CC40(v411);
                          (*(void (__fastcall **)(int *))(v416 + 48))(v411);
                        }
                      }
                      goto LABEL_686;
                    }
// Original IDA line 3198 | qword_1803B7788 -> novels.pipeline.short_fiction_runner#81 bytes [4004291, 4004300]: "v002.md"
                    v417 = (int *)sub_180382540(v413, qword_1803B7788);
                    if ( *v414 >= 0 )
                    {
                      v9 = (*(_QWORD *)v414)-- == 1;
                      if ( v9 )
                      {
                        v418 = sub_18000CC40(v414);
                        (*(void (__fastcall **)(int *))(v418 + 48))(v414);
                      }
                    }
                    v415 = a1;
                    if ( v417 == nullptr )
                      goto LABEL_683;
                    v419 = sub_18015F160(a1);
                    if ( v419 == 0 )
                    {
// Original IDA line 3214 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
                      sub_180373340(a1, &v862, qword_1803B78D0);
                      if ( *v411 >= 0 )
                      {
                        v9 = (*(_QWORD *)v411)-- == 1;
                        if ( v9 )
                        {
                          v420 = sub_18000CC40(v411);
                          (*(void (__fastcall **)(int *))(v420 + 48))(v411);
                        }
                      }
                      if ( *v417 >= 0 )
                      {
                        v9 = (*(_QWORD *)v417)-- == 1;
                        if ( v9 )
                        {
                          v421 = sub_18000CC40(v417);
                          (*(void (__fastcall **)(int *))(v421 + 48))(v417);
                          v342 = (int *)v872;
                          v348 = 353;
                          goto LABEL_720;
                        }
                      }
LABEL_686:
                      v342 = (int *)v872;
                      v348 = 353;
LABEL_720:
                      v907 = v862;
                      sub_18000CE70(&v862);
                      v925 = sub_18000CCC0(a1);
                      v430 = sub_18000CE80(&v907);
                      if ( v430 != 0 )
                        v431 = sub_18000CC90(v430, v4, v348);
                      else
                        v431 = sub_18038E2A0(v4, v348);
                      sub_180001300(&v907, v431);
                      sub_18000D050(a1, &v907);
                      v432 = sub_18000CC40(*(_QWORD *)(a1 + 248));
                      v433 = (_DWORD *)sub_18015DB80(a1);
                      if ( v433 == nullptr )
                      {
// Original IDA line 3254 | qword_1803B7A18 -> novels.pipeline.short_fiction_runner#163 bytes [4006161, 4006195]: "ShortFictionDraftValidationError"
                        sub_180373340(a1, &v862, qword_1803B7A18);
                        v10 = 355;
LABEL_815:
                        v478 = v862;
                        sub_18000CE70(&v862);
                        sub_18000CCE0(a1, &v925);
                        v862 = v478;
                        goto LABEL_1388;
                      }
                      v434 = sub_180374A40(a1, 2);
                      v435 = (int *)v434;
                      if ( *v433 != -1 )
                        ++*v433;
                      sub_18000CC80(v434, 0, v433);
                      if ( *PyExc_Exception != -1 )
                        ++*PyExc_Exception;
                      sub_18000CC80(v435, 1, PyExc_Exception);
                      v436 = sub_18000CE90(a1, v432, v435);
                      if ( *v435 >= 0 )
                      {
                        v9 = (*(_QWORD *)v435)-- == 1;
                        if ( v9 )
                        {
                          v437 = sub_18000CC40(v435);
                          (*(void (__fastcall **)(int *))(v437 + 48))(v435);
                        }
                      }
                      if ( v436 == -1 )
                      {
                        sub_18000CE30(a1, &v862);
                        v10 = 355;
                        goto LABEL_815;
                      }
                      if ( v436 == 0 )
                      {
                        v10 = 338;
                        if ( (unsigned __int8)sub_180373290(a1, &v862) != 0 )
                          v10 = 0;
                        v477 = sub_18000CE80(&v862);
                        if ( v477 != 0 && *(_QWORD *)(v477 + 24) == v4 )
                          *(_DWORD *)(v4 + 40) = *(_DWORD *)(v477 + 36);
                        goto LABEL_815;
                      }
                      v438 = *(_DWORD **)(a1 + 248);
                      if ( *v438 != -1 )
                        ++*v438;
                      v439 = sub_1803749A0(v438);
                      v879 = (_DWORD *)v439;
                      if ( v439 == 0 )
                      {
                        sub_18000CE30(a1, &v862);
                        v10 = 356;
                        goto LABEL_797;
                      }
                      v22 = *v342 < 0;
                      v886 = (int *)v439;
                      if ( !v22 )
                      {
                        v9 = (*(_QWORD *)v342)-- == 1;
                        if ( v9 )
                        {
                          v440 = sub_18000CC40(v342);
                          (*(void (__fastcall **)(int *))(v440 + 48))(v342);
                        }
                      }
                      v441 = sub_18015D300(a1);
                      if ( v441 == 0 )
                      {
// Original IDA line 3322 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                        sub_180373340(a1, &v862, qword_1803B7610);
                        v10 = 357;
                        goto LABEL_797;
                      }
// Original IDA line 3326 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
                      v442 = (int *)sub_18037F0C0(a1, v441, qword_1803B78C0);
                      if ( v442 == nullptr )
                      {
                        sub_18000CE30(a1, &v862);
                        v10 = 357;
                        goto LABEL_797;
                      }
// Original IDA line 3333 | qword_1803B7720 -> novels.pipeline.short_fiction_runner#68 bytes [4004164, 4004173]: "reviews"
                      v443 = sub_180382540(v412, qword_1803B7720);
                      v444 = (int *)v443;
                      if ( v443 == 0 )
                      {
                        v445 = a1;
LABEL_748:
                        sub_18000CE30(v445, &v862);
                        if ( *v442 >= 0 )
                        {
                          v9 = (*(_QWORD *)v442)-- == 1;
                          if ( v9 )
                          {
                            v446 = sub_18000CC40(v442);
                            (*(void (__fastcall **)(int *))(v446 + 48))(v442);
                          }
                        }
LABEL_751:
                        v10 = 357;
                        goto LABEL_797;
                      }
// Original IDA line 3353 | qword_1803B7A20 -> novels.pipeline.short_fiction_runner#164 bytes [4006195, 4006218]: "draft-v002-warning.md"
                      v447 = (int *)sub_180382540(v443, qword_1803B7A20);
                      if ( *v444 >= 0 )
                      {
                        v9 = (*(_QWORD *)v444)-- == 1;
                        if ( v9 )
                        {
                          v448 = sub_18000CC40(v444);
                          (*(void (__fastcall **)(int *))(v448 + 48))(v444);
                        }
                      }
                      v445 = a1;
                      if ( v447 == nullptr )
                        goto LABEL_748;
// Original IDA line 3366 | qword_1803B7A28 -> novels.pipeline.short_fiction_runner#165 bytes [4006218, 4006220]: "\n"
                      v449 = qword_1803B7A28;
// Original IDA line 3367 | qword_1803B7A30 -> novels.pipeline.short_fiction_runner#166 bytes [4006220, 4006248]: "# 第二轮改稿未采用"
                      v450 = (_DWORD *)qword_1803B7A30;
                      v451 = sub_180372160(a1, 7);
                      v452 = (int *)v451;
                      if ( *v450 != -1 )
                        ++*v450;
                      sub_180024750(v451, 0, v450);
// Original IDA line 3373 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      v453 = qword_1803B7598;
// Original IDA line 3374 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      if ( *(_DWORD *)qword_1803B7598 != -1 )
// Original IDA line 3375 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                        ++*(_DWORD *)qword_1803B7598;
                      sub_180024750(v452, 1, v453);
// Original IDA line 3377 | qword_1803B7A38 -> novels.pipeline.short_fiction_runner#167 bytes [4006248, 4006319]: "系统没有用不完整或解析失败的改稿覆盖完整首稿。"
                      v454 = qword_1803B7A38;
// Original IDA line 3378 | qword_1803B7A38 -> novels.pipeline.short_fiction_runner#167 bytes [4006248, 4006319]: "系统没有用不完整或解析失败的改稿覆盖完整首稿。"
                      if ( *(_DWORD *)qword_1803B7A38 != -1 )
// Original IDA line 3379 | qword_1803B7A38 -> novels.pipeline.short_fiction_runner#167 bytes [4006248, 4006319]: "系统没有用不完整或解析失败的改稿覆盖完整首稿。"
                        ++*(_DWORD *)qword_1803B7A38;
                      sub_180024750(v452, 2, v454);
// Original IDA line 3381 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      v455 = qword_1803B7598;
// Original IDA line 3382 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      if ( *(_DWORD *)qword_1803B7598 != -1 )
// Original IDA line 3383 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                        ++*(_DWORD *)qword_1803B7598;
                      sub_180024750(v452, 3, v455);
// Original IDA line 3385 | qword_1803B7A40 -> novels.pipeline.short_fiction_runner#168 bytes [4006319, 4006330]: "## 原因"
                      v456 = qword_1803B7A40;
// Original IDA line 3386 | qword_1803B7A40 -> novels.pipeline.short_fiction_runner#168 bytes [4006319, 4006330]: "## 原因"
                      if ( *(_DWORD *)qword_1803B7A40 != -1 )
// Original IDA line 3387 | qword_1803B7A40 -> novels.pipeline.short_fiction_runner#168 bytes [4006319, 4006330]: "## 原因"
                        ++*(_DWORD *)qword_1803B7A40;
                      sub_180024750(v452, 4, v456);
// Original IDA line 3389 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      v457 = qword_1803B7598;
// Original IDA line 3390 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                      if ( *(_DWORD *)qword_1803B7598 != -1 )
// Original IDA line 3391 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                        ++*(_DWORD *)qword_1803B7598;
                      sub_180024750(v452, 5, v457);
                      v458 = v879;
                      if ( *v879 != -1 )
                        ++*v879;
                      sub_180024750(v452, 6, v458);
                      v459 = (int *)sub_180374990(a1, v449, v452);
                      if ( *v452 >= 0 )
                      {
                        v9 = (*(_QWORD *)v452)-- == 1;
                        if ( v9 )
                        {
                          v460 = sub_18000CC40(v452);
                          (*(void (__fastcall **)(int *))(v460 + 48))(v452);
                        }
                      }
                      if ( v459 == nullptr )
                      {
                        sub_18000CE30(a1, &v862);
                        if ( *v442 >= 0 )
                        {
                          v9 = (*(_QWORD *)v442)-- == 1;
                          if ( v9 )
                          {
                            v461 = sub_18000CC40(v442);
                            (*(void (__fastcall **)(int *))(v461 + 48))(v442);
                          }
                        }
                        if ( *v447 < 0 )
                          goto LABEL_751;
                        v9 = (*(_QWORD *)v447)-- == 1;
                        if ( !v9 )
                          goto LABEL_751;
                        v462 = sub_18000CC40(v447);
                        (*(void (__fastcall **)(int *))(v462 + 48))(v447);
                        v10 = 357;
LABEL_797:
                        v472 = v862;
                        sub_18000CE70(&v862);
                        if ( v438 != nullptr && (int)*v438 >= 0 )
                        {
                          v9 = (*(_QWORD *)v438)-- == 1;
                          if ( v9 )
                          {
                            v473 = sub_18000CC40(v438);
                            (*(void (__fastcall **)(_DWORD *))(v473 + 48))(v438);
                          }
                        }
                        v862 = v472;
                        goto LABEL_815;
                      }
                      v10 = 357;
                      *(_DWORD *)(v866 + 40) = 357;
                      v950[0] = v447;
                      v950[1] = v459;
                      v463 = (int *)sub_180377750(a1, v442, v950);
                      if ( *v442 >= 0 )
                      {
                        v9 = (*(_QWORD *)v442)-- == 1;
                        if ( v9 )
                        {
                          v464 = sub_18000CC40(v442);
                          (*(void (__fastcall **)(int *))(v464 + 48))(v442);
                        }
                      }
                      if ( *v447 >= 0 )
                      {
                        v9 = (*(_QWORD *)v447)-- == 1;
                        if ( v9 )
                        {
                          v465 = sub_18000CC40(v447);
                          (*(void (__fastcall **)(int *))(v465 + 48))(v447);
                        }
                      }
                      if ( *v459 >= 0 )
                      {
                        v9 = (*(_QWORD *)v459)-- == 1;
                        if ( v9 )
                        {
                          v466 = sub_18000CC40(v459);
                          (*(void (__fastcall **)(int *))(v466 + 48))(v459);
                        }
                      }
                      if ( v463 != nullptr )
                      {
                        if ( *v463 >= 0 )
                        {
                          v9 = (*(_QWORD *)v463)-- == 1;
                          if ( v9 )
                          {
                            v467 = sub_18000CC40(v463);
                            (*(void (__fastcall **)(int *))(v467 + 48))(v463);
                          }
                        }
                        v468 = sub_18015ED20(a1);
                        if ( v468 == 0 )
                        {
// Original IDA line 3488 | qword_1803B7800 -> novels.pipeline.short_fiction_runner#96 bytes [4004504, 4004512]: "logger"
                          sub_180373340(a1, &v862, qword_1803B7800);
                          v10 = 366;
                          goto LABEL_797;
                        }
// Original IDA line 3492 | qword_1803B7A50 -> novels.pipeline.short_fiction_runner#170 bytes [4006339, 4006381]: "Draft revision skipped, keeping v001: %s"
                        v469 = qword_1803B7A50;
                        v10 = 366;
                        *(_DWORD *)(v866 + 40) = 366;
                        v951[0] = v469;
                        v951[1] = v879;
// Original IDA line 3497 | qword_1803B7A48 -> novels.pipeline.short_fiction_runner#169 bytes [4006330, 4006339]: "warning"
                        v470 = (int *)sub_18037E820(a1, v468, qword_1803B7A48, v951);
                        v471 = v470;
                        if ( v470 != nullptr )
                        {
                          if ( *v470 >= 0 )
                          {
                            v9 = (*(_QWORD *)v470)-- == 1;
                            if ( v9 )
                            {
                              v474 = sub_18000CC40(v470);
                              (*(void (__fastcall **)(int *))(v474 + 48))(v471);
                            }
                          }
                          if ( v438 != nullptr && (int)*v438 >= 0 )
                          {
                            v9 = (*(_QWORD *)v438)-- == 1;
                            if ( v9 )
                            {
                              v475 = sub_18000CC40(v438);
                              (*(void (__fastcall **)(_DWORD *))(v475 + 48))(v438);
                            }
                          }
                          sub_18000CCE0(a1, &v925);
                          v6 = v867;
                          v476 = (int *)v867;
LABEL_824:
                          v10 = 368;
                          v4 = v866;
                          v483 = v875;
                          v905 = (__int64)v885;
                          v484 = v875;
                          v485 = v887;
                          *(_DWORD *)(v866 + 40) = 368;
                          v902 = (__int64)v485;
                          v914 = v476;
// Original IDA line 3532 | qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"; qword_1803B7A58 -> novels.pipeline.short_fiction_runner#171 bytes [4006381, 4006396]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4006383, 4006385], "value": 5}, {"tag": "a", "fileRange": [4006385, 4006396], "value": "completed"}]}
                          v486 = (int *)sub_18037E820(a1, v484, *(_QWORD *)&qword_1803B7828, qword_1803B7A58 + 24);
                          v487 = v486;
                          if ( v486 == nullptr )
                            goto LABEL_829;
                          if ( *v486 >= 0 )
                          {
                            v9 = (*(_QWORD *)v486)-- == 1;
                            if ( v9 )
                            {
                              v488 = sub_18000CC40(v486);
                              (*(void (__fastcall **)(int *))(v488 + 48))(v487);
                            }
                          }
                          v10 = 373;
                          *(_DWORD *)(v4 + 40) = 373;
// Original IDA line 3547 | qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"; qword_1803B7A60 -> novels.pipeline.short_fiction_runner#172 bytes [4006396, 4006422]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4006398, 4006400], "value": 6}, {"tag": "a", "fileRange": [4006400, 4006408], "value": "active"}, {"tag": "u", "fileRange": [4006408, 4006422], "value": "打包输出"}]}
                          v489 = (int *)sub_18037E980(a1, v483, *(_QWORD *)&qword_1803B7828, qword_1803B7A60 + 24);
                          v490 = v489;
                          if ( v489 == nullptr )
                            goto LABEL_829;
                          if ( *v489 >= 0 )
                          {
                            v9 = (*(_QWORD *)v489)-- == 1;
                            if ( v9 )
                            {
                              v491 = sub_18000CC40(v489);
                              (*(void (__fastcall **)(int *))(v491 + 48))(v490);
                            }
                          }
                          v492 = sub_18015F160(a1);
                          if ( v492 == 0 )
                          {
// Original IDA line 3563 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
                            sub_180373340(a1, &v862, qword_1803B78D0);
                            v7 = v864;
                            v10 = 374;
                            v8 = v865;
                            goto LABEL_1392;
                          }
                          v10 = 374;
                          *(_DWORD *)(v4 + 40) = 374;
                          v493 = sub_180376770(a1, v492, v476);
                          v939 = (int *)v493;
                          v494 = v493;
                          if ( v493 == 0 )
                          {
LABEL_479:
                            sub_18000CE30(a1, &v862);
LABEL_1389:
                            v7 = v864;
LABEL_1390:
                            v8 = v865;
                            goto LABEL_1391;
                          }
                          v901 = (int *)v493;
                          v495 = sub_18015D300(a1);
                          if ( v495 == 0 )
                          {
// Original IDA line 3588 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                            sub_180373340(a1, &v862, qword_1803B7610);
                            v10 = 375;
                            goto LABEL_1389;
                          }
// Original IDA line 3592 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
                          v496 = (int *)sub_18037F0C0(a1, v495, qword_1803B78C0);
                          if ( v496 == nullptr )
                          {
                            sub_18000CE30(a1, &v862);
                            v10 = 375;
                            goto LABEL_1389;
                          }
// Original IDA line 3599 | qword_1803B7730 -> novels.pipeline.short_fiction_runner#70 bytes [4004181, 4004188]: "final"
                          v497 = sub_180382540(v871, qword_1803B7730);
                          v498 = (int *)v497;
                          if ( v497 == 0 )
                          {
                            v499 = a1;
LABEL_842:
                            sub_18000CE30(v499, &v862);
                            if ( *v496 >= 0 )
                            {
                              v9 = (*(_QWORD *)v496)-- == 1;
                              if ( v9 )
                              {
                                v500 = sub_18000CC40(v496);
                                (*(void (__fastcall **)(int *))(v500 + 48))(v496);
                              }
                            }
                            v10 = 375;
                            goto LABEL_1389;
                          }
// Original IDA line 3618 | qword_1803B7738 -> novels.pipeline.short_fiction_runner#71 bytes [4004188, 4004197]: "full.md"
                          v501 = (int *)sub_180382540(v497, qword_1803B7738);
                          if ( *v498 >= 0 )
                          {
                            v9 = (*(_QWORD *)v498)-- == 1;
                            if ( v9 )
                            {
                              v502 = sub_18000CC40(v498);
                              (*(void (__fastcall **)(int *))(v502 + 48))(v498);
                            }
                          }
                          v499 = a1;
                          if ( v501 == nullptr )
                            goto LABEL_842;
                          v111 = v866;
                          v10 = 375;
                          *(_DWORD *)(v866 + 40) = 375;
                          v952[0] = v501;
                          v952[1] = v494;
                          v503 = (int *)sub_180377750(a1, v496, v952);
                          if ( *v496 >= 0 )
                          {
                            v9 = (*(_QWORD *)v496)-- == 1;
                            if ( v9 )
                            {
                              v504 = sub_18000CC40(v496);
                              (*(void (__fastcall **)(int *))(v504 + 48))(v496);
                            }
                          }
                          if ( *v501 >= 0 )
                          {
                            v9 = (*(_QWORD *)v501)-- == 1;
                            if ( v9 )
                            {
                              v505 = sub_18000CC40(v501);
                              (*(void (__fastcall **)(int *))(v505 + 48))(v501);
                            }
                          }
                          if ( v503 != nullptr )
                          {
                            if ( *v503 >= 0 )
                            {
                              v9 = (*(_QWORD *)v503)-- == 1;
                              if ( v9 )
                              {
                                v506 = sub_18000CC40(v503);
                                (*(void (__fastcall **)(int *))(v506 + 48))(v503);
                              }
                            }
                            v507 = sub_18015DDA0(a1);
                            if ( v507 == 0 )
                            {
// Original IDA line 3669 | qword_1803B7A68 -> novels.pipeline.short_fiction_runner#173 bytes [4006422, 4006454]: "ShortFictionPackagePromptInput"
                              sub_180373340(a1, &v862, qword_1803B7A68);
                              v7 = v864;
                              v4 = v111;
                              v6 = v867;
                              v10 = 377;
                              v8 = v865;
                              goto LABEL_1392;
                            }
// Original IDA line 3677 | qword_1803B7A70 -> novels.pipeline.short_fiction_runner#174 bytes [4006454, 4006467]: "story_title"
                            v508 = (int *)sub_18037F0C0(a1, v877, qword_1803B7A70);
                            if ( v508 == nullptr )
                            {
                              sub_18000CE30(a1, &v862);
                              v7 = v864;
                              v4 = v111;
                              v6 = v867;
                              v10 = 381;
                              v8 = v865;
                              goto LABEL_1392;
                            }
                            v509 = v881;
                            v10 = 377;
                            *(_DWORD *)(v111 + 40) = 377;
                            v963[0] = v509;
                            v963[1] = v880;
                            v963[2] = v494;
                            v963[3] = v508;
// Original IDA line 3695 | qword_1803B7A78 -> novels.pipeline.short_fiction_runner#175 bytes [4006467, 4006527]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4006469, 4006480], "value": "direction"}, {"tag": "a", "fileRange": [4006480, 4006498], "value": "outline_markdown"}, {"tag": "a", "fileRange": [4006498, 4006514], "value": "draft_markd... (truncated literal)
                            v510 = sub_18037CCB0(a1, v507, v963, qword_1803B7A78);
                            v22 = *v508 < 0;
                            v879 = (_DWORD *)v510;
                            if ( !v22 )
                            {
                              v9 = (*(_QWORD *)v508)-- == 1;
                              if ( v9 )
                              {
                                v511 = sub_18000CC40(v508);
                                (*(void (__fastcall **)(int *))(v511 + 48))(v508);
                                v510 = (__int64)v879;
                              }
                            }
                            v120 = a1;
                            if ( v510 == 0 )
                            {
LABEL_198:
                              sub_18000CE30(v120, &v862);
                              v7 = v864;
                              v4 = v111;
                              v6 = v867;
                              v8 = v865;
                              goto LABEL_1392;
                            }
                            v889 = (int *)v510;
// Original IDA line 3720 | qword_1803B7850 -> novels.pipeline.short_fiction_runner#106 bytes [4004772, 4004779]: "_chat"
                            v512 = (int *)sub_18037F0C0(a1, v875, *(_QWORD *)&qword_1803B7850);
                            if ( v512 == nullptr )
                            {
                              sub_18000CE30(a1, &v862);
                              v7 = v864;
                              v4 = v111;
                              v6 = v867;
                              v10 = 383;
                              v8 = v865;
                              goto LABEL_1392;
                            }
// Original IDA line 3731 | qword_1803B7A80 -> novels.pipeline.short_fiction_runner#176 bytes [4006527, 4006538]: "architect"
                            v513 = (_DWORD *)qword_1803B7A80;
                            v514 = sub_180374A40(a1, 2);
                            v515 = (int *)v514;
                            if ( *v513 != -1 )
                              ++*v513;
                            sub_18000CC80(v514, 0, v513);
// Original IDA line 3737 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
                            v516 = qword_1803B7860;
// Original IDA line 3738 | qword_1803B7868 -> novels.pipeline.short_fiction_runner#109 bytes [4004793, 4004801]: "system"
                            v517 = qword_1803B7868;
                            v518 = (int *)PyDict_NewPresized(2);
                            PyDict_SetItem(v518, v516, v517);
// Original IDA line 3741 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
                            v519 = qword_1803B7870;
                            v520 = sub_18015E7D0(a1);
                            if ( v520 == 0 )
                            {
// Original IDA line 3745 | qword_1803B7A88 -> novels.pipeline.short_fiction_runner#177 bytes [4006538, 4006581]: "build_short_fiction_package_system_prompt"
                              sub_180373340(a1, &v862, qword_1803B7A88);
                              v10 = 386;
                              goto LABEL_876;
                            }
                            v10 = 386;
                            *(_DWORD *)(v866 + 40) = 386;
                            v521 = sub_180376210(a1, v520);
                            v522 = (int *)v521;
                            if ( v521 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
LABEL_876:
                              if ( *v518 < 0 || (v9 = *(_QWORD *)v518 == 1, --*(_QWORD *)v518, !v9) )
                              {
LABEL_893:
                                if ( *v512 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v512)-- == 1;
                                  if ( v9 )
                                  {
                                    v536 = sub_18000CC40(v512);
                                    (*(void (__fastcall **)(int *))(v536 + 48))(v512);
                                  }
                                }
LABEL_896:
                                if ( *v515 < 0 )
                                  goto LABEL_1388;
                                v9 = (*(_QWORD *)v515)-- == 1;
                                if ( !v9 )
                                  goto LABEL_1388;
                                goto LABEL_1387;
                              }
                              v523 = sub_18000CC40(v518);
                              v524 = v518;
LABEL_892:
                              (*(void (__fastcall **)(int *))(v523 + 48))(v524);
                              goto LABEL_893;
                            }
                            PyDict_SetItem(v518, v519, v521);
                            if ( *v522 >= 0 )
                            {
                              v9 = (*(_QWORD *)v522)-- == 1;
                              if ( v9 )
                              {
                                v525 = sub_18000CC40(v522);
                                (*(void (__fastcall **)(int *))(v525 + 48))(v522);
                              }
                            }
                            v868 = (int *)sub_180372160(a1, 2);
                            sub_180024750(v868, 0, v518);
// Original IDA line 3795 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
                            v526 = qword_1803B7860;
// Original IDA line 3796 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
                            v527 = qword_1803B7878;
                            v528 = (int *)PyDict_NewPresized(2);
                            PyDict_SetItem(v528, v526, v527);
// Original IDA line 3799 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
                            v529 = qword_1803B7870;
                            v530 = sub_18015E8E0(a1);
                            if ( v530 == 0 )
                            {
// Original IDA line 3803 | qword_1803B7A90 -> novels.pipeline.short_fiction_runner#178 bytes [4006581, 4006622]: "build_short_fiction_package_user_prompt"
                              sub_180373340(a1, &v862, qword_1803B7A90);
                              v10 = 387;
                              goto LABEL_886;
                            }
                            v10 = 387;
                            v531 = v879;
                            *(_DWORD *)(v866 + 40) = 387;
                            v532 = sub_180376770(a1, v530, v531);
                            v533 = (int *)v532;
                            if ( v532 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
LABEL_886:
                              if ( *v528 >= 0 )
                              {
                                v9 = (*(_QWORD *)v528)-- == 1;
                                if ( v9 )
                                {
                                  v534 = sub_18000CC40(v528);
                                  (*(void (__fastcall **)(int *))(v534 + 48))(v528);
                                }
                              }
                              v535 = v868;
                              if ( *v868 < 0 )
                                goto LABEL_893;
                              v9 = (*(_QWORD *)v868)-- == 1;
                              if ( !v9 )
                                goto LABEL_893;
                              v523 = sub_18000CC40(v535);
                              v524 = v535;
                              goto LABEL_892;
                            }
                            PyDict_SetItem(v528, v529, v532);
                            if ( *v533 >= 0 )
                            {
                              v9 = (*(_QWORD *)v533)-- == 1;
                              if ( v9 )
                              {
                                v537 = sub_18000CC40(v533);
                                (*(void (__fastcall **)(int *))(v537 + 48))(v533);
                              }
                            }
                            sub_180024750(v868, 1, v528);
                            sub_18000CC80(v515, 1, v868);
// Original IDA line 3847 | qword_1803B7A98 -> novels.pipeline.short_fiction_runner#179 bytes [4006622, 4006688]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4006624, 4006637], "value": "stage_label"}, {"tag": "a", "fileRange": [4006637, 4006649], "value": "max_tokens"}, {"tag": "a", "fileRange": [4006649, 4006662], "value": "temperature"}], "... (truncated literal)
                            v538 = (int *)sub_18036FB80(a1, qword_1803B7A98);
                            v10 = 383;
                            *(_DWORD *)(v866 + 40) = 383;
                            v539 = sub_180001790(a1, v512, v515, v538);
                            v22 = *v512 < 0;
                            v540 = (int *)v539;
                            v863 = (int *)v539;
                            if ( !v22 )
                            {
                              v9 = (*(_QWORD *)v512)-- == 1;
                              if ( v9 )
                              {
                                v541 = sub_18000CC40(v512);
                                (*(void (__fastcall **)(int *))(v541 + 48))(v512);
                              }
                            }
                            if ( *v515 >= 0 )
                            {
                              v9 = (*(_QWORD *)v515)-- == 1;
                              if ( v9 )
                              {
                                v542 = sub_18000CC40(v515);
                                (*(void (__fastcall **)(int *))(v542 + 48))(v515);
                              }
                            }
                            if ( *v538 >= 0 )
                            {
                              v9 = (*(_QWORD *)v538)-- == 1;
                              if ( v9 )
                              {
                                v543 = sub_18000CC40(v538);
                                (*(void (__fastcall **)(int *))(v543 + 48))(v538);
                              }
                            }
                            if ( v540 == nullptr )
                            {
LABEL_1178:
                              sub_18000CE30(a1, &v862);
                              goto LABEL_1388;
                            }
                            v890 = v540;
                            v544 = sub_18015F050(a1);
                            if ( v544 == 0 )
                            {
// Original IDA line 3891 | qword_1803B7AA0 -> novels.pipeline.short_fiction_runner#180 bytes [4006688, 4006723]: "parse_short_fiction_sales_package"
                              sub_180373340(a1, &v862, qword_1803B7AA0);
                              v10 = 393;
                              goto LABEL_1388;
                            }
// Original IDA line 3895 | qword_1803B7A70 -> novels.pipeline.short_fiction_runner#174 bytes [4006454, 4006467]: "story_title"
                            v545 = sub_18037F0C0(a1, v877, qword_1803B7A70);
                            v546 = (int *)v545;
                            if ( v545 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              v10 = 393;
                              goto LABEL_1388;
                            }
                            v4 = v866;
                            v10 = 393;
                            *(_DWORD *)(v866 + 40) = 393;
                            v953[0] = v540;
                            v953[1] = v545;
                            v547 = sub_180377750(a1, v544, v953);
                            v22 = *v546 < 0;
                            v548 = (int *)v547;
                            v868 = (int *)v547;
                            if ( !v22 )
                            {
                              v9 = (*(_QWORD *)v546)-- == 1;
                              if ( v9 )
                              {
                                v549 = sub_18000CC40(v546);
                                (*(void (__fastcall **)(int *))(v549 + 48))(v546);
                              }
                            }
                            v346 = a1;
                            if ( v548 == nullptr )
                              goto LABEL_593;
                            v891 = v548;
                            v550 = sub_18015D300(a1);
                            if ( v550 == 0 )
                            {
// Original IDA line 3928 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                              sub_180373340(a1, &v862, qword_1803B7610);
                              v6 = v867;
                              v10 = 394;
                              v7 = v864;
                              v8 = v865;
                              goto LABEL_1392;
                            }
// Original IDA line 3935 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
                            v551 = (int *)sub_18037F0C0(a1, v550, qword_1803B78C0);
                            if ( v551 == nullptr )
                            {
                              sub_18000CE30(a1, &v862);
                              v6 = v867;
                              v10 = 394;
                              v7 = v864;
                              v8 = v865;
                              goto LABEL_1392;
                            }
// Original IDA line 3945 | qword_1803B7730 -> novels.pipeline.short_fiction_runner#70 bytes [4004181, 4004188]: "final"
                            v552 = sub_180382540(v871, qword_1803B7730);
                            v553 = (int *)v552;
                            if ( v552 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v554 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v554 + 48))(v551);
                                }
                              }
                              v6 = v867;
                              v10 = 394;
                              v7 = v864;
                              v8 = v865;
                              goto LABEL_1392;
                            }
// Original IDA line 3965 | qword_1803B7AA8 -> novels.pipeline.short_fiction_runner#181 bytes [4006723, 4006733]: "sales.md"
                            v555 = (int *)sub_180382540(v552, qword_1803B7AA8);
                            if ( *v553 >= 0 )
                            {
                              v9 = (*(_QWORD *)v553)-- == 1;
                              if ( v9 )
                              {
                                v556 = sub_18000CC40(v553);
                                (*(void (__fastcall **)(int *))(v556 + 48))(v553);
                              }
                            }
                            if ( v555 == nullptr )
                            {
                              sub_18000CE30(a1, &v862);
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v557 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v557 + 48))(v551);
                                }
                              }
LABEL_936:
                              v10 = 394;
                              goto LABEL_1388;
                            }
// Original IDA line 3991 | qword_1803B7AB0 -> novels.pipeline.short_fiction_runner#182 bytes [4006733, 4006737]: "# "
                            v558 = (_DWORD *)qword_1803B7AB0;
// Original IDA line 3992 | qword_1803B7A28 -> novels.pipeline.short_fiction_runner#165 bytes [4006218, 4006220]: "\n"
                            v870 = qword_1803B7A28;
                            v559 = sub_180374A40(a1, 2);
                            v560 = (int *)v559;
                            if ( *v558 != -1 )
                              ++*v558;
                            sub_18000CC80(v559, 0, v558);
// Original IDA line 3998 | qword_1803B7778 -> novels.pipeline.short_fiction_runner#79 bytes [4004271, 4004278]: "title"
                            v561 = sub_18037F0C0(a1, v548, qword_1803B7778);
                            v562 = (int *)v561;
                            if ( v561 == 0 )
                              goto LABEL_944;
// Original IDA line 4002 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                            v563 = sub_180375520(a1, v561, qword_1803B7598);
                            if ( *v562 >= 0 )
                            {
                              v9 = (*(_QWORD *)v562)-- == 1;
                              if ( v9 )
                              {
                                v564 = sub_18000CC40(v562);
                                (*(void (__fastcall **)(int *))(v564 + 48))(v562);
                              }
                            }
                            if ( v563 == 0 )
                            {
LABEL_944:
                              v10 = 395;
                              sub_18000CE30(a1, &v862);
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v565 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v565 + 48))(v551);
                                }
                              }
                              if ( *v555 >= 0 )
                              {
                                v9 = (*(_QWORD *)v555)-- == 1;
                                if ( v9 )
                                {
                                  v566 = sub_18000CC40(v555);
                                  (*(void (__fastcall **)(int *))(v566 + 48))(v555);
                                }
                              }
                              if ( *v560 < 0 )
                                goto LABEL_1388;
                              v9 = (*(_QWORD *)v560)-- == 1;
                              if ( !v9 )
                                goto LABEL_1388;
LABEL_952:
                              v567 = sub_18000CC40(v560);
                              (*(void (__fastcall **)(int *))(v567 + 48))(v560);
                              goto LABEL_1388;
                            }
                            sub_18000CC80(v560, 1, v563);
// Original IDA line 4046 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                            v568 = PyUnicode_Join(qword_1803B7598, v560);
                            if ( *v560 >= 0 )
                            {
                              v9 = (*(_QWORD *)v560)-- == 1;
                              if ( v9 )
                              {
                                v569 = sub_18000CC40(v560);
                                (*(void (__fastcall **)(int *))(v569 + 48))(v560);
                              }
                            }
                            if ( v568 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v570 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v570 + 48))(v551);
                                }
                              }
                              if ( *v555 >= 0 )
                              {
                                v9 = (*(_QWORD *)v555)-- == 1;
                                if ( v9 )
                                {
                                  v571 = sub_18000CC40(v555);
                                  (*(void (__fastcall **)(int *))(v571 + 48))(v555);
                                }
                              }
                              v10 = 395;
                              goto LABEL_1388;
                            }
                            v572 = sub_180374A40(a1, 1);
                            sub_18000CC80(v572, 0, v568);
                            v872 = sub_180374A40(a1, 13);
                            v515 = (int *)v872;
                            sub_18000CC80(v872, 0, v572);
// Original IDA line 4085 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v573 = qword_1803B7AB8;
// Original IDA line 4086 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4087 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 1, v573);
// Original IDA line 4089 | qword_1803B7AC0 -> novels.pipeline.short_fiction_runner#184 bytes [4006740, 4006753]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006742, 4006753], "value": "## 简介"}]}
                            v574 = qword_1803B7AC0;
// Original IDA line 4090 | qword_1803B7AC0 -> novels.pipeline.short_fiction_runner#184 bytes [4006740, 4006753]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006742, 4006753], "value": "## 简介"}]}
                            if ( *(_DWORD *)qword_1803B7AC0 != -1 )
// Original IDA line 4091 | qword_1803B7AC0 -> novels.pipeline.short_fiction_runner#184 bytes [4006740, 4006753]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006742, 4006753], "value": "## 简介"}]}
                              ++*(_DWORD *)qword_1803B7AC0;
                            sub_18000CC80(v515, 2, v574);
// Original IDA line 4093 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v575 = qword_1803B7AB8;
// Original IDA line 4094 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4095 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 3, v575);
// Original IDA line 4097 | qword_1803B7AC8 -> novels.pipeline.short_fiction_runner#185 bytes [4006753, 4006760]: "intro"
                            v576 = sub_18037F0C0(a1, v868, qword_1803B7AC8);
                            if ( v576 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              v10 = 399;
LABEL_1379:
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v796 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v796 + 48))(v551);
                                }
                              }
                              if ( *v555 >= 0 )
                              {
                                v9 = (*(_QWORD *)v555)-- == 1;
                                if ( v9 )
                                {
                                  v797 = sub_18000CC40(v555);
                                  (*(void (__fastcall **)(int *))(v797 + 48))(v555);
                                }
                              }
                              if ( *v515 < 0 )
                                goto LABEL_1388;
                              v9 = (*(_QWORD *)v515)-- == 1;
                              if ( !v9 )
                                goto LABEL_1388;
LABEL_1387:
                              v798 = sub_18000CC40(v515);
                              (*(void (__fastcall **)(int *))(v798 + 48))(v515);
                              goto LABEL_1388;
                            }
                            v577 = sub_180374A40(a1, 1);
                            sub_18000CC80(v577, 0, v576);
                            sub_18000CC80(v515, 4, v577);
// Original IDA line 4134 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v578 = qword_1803B7AB8;
// Original IDA line 4135 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4136 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 5, v578);
// Original IDA line 4138 | qword_1803B7AD0 -> novels.pipeline.short_fiction_runner#186 bytes [4006760, 4006773]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006762, 4006773], "value": "## 卖点"}]}
                            v579 = qword_1803B7AD0;
// Original IDA line 4139 | qword_1803B7AD0 -> novels.pipeline.short_fiction_runner#186 bytes [4006760, 4006773]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006762, 4006773], "value": "## 卖点"}]}
                            if ( *(_DWORD *)qword_1803B7AD0 != -1 )
// Original IDA line 4140 | qword_1803B7AD0 -> novels.pipeline.short_fiction_runner#186 bytes [4006760, 4006773]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006762, 4006773], "value": "## 卖点"}]}
                              ++*(_DWORD *)qword_1803B7AD0;
                            sub_18000CC80(v515, 6, v579);
// Original IDA line 4142 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v580 = qword_1803B7AB8;
// Original IDA line 4143 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4144 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 7, v580);
                            v581 = v868;
// Original IDA line 4147 | qword_1803B7AD8 -> novels.pipeline.short_fiction_runner#187 bytes [4006773, 4006789]: "selling_points"
                            v582 = sub_18037F0C0(a1, v868, qword_1803B7AD8);
                            v583 = (int *)v582;
                            if ( v582 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              v584 = v862;
                              sub_18000CE70(&v862);
LABEL_1378:
                              v862 = v584;
                              v10 = 403;
                              goto LABEL_1379;
                            }
                            v585 = (int *)sub_18000D280(a1, v582);
                            if ( *v583 >= 0 )
                            {
                              v9 = (*(_QWORD *)v583)-- == 1;
                              if ( v9 )
                              {
                                v586 = sub_18000CC40(v583);
                                (*(void (__fastcall **)(int *))(v586 + 48))(v583);
                              }
                            }
                            if ( v585 == nullptr )
                            {
                              sub_18000CE30(a1, &v862);
                              v584 = v862;
                              sub_18000CE70(&v862);
LABEL_1377:
                              v515 = (int *)v872;
                              goto LABEL_1378;
                            }
                            v869 = (_QWORD *)sub_180372160(a1, 0);
                            v587 = v869;
                            v588 = (int *)sub_18000D430(v585);
                            if ( v588 != nullptr )
                            {
                              while ( 1 )
                              {
                                v589 = v906;
                                v590 = v874;
                                v591 = v906;
                                v906 = v588;
                                v592 = v874;
                                if ( v591 != nullptr && *v589 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v589)-- == 1;
                                  if ( v9 )
                                  {
                                    v593 = sub_18000CC40(v589);
                                    (*(void (__fastcall **)(int *))(v593 + 48))(v591);
                                    v590 = v592;
                                  }
                                }
                                v594 = *v588;
                                v595 = v590;
                                v874 = v588;
                                v596 = v594 + 1;
                                if ( v596 != 0 )
                                  *v588 = v596;
                                if ( v590 != nullptr && *v592 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v592)-- == 1;
                                  if ( v9 )
                                  {
                                    v597 = sub_18000CC40(v590);
                                    (*(void (__fastcall **)(int *))(v597 + 48))(v595);
                                  }
                                }
// Original IDA line 4215 | qword_1803B7AE0 -> novels.pipeline.short_fiction_runner#188 bytes [4006789, 4006793]: "- "
                                v598 = (_DWORD *)qword_1803B7AE0;
                                v599 = sub_180374A40(a1, 2);
                                v600 = (int *)v599;
                                if ( *v598 != -1 )
                                  ++*v598;
                                sub_18000CC80(v599, 0, v598);
// Original IDA line 4221 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                                v601 = sub_180375520(a1, v588, qword_1803B7598);
                                if ( v601 == 0 )
                                  break;
                                sub_18000CC80(v600, 1, v601);
// Original IDA line 4225 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                                v602 = PyUnicode_Join(qword_1803B7598, v600);
                                if ( *v600 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v600)-- == 1;
                                  if ( v9 )
                                  {
                                    v603 = sub_18000CC40(v600);
                                    (*(void (__fastcall **)(int *))(v603 + 48))(v600);
                                  }
                                }
                                if ( v602 == 0 )
                                  goto LABEL_1361;
                                v587 = v869;
                                v604 = sub_180372560(v869, v602);
                                v605 = a1;
                                if ( v604 == 0 )
                                  goto LABEL_1362;
                                if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                  goto LABEL_1361;
                                v588 = (int *)sub_18000D430(v585);
                                if ( v588 == nullptr )
                                {
                                  v581 = v868;
                                  goto LABEL_1007;
                                }
                              }
                              sub_18000CE30(a1, &v862);
                              if ( *v600 >= 0 )
                              {
                                v9 = (*(_QWORD *)v600)-- == 1;
                                if ( v9 )
                                {
                                  v616 = sub_18000CC40(v600);
                                  (*(void (__fastcall **)(int *))(v616 + 48))(v600);
                                }
                              }
                              goto LABEL_1363;
                            }
LABEL_1007:
                            if ( (unsigned __int8)sub_18000CDA0(a1) == 0 )
                            {
LABEL_1361:
                              v605 = a1;
LABEL_1362:
                              sub_18000CE30(v605, &v862);
LABEL_1363:
                              v584 = v862;
                              sub_18000CE70(&v862);
                              if ( *v585 >= 0 )
                              {
                                v9 = (*(_QWORD *)v585)-- == 1;
                                if ( v9 )
                                {
                                  v790 = sub_18000CC40(v585);
                                  (*(void (__fastcall **)(int *))(v790 + 48))(v585);
                                }
                              }
                              v791 = v869;
                              if ( *(int *)v869 >= 0 )
                              {
                                v9 = (*v869)-- == 1;
                                if ( v9 )
                                {
                                  v792 = sub_18000CC40(v791);
                                  (*(void (__fastcall **)(_QWORD *))(v792 + 48))(v791);
                                }
                              }
                              v793 = v874;
                              if ( v874 != nullptr && *v874 >= 0 )
                              {
                                v9 = (*(_QWORD *)v874)-- == 1;
                                if ( v9 )
                                {
                                  v794 = sub_18000CC40(v793);
                                  (*(void (__fastcall **)(int *))(v794 + 48))(v793);
                                }
                              }
                              v862 = v584;
                              sub_18000CE70(&v862);
                              if ( v793 != nullptr && *v793 >= 0 )
                              {
                                v9 = (*(_QWORD *)v793)-- == 1;
                                if ( v9 )
                                {
                                  v795 = sub_18000CC40(v793);
                                  (*(void (__fastcall **)(int *))(v795 + 48))(v793);
                                }
                              }
                              goto LABEL_1377;
                            }
                            if ( *v587 != -1 )
                              ++*v587;
                            if ( *v585 >= 0 )
                            {
                              v9 = (*(_QWORD *)v585)-- == 1;
                              if ( v9 )
                              {
                                v606 = sub_18000CC40(v585);
                                (*(void (__fastcall **)(int *))(v606 + 48))(v585);
                              }
                            }
                            if ( (int)*v587 >= 0 )
                            {
                              v9 = (*(_QWORD *)v587)-- == 1;
                              if ( v9 )
                              {
                                v607 = sub_18000CC40(v587);
                                (*(void (__fastcall **)(_DWORD *))(v607 + 48))(v587);
                              }
                            }
                            v608 = v874;
                            if ( v874 != nullptr && *v874 >= 0 )
                            {
                              v9 = (*(_QWORD *)v874)-- == 1;
                              v609 = *(_QWORD *)v608;
                              if ( v9 )
                              {
                                v610 = sub_18000CC40(v608);
                                (*(void (__fastcall **)(int *))(v610 + 48))(v608);
                                LODWORD(v609) = *v608;
                              }
                              if ( (int)v609 >= 0 )
                              {
                                v9 = (*(_QWORD *)v608)-- == 1;
                                if ( v9 )
                                {
                                  v611 = sub_18000CC40(v608);
                                  (*(void (__fastcall **)(int *))(v611 + 48))(v608);
                                }
                              }
                            }
                            v515 = (int *)v872;
                            sub_18000CC80(v872, 8, v587);
// Original IDA line 4358 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v612 = qword_1803B7AB8;
// Original IDA line 4359 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4360 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 9, v612);
// Original IDA line 4362 | qword_1803B7AE8 -> novels.pipeline.short_fiction_runner#189 bytes [4006793, 4006815]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006795, 4006815], "value": "## 封面提示词"}]}
                            v613 = qword_1803B7AE8;
// Original IDA line 4363 | qword_1803B7AE8 -> novels.pipeline.short_fiction_runner#189 bytes [4006793, 4006815]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006795, 4006815], "value": "## 封面提示词"}]}
                            if ( *(_DWORD *)qword_1803B7AE8 != -1 )
// Original IDA line 4364 | qword_1803B7AE8 -> novels.pipeline.short_fiction_runner#189 bytes [4006793, 4006815]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006795, 4006815], "value": "## 封面提示词"}]}
                              ++*(_DWORD *)qword_1803B7AE8;
                            sub_18000CC80(v515, 10, v613);
// Original IDA line 4366 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            v614 = qword_1803B7AB8;
// Original IDA line 4367 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                            if ( *(_DWORD *)qword_1803B7AB8 != -1 )
// Original IDA line 4368 | qword_1803B7AB8 -> novels.pipeline.short_fiction_runner#183 bytes [4006737, 4006740]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [4006739, 4006740], "value": ""}]}
                              ++*(_DWORD *)qword_1803B7AB8;
                            sub_18000CC80(v515, 11, v614);
// Original IDA line 4370 | qword_1803B7AF0 -> novels.pipeline.short_fiction_runner#190 bytes [4006815, 4006829]: "cover_prompt"
                            v615 = sub_18037F0C0(a1, v581, qword_1803B7AF0);
                            if ( v615 == 0 )
                            {
                              sub_18000CE30(a1, &v862);
                              v10 = 407;
                              goto LABEL_1379;
                            }
                            v617 = sub_180374A40(a1, 1);
                            sub_18000CC80(v617, 0, v615);
                            sub_18000CC80(v515, 12, v617);
                            v937 = v515;
                            v618 = sub_180096B10(a1, &v937);
                            v619 = (int *)v618;
                            v620 = a1;
                            if ( v618 == 0 )
                              goto LABEL_1035;
                            v623 = (int *)sub_180374990(a1, v870, v618);
                            if ( *v619 >= 0 )
                            {
                              v9 = (*(_QWORD *)v619)-- == 1;
                              if ( v9 )
                              {
                                v624 = sub_18000CC40(v619);
                                (*(void (__fastcall **)(int *))(v624 + 48))(v619);
                              }
                            }
                            v620 = a1;
                            if ( v623 == nullptr )
                            {
LABEL_1035:
                              sub_18000CE30(v620, &v862);
                              if ( *v551 >= 0 )
                              {
                                v9 = (*(_QWORD *)v551)-- == 1;
                                if ( v9 )
                                {
                                  v621 = sub_18000CC40(v551);
                                  (*(void (__fastcall **)(int *))(v621 + 48))(v551);
                                }
                              }
                              if ( *v555 >= 0 )
                              {
                                v9 = (*(_QWORD *)v555)-- == 1;
                                if ( v9 )
                                {
                                  v622 = sub_18000CC40(v555);
                                  (*(void (__fastcall **)(int *))(v622 + 48))(v555);
                                  v10 = 394;
                                  goto LABEL_1388;
                                }
                              }
                              goto LABEL_936;
                            }
                            v111 = v866;
                            v10 = 394;
                            *(_DWORD *)(v866 + 40) = 394;
                            v954[0] = v555;
                            v954[1] = v623;
                            v625 = (int *)sub_180377750(a1, v551, v954);
                            if ( *v551 >= 0 )
                            {
                              v9 = (*(_QWORD *)v551)-- == 1;
                              if ( v9 )
                              {
                                v626 = sub_18000CC40(v551);
                                (*(void (__fastcall **)(int *))(v626 + 48))(v551);
                              }
                            }
                            if ( *v555 >= 0 )
                            {
                              v9 = (*(_QWORD *)v555)-- == 1;
                              if ( v9 )
                              {
                                v627 = sub_18000CC40(v555);
                                (*(void (__fastcall **)(int *))(v627 + 48))(v555);
                              }
                            }
                            if ( *v623 >= 0 )
                            {
                              v9 = (*(_QWORD *)v623)-- == 1;
                              if ( v9 )
                              {
                                v628 = sub_18000CC40(v623);
                                (*(void (__fastcall **)(int *))(v628 + 48))(v623);
                              }
                            }
                            if ( v625 != nullptr )
                            {
                              if ( *v625 >= 0 )
                              {
                                v9 = (*(_QWORD *)v625)-- == 1;
                                if ( v9 )
                                {
                                  v629 = sub_18000CC40(v625);
                                  (*(void (__fastcall **)(int *))(v629 + 48))(v625);
                                }
                              }
                              v630 = sub_18015D300(a1);
                              if ( v630 == 0 )
                              {
// Original IDA line 4470 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                                sub_180373340(a1, &v862, qword_1803B7610);
                                v7 = v864;
                                v4 = v111;
                                v6 = v867;
                                v10 = 409;
                                v8 = v865;
                                goto LABEL_1392;
                              }
// Original IDA line 4478 | qword_1803B78C0 -> novels.pipeline.short_fiction_runner#120 bytes [4005043, 4005055]: "write_text"
                              v631 = (int *)sub_18037F0C0(a1, v630, qword_1803B78C0);
                              if ( v631 == nullptr )
                              {
                                sub_18000CE30(a1, &v862);
                                v7 = v864;
                                v4 = v111;
                                v6 = v867;
                                v10 = 409;
                                v8 = v865;
                                goto LABEL_1392;
                              }
// Original IDA line 4489 | qword_1803B7730 -> novels.pipeline.short_fiction_runner#70 bytes [4004181, 4004188]: "final"
                              v632 = sub_180382540(v871, qword_1803B7730);
                              v633 = (int *)v632;
                              if ( v632 == 0 )
                              {
                                v634 = a1;
LABEL_1064:
                                sub_18000CE30(v634, &v862);
                                if ( *v631 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v631)-- == 1;
                                  if ( !v9 )
                                    goto LABEL_1068;
                                  v635 = sub_18000CC40(v631);
                                  v636 = v631;
LABEL_1067:
                                  (*(void (__fastcall **)(int *))(v635 + 48))(v636);
                                }
LABEL_1068:
                                v7 = v864;
                                v4 = v111;
                                v6 = v867;
                                v10 = 409;
                                v8 = v865;
                                goto LABEL_1392;
                              }
// Original IDA line 4514 | qword_1803B7AF8 -> novels.pipeline.short_fiction_runner#191 bytes [4006829, 4006846]: "cover-prompt.md"
                              v637 = (int *)sub_180382540(v632, qword_1803B7AF8);
                              if ( *v633 >= 0 )
                              {
                                v9 = (*(_QWORD *)v633)-- == 1;
                                if ( v9 )
                                {
                                  v638 = sub_18000CC40(v633);
                                  (*(void (__fastcall **)(int *))(v638 + 48))(v633);
                                }
                              }
                              v634 = a1;
                              if ( v637 == nullptr )
                                goto LABEL_1064;
// Original IDA line 4527 | qword_1803B7AF0 -> novels.pipeline.short_fiction_runner#190 bytes [4006815, 4006829]: "cover_prompt"
                              v639 = sub_18037F0C0(a1, v868, qword_1803B7AF0);
                              v640 = (int *)v639;
                              if ( v639 == 0 )
                              {
                                sub_18000CE30(a1, &v862);
                                if ( *v631 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v631)-- == 1;
                                  if ( v9 )
                                  {
                                    v641 = sub_18000CC40(v631);
                                    (*(void (__fastcall **)(int *))(v641 + 48))(v631);
                                  }
                                }
                                if ( *v637 < 0 )
                                  goto LABEL_1068;
                                v9 = (*(_QWORD *)v637)-- == 1;
                                if ( !v9 )
                                  goto LABEL_1068;
                                v635 = sub_18000CC40(v637);
                                v636 = v637;
                                goto LABEL_1067;
                              }
                              v642 = sub_180001510(v639);
                              if ( v642 == -1 )
                              {
                                sub_18000CE30(a1, &v862);
                                if ( *v631 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v631)-- == 1;
                                  if ( v9 )
                                  {
                                    v643 = sub_18000CC40(v631);
                                    (*(void (__fastcall **)(int *))(v643 + 48))(v631);
                                  }
                                }
                                if ( *v637 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v637)-- == 1;
                                  if ( v9 )
                                  {
                                    v644 = sub_18000CC40(v637);
                                    (*(void (__fastcall **)(int *))(v644 + 48))(v637);
                                  }
                                }
                                if ( *v640 < 0 )
                                  goto LABEL_1068;
                                v9 = (*(_QWORD *)v640)-- == 1;
                                if ( !v9 )
                                  goto LABEL_1068;
                                v635 = sub_18000CC40(v640);
                                v636 = v640;
                                goto LABEL_1067;
                              }
                              if ( v642 != 1 )
                              {
                                if ( *v640 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v640)-- == 1;
                                  if ( v9 )
                                  {
                                    v645 = sub_18000CC40(v640);
                                    (*(void (__fastcall **)(int *))(v645 + 48))(v640);
                                  }
                                }
// Original IDA line 4592 | qword_1803B7B00 -> novels.pipeline.short_fiction_runner#192 bytes [4006846, 4006855]: "(empty)"
                                v640 = (int *)qword_1803B7B00;
// Original IDA line 4593 | qword_1803B7B00 -> novels.pipeline.short_fiction_runner#192 bytes [4006846, 4006855]: "(empty)"
                                if ( *(_DWORD *)qword_1803B7B00 != -1 )
// Original IDA line 4594 | qword_1803B7B00 -> novels.pipeline.short_fiction_runner#192 bytes [4006846, 4006855]: "(empty)"
                                  ++*(_DWORD *)qword_1803B7B00;
                              }
                              v10 = 409;
                              *(_DWORD *)(v111 + 40) = 409;
                              v955[0] = v637;
                              v955[1] = v640;
                              v646 = (int *)sub_180377750(a1, v631, v955);
                              if ( *v631 >= 0 )
                              {
                                v9 = (*(_QWORD *)v631)-- == 1;
                                if ( v9 )
                                {
                                  v647 = sub_18000CC40(v631);
                                  (*(void (__fastcall **)(int *))(v647 + 48))(v631);
                                }
                              }
                              if ( *v637 >= 0 )
                              {
                                v9 = (*(_QWORD *)v637)-- == 1;
                                if ( v9 )
                                {
                                  v648 = sub_18000CC40(v637);
                                  (*(void (__fastcall **)(int *))(v648 + 48))(v637);
                                }
                              }
                              if ( *v640 >= 0 )
                              {
                                v9 = (*(_QWORD *)v640)-- == 1;
                                if ( v9 )
                                {
                                  v649 = sub_18000CC40(v640);
                                  (*(void (__fastcall **)(int *))(v649 + 48))(v640);
                                }
                              }
                              if ( v646 != nullptr )
                              {
                                if ( *v646 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v646)-- == 1;
                                  if ( v9 )
                                  {
                                    v650 = sub_18000CC40(v646);
                                    (*(void (__fastcall **)(int *))(v650 + 48))(v646);
                                  }
                                }
// Original IDA line 4639 | qword_1803B7558 -> novels.pipeline.short_fiction_runner#11 bytes [4003453, 4003461]: "status"
                                v651 = qword_1803B7558;
// Original IDA line 4640 | qword_1803B7B08 -> novels.pipeline.short_fiction_runner#193 bytes [4006855, 4006866]: "completed"
                                v652 = qword_1803B7B08;
                                v515 = (int *)PyDict_NewPresized(3);
                                PyDict_SetItem(v515, v651, v652);
// Original IDA line 4643 | qword_1803B7570 -> novels.pipeline.short_fiction_runner#14 bytes [4003489, 4003498]: "storyId"
                                PyDict_SetItem(v515, qword_1803B7570, v888);
// Original IDA line 4644 | qword_1803B7778 -> novels.pipeline.short_fiction_runner#79 bytes [4004271, 4004278]: "title"
                                v653 = qword_1803B7778;
// Original IDA line 4645 | qword_1803B7A70 -> novels.pipeline.short_fiction_runner#174 bytes [4006454, 4006467]: "story_title"
                                v654 = sub_18037F0C0(a1, v877, qword_1803B7A70);
                                v655 = (int *)v654;
                                if ( v654 == 0 )
                                {
                                  sub_18000CE30(a1, &v862);
                                  v10 = 411;
                                  goto LABEL_896;
                                }
                                PyDict_SetItem(v515, v653, v654);
                                if ( *v655 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v655)-- == 1;
                                  if ( v9 )
                                  {
                                    v656 = sub_18000CC40(v655);
                                    (*(void (__fastcall **)(int *))(v656 + 48))(v655);
                                  }
                                }
                                v657 = v886;
                                v920 = v515;
                                if ( v886 == nullptr )
                                {
// Original IDA line 4667 | qword_1803B7B10 -> novels.pipeline.short_fiction_runner#194 bytes [4006866, 4006884]: "revision_warning"
                                  sub_180370FE0(a1, &v862, qword_1803B7B10);
                                  sub_18000D1D0(a1, v862);
                                  v10 = 412;
                                  goto LABEL_1388;
                                }
                                v658 = sub_180001510(v886);
                                if ( v658 == -1 )
                                {
                                  sub_18000CE30(a1, &v862);
                                  v10 = 412;
                                  goto LABEL_1388;
                                }
                                if ( v658 != 0 )
                                {
// Original IDA line 4681 | qword_1803B7B18 -> novels.pipeline.short_fiction_runner#195 bytes [4006884, 4006904]: "revision skipped: "
                                  v659 = (_DWORD *)qword_1803B7B18;
                                  v660 = sub_180374A40(a1, 2);
                                  v560 = (int *)v660;
                                  if ( *v659 != -1 )
                                    ++*v659;
                                  sub_18000CC80(v660, 0, v659);
// Original IDA line 4687 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                                  v661 = sub_180375520(a1, v657, qword_1803B7598);
                                  if ( v661 == 0 )
                                  {
                                    sub_18000CE30(a1, &v862);
                                    v10 = 413;
                                    if ( *v560 < 0 )
                                      goto LABEL_1388;
                                    v9 = (*(_QWORD *)v560)-- == 1;
                                    if ( !v9 )
                                      goto LABEL_1388;
                                    goto LABEL_952;
                                  }
                                  sub_18000CC80(v560, 1, v661);
// Original IDA line 4700 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
                                  v662 = (int *)PyUnicode_Join(qword_1803B7598, v560);
                                  if ( *v560 >= 0 )
                                  {
                                    v9 = (*(_QWORD *)v560)-- == 1;
                                    if ( v9 )
                                    {
                                      v663 = sub_18000CC40(v560);
                                      (*(void (__fastcall **)(int *))(v663 + 48))(v560);
                                    }
                                  }
                                  if ( v662 == nullptr )
                                  {
                                    sub_18000CE30(a1, &v862);
                                    v10 = 413;
                                    goto LABEL_1388;
                                  }
// Original IDA line 4716 | qword_1803B7A48 -> novels.pipeline.short_fiction_runner#169 bytes [4006330, 4006339]: "warning"
                                  PyDict_SetItem(v515, qword_1803B7A48, v662);
                                  if ( *v662 >= 0 )
                                  {
                                    v9 = (*(_QWORD *)v662)-- == 1;
                                    if ( v9 )
                                    {
                                      v664 = sub_18000CC40(v662);
                                      (*(void (__fastcall **)(int *))(v664 + 48))(v662);
                                    }
                                  }
                                }
                                v665 = sub_18015D300(a1);
                                if ( v665 == 0 )
                                {
// Original IDA line 4730 | qword_1803B7610 -> novels.pipeline.short_fiction_runner#34 bytes [4003814, 4003830]: "FileRepository"
                                  sub_180373340(a1, &v862, qword_1803B7610);
                                  v10 = 414;
                                  goto LABEL_1388;
                                }
// Original IDA line 4734 | qword_1803B77E0 -> novels.pipeline.short_fiction_runner#92 bytes [4004441, 4004453]: "write_json"
                                v666 = (int *)sub_18037F0C0(a1, v665, qword_1803B77E0);
                                if ( v666 == nullptr )
                                {
                                  sub_18000CE30(a1, &v862);
                                  v10 = 414;
                                  goto LABEL_1388;
                                }
// Original IDA line 4741 | qword_1803B7750 -> novels.pipeline.short_fiction_runner#74 bytes [4004221, 4004234]: "status.json"
                                v667 = sub_180382540(v871, qword_1803B7750);
                                v668 = (int *)v667;
                                if ( v667 == 0 )
                                {
                                  sub_18000CE30(a1, &v862);
                                  if ( *v666 >= 0 )
                                  {
                                    v9 = (*(_QWORD *)v666)-- == 1;
                                    if ( v9 )
                                    {
                                      v669 = sub_18000CC40(v666);
                                      (*(void (__fastcall **)(int *))(v669 + 48))(v666);
                                    }
                                  }
                                  v10 = 414;
                                  goto LABEL_1388;
                                }
                                v4 = v866;
                                v10 = 414;
                                *(_DWORD *)(v866 + 40) = 414;
                                v956[0] = v667;
                                v956[1] = v515;
                                v670 = (int *)sub_180377750(a1, v666, v956);
                                if ( *v666 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v666)-- == 1;
                                  if ( v9 )
                                  {
                                    v671 = sub_18000CC40(v666);
                                    (*(void (__fastcall **)(int *))(v671 + 48))(v666);
                                  }
                                }
                                if ( *v668 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v668)-- == 1;
                                  if ( v9 )
                                  {
                                    v672 = sub_18000CC40(v668);
                                    (*(void (__fastcall **)(int *))(v672 + 48))(v668);
                                  }
                                }
                                if ( v670 == nullptr )
                                  goto LABEL_1153;
                                if ( *v670 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v670)-- == 1;
                                  if ( v9 )
                                  {
                                    v673 = sub_18000CC40(v670);
                                    (*(void (__fastcall **)(int *))(v673 + 48))(v670);
                                  }
                                }
                                v674 = v875;
                                v10 = 415;
                                *(_DWORD *)(v4 + 40) = 415;
// Original IDA line 4796 | qword_1803B7828 -> novels.pipeline.short_fiction_runner#101 bytes [4004623, 4004631]: "_stage"; qword_1803B7B20 -> novels.pipeline.short_fiction_runner#196 bytes [4006904, 4006919]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [4006906, 4006908], "value": 6}, {"tag": "a", "fileRange": [4006908, 4006919], "value": "completed"}]}
                                v675 = (int *)sub_18037E820(a1, v674, *(_QWORD *)&qword_1803B7828, qword_1803B7B20 + 24);
                                v676 = v675;
                                if ( v675 == nullptr )
                                  goto LABEL_1153;
                                if ( *v675 >= 0 )
                                {
                                  v9 = (*(_QWORD *)v675)-- == 1;
                                  if ( v9 )
                                  {
                                    v677 = sub_18000CC40(v675);
                                    (*(void (__fastcall **)(int *))(v677 + 48))(v676);
                                  }
                                }
// Original IDA line 4809 | qword_1803B7540 -> novels.pipeline.short_fiction_runner#8 bytes [4003422, 4003431]: "_nested"
                                v678 = sub_18037F0C0(a1, v674, qword_1803B7540);
                                v679 = (int *)v678;
                                if ( v678 != 0 )
                                {
                                  v680 = sub_180001510(v678);
                                  if ( *v679 >= 0 )
                                  {
                                    v9 = (*(_QWORD *)v679)-- == 1;
                                    if ( v9 )
                                    {
                                      v681 = sub_18000CC40(v679);
                                      (*(void (__fastcall **)(int *))(v681 + 48))(v679);
                                    }
                                  }
                                  if ( v680 != -1 )
                                  {
                                    if ( v680 == 0 )
                                    {
// Original IDA line 4827 | qword_1803B7520 -> novels.pipeline.short_fiction_runner#4 bytes [4003358, 4003365]: "_emit"
                                      v682 = (int *)sub_18037F0C0(a1, v674, qword_1803B7520);
                                      if ( v682 == nullptr )
                                      {
                                        sub_18000CE30(a1, &v862);
                                        v6 = v867;
                                        v10 = 417;
                                        v7 = v864;
                                        v8 = v865;
                                        goto LABEL_1392;
                                      }
// Original IDA line 4837 | qword_1803B77A0 -> novels.pipeline.short_fiction_runner#84 bytes [4004328, 4004347]: "short_fiction_run"
                                      v683 = qword_1803B77A0;
// Original IDA line 4838 | qword_1803B7A70 -> novels.pipeline.short_fiction_runner#174 bytes [4006454, 4006467]: "story_title"
                                      v684 = (int *)sub_18037F0C0(a1, v877, qword_1803B7A70);
                                      if ( v684 == nullptr )
                                      {
                                        sub_18000CE30(a1, &v862);
                                        if ( *v682 >= 0 )
                                        {
                                          v9 = (*(_QWORD *)v682)-- == 1;
                                          if ( v9 )
                                          {
                                            v685 = sub_18000CC40(v682);
                                            (*(void (__fastcall **)(int *))(v685 + 48))(v682);
                                          }
                                        }
                                        v10 = 418;
                                        goto LABEL_1388;
                                      }
                                      v10 = 417;
                                      *(_DWORD *)(v866 + 40) = 417;
                                      v964[2] = v888;
                                      v964[0] = v683;
                                      v964[1] = Py_TrueStruct;
                                      v964[3] = v684;
// Original IDA line 4860 | qword_1803B7B28 -> novels.pipeline.short_fiction_runner#197 bytes [4006919, 4006931]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4006921, 4006931], "value": "tool:end"}]}; qword_1803B7B30 -> novels.pipeline.short_fiction_runner#198 bytes [4006931, 4006964]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4006933, 4006939], "value": "tool"}, {"tag": "a", "fileRange": [4006939, 4006948], "value": "success"}, {"tag": "a", "fileRange": [4006948, 4006957], "value": "storyId"}, {"tag": "a", "... (truncated literal)
                                      v686 = (int *)sub_18037D280(a1, v682, qword_1803B7B28, v964, qword_1803B7B30);
                                      if ( *v682 >= 0 )
                                      {
                                        v9 = (*(_QWORD *)v682)-- == 1;
                                        if ( v9 )
                                        {
                                          v687 = sub_18000CC40(v682);
                                          (*(void (__fastcall **)(int *))(v687 + 48))(v682);
                                        }
                                      }
                                      if ( *v684 >= 0 )
                                      {
                                        v9 = (*(_QWORD *)v684)-- == 1;
                                        if ( v9 )
                                        {
                                          v688 = sub_18000CC40(v684);
                                          (*(void (__fastcall **)(int *))(v688 + 48))(v684);
                                        }
                                      }
                                      if ( v686 == nullptr )
                                        goto LABEL_1178;
                                      if ( *v686 >= 0 )
                                      {
                                        v9 = (*(_QWORD *)v686)-- == 1;
                                        if ( v9 )
                                        {
                                          v689 = sub_18000CC40(v686);
                                          (*(void (__fastcall **)(int *))(v689 + 48))(v686);
                                        }
                                      }
                                      v4 = v866;
                                    }
                                    v690 = v877;
// Original IDA line 4893 | qword_1803B7938 -> novels.pipeline.short_fiction_runner#135 bytes [4005310, 4005320]: "chapters"
                                    v691 = sub_18037F0C0(a1, v877, qword_1803B7938);
                                    v692 = (int *)v691;
                                    if ( v691 != 0 )
                                    {
                                      v693 = (int *)sub_18000D280(a1, v691);
                                      if ( *v692 >= 0 )
                                      {
                                        v9 = (*(_QWORD *)v692)-- == 1;
                                        if ( v9 )
                                        {
                                          v694 = sub_18000CC40(v692);
                                          (*(void (__fastcall **)(int *))(v694 + 48))(v692);
                                        }
                                      }
                                      if ( v693 != nullptr )
                                      {
                                        v938 = sub_18036D750(v693);
                                        v695 = (int *)sub_18016C1C0(a1, &v938);
                                        if ( *v693 >= 0 )
                                        {
                                          v9 = (*(_QWORD *)v693)-- == 1;
                                          if ( v9 )
                                          {
                                            v696 = sub_18000CC40(v693);
                                            (*(void (__fastcall **)(int *))(v696 + 48))(v693);
                                          }
                                        }
                                        v697 = sub_180375A90(a1, v695);
                                        if ( *v695 >= 0 )
                                        {
                                          v9 = (*(_QWORD *)v695)-- == 1;
                                          if ( v9 )
                                          {
                                            v698 = sub_18000CC40(v695);
                                            (*(void (__fastcall **)(int *))(v698 + 48))(v695);
                                          }
                                        }
                                        if ( v697 != 0 )
                                        {
// Original IDA line 4932 | qword_1803B7570 -> novels.pipeline.short_fiction_runner#14 bytes [4003489, 4003498]: "storyId"
                                          v699 = qword_1803B7570;
                                          v892 = (int *)v697;
                                          v50 = (int *)PyDict_NewPresized(5);
                                          PyDict_SetItem(v50, v699, v888);
// Original IDA line 4936 | qword_1803B7778 -> novels.pipeline.short_fiction_runner#79 bytes [4004271, 4004278]: "title"
                                          v700 = qword_1803B7778;
// Original IDA line 4937 | qword_1803B7A70 -> novels.pipeline.short_fiction_runner#174 bytes [4006454, 4006467]: "story_title"
                                          v701 = sub_18037F0C0(a1, v690, qword_1803B7A70);
                                          v702 = (int *)v701;
                                          if ( v701 != 0 )
                                          {
                                            PyDict_SetItem(v50, v700, v701);
                                            if ( *v702 >= 0 )
                                            {
                                              v9 = (*(_QWORD *)v702)-- == 1;
                                              if ( v9 )
                                              {
                                                v703 = sub_18000CC40(v702);
                                                (*(void (__fastcall **)(int *))(v703 + 48))(v702);
                                              }
                                            }
// Original IDA line 4951 | qword_1803B7780 -> novels.pipeline.short_fiction_runner#80 bytes [4004278, 4004291]: "outlinePath"
                                            v704 = qword_1803B7780;
// Original IDA line 4952 | qword_1803B7718 -> novels.pipeline.short_fiction_runner#67 bytes [4004155, 4004164]: "outline"
                                            v705 = sub_180382540(v871, qword_1803B7718);
                                            v706 = (int *)v705;
                                            if ( v705 != 0 )
                                            {
// Original IDA line 4956 | qword_1803B7788 -> novels.pipeline.short_fiction_runner#81 bytes [4004291, 4004300]: "v002.md"
                                              v707 = (int *)sub_180382540(v705, qword_1803B7788);
                                              if ( *v706 >= 0 )
                                              {
                                                v9 = (*(_QWORD *)v706)-- == 1;
                                                if ( v9 )
                                                {
                                                  v708 = sub_18000CC40(v706);
                                                  (*(void (__fastcall **)(int *))(v708 + 48))(v706);
                                                }
                                              }
                                              if ( v707 != nullptr )
                                              {
                                                v709 = (int *)sub_1803749A0(v707);
                                                if ( *v707 >= 0 )
                                                {
                                                  v9 = (*(_QWORD *)v707)-- == 1;
                                                  if ( v9 )
                                                  {
                                                    v710 = sub_18000CC40(v707);
                                                    (*(void (__fastcall **)(int *))(v710 + 48))(v707);
                                                  }
                                                }
                                                if ( v709 != nullptr )
                                                {
                                                  PyDict_SetItem(v50, v704, v709);
                                                  if ( *v709 >= 0 )
                                                  {
                                                    v9 = (*(_QWORD *)v709)-- == 1;
                                                    if ( v9 )
                                                    {
                                                      v711 = sub_18000CC40(v709);
                                                      (*(void (__fastcall **)(int *))(v711 + 48))(v709);
                                                    }
                                                  }
// Original IDA line 4990 | qword_1803B7790 -> novels.pipeline.short_fiction_runner#82 bytes [4004300, 4004319]: "finalMarkdownPath"
                                                  v712 = qword_1803B7790;
// Original IDA line 4991 | qword_1803B7730 -> novels.pipeline.short_fiction_runner#70 bytes [4004181, 4004188]: "final"
                                                  v713 = sub_180382540(v871, qword_1803B7730);
                                                  v714 = (int *)v713;
                                                  if ( v713 != 0 )
                                                  {
// Original IDA line 4995 | qword_1803B7738 -> novels.pipeline.short_fiction_runner#71 bytes [4004188, 4004197]: "full.md"
                                                    v715 = (int *)sub_180382540(v713, qword_1803B7738);
                                                    if ( *v714 >= 0 )
                                                    {
                                                      v9 = (*(_QWORD *)v714)-- == 1;
                                                      if ( v9 )
                                                      {
                                                        v716 = sub_18000CC40(v714);
                                                        (*(void (__fastcall **)(int *))(v716 + 48))(v714);
                                                      }
                                                    }
                                                    if ( v715 != nullptr )
                                                    {
                                                      v717 = (int *)sub_1803749A0(v715);
                                                      if ( *v715 >= 0 )
                                                      {
                                                        v9 = (*(_QWORD *)v715)-- == 1;
                                                        if ( v9 )
                                                        {
                                                          v718 = sub_18000CC40(v715);
                                                          (*(void (__fastcall **)(int *))(v718 + 48))(v715);
                                                        }
                                                      }
                                                      if ( v717 != nullptr )
                                                      {
                                                        PyDict_SetItem(v50, v712, v717);
                                                        if ( *v717 >= 0 )
                                                        {
                                                          v9 = (*(_QWORD *)v717)-- == 1;
                                                          if ( v9 )
                                                          {
                                                            v720 = sub_18000CC40(v717);
                                                            (*(void (__fastcall **)(int *))(v720 + 48))(v717);
                                                          }
                                                        }
// Original IDA line 5029 | qword_1803B7B38 -> novels.pipeline.short_fiction_runner#199 bytes [4006964, 4006975]: "wordCount"
                                                        PyDict_SetItem(v50, qword_1803B7B38, v697);
                                                        sub_180001D80(a1, v721);
                                                        v722 = v939;
                                                        v723 = v879;
                                                        v724 = v863;
                                                        v725 = v868;
                                                        v897 = v908;
                                                        v896 = v909;
                                                        v893 = v924;
                                                        v864 = v922;
                                                        v882 = v916;
                                                        v883 = v917;
                                                        v884 = v918;
                                                        v900 = v919;
                                                        v894 = (int *)v903;
                                                        v898 = v904;
                                                        v877 = v914;
                                                        v899 = v873;
                                                        v865 = v915;
                                                        v726 = v886;
                                                        v887 = (int *)v902;
                                                        v885 = (int *)v905;
                                                        v22 = *v886 < 0;
                                                        v867 = v923;
                                                        v901 = v939;
                                                        v889 = v879;
                                                        v890 = v863;
                                                        v891 = v868;
                                                        v892 = (int *)v697;
                                                        if ( !v22 )
                                                        {
                                                          v9 = (*(_QWORD *)v886)-- == 1;
                                                          if ( v9 )
                                                          {
                                                            v727 = sub_18000CC40(v726);
                                                            (*(void (__fastcall **)(int *))(v727 + 48))(v886);
                                                          }
                                                          v897 = v908;
                                                          v896 = v909;
                                                          v893 = v924;
                                                          v864 = v922;
                                                          v882 = v916;
                                                          v883 = v917;
                                                          v884 = v918;
                                                          v900 = v919;
                                                          v894 = (int *)v903;
                                                          v898 = v904;
                                                          v877 = v914;
                                                          v899 = v873;
                                                          v887 = (int *)v902;
                                                          v885 = (int *)v905;
                                                          v867 = v923;
                                                          v865 = v915;
                                                          v901 = v722;
                                                          v889 = v723;
                                                          v890 = v724;
                                                          v891 = v725;
                                                          v892 = (int *)v697;
                                                        }
                                                        if ( v515 != nullptr && *v515 >= 0 )
                                                        {
                                                          v9 = (*(_QWORD *)v515)-- == 1;
                                                          if ( v9 )
                                                          {
                                                            v728 = sub_18000CC40(v515);
                                                            (*(void (__fastcall **)(int *))(v728 + 48))(v515);
                                                          }
                                                        }
                                                        v47 = v895;
                                                        v23 = v921;
                                                        v45 = v875;
                                                        v26 = (int *)v880;
                                                        goto LABEL_1242;
                                                      }
                                                    }
                                                  }
                                                  v10 = 425;
                                                }
                                                else
                                                {
                                                  v10 = 424;
                                                }
                                              }
                                              else
                                              {
                                                v10 = 424;
                                              }
                                            }
                                            else
                                            {
                                              v10 = 424;
                                            }
                                          }
                                          else
                                          {
                                            v10 = 423;
                                          }
                                          sub_18000CE30(a1, &v862);
                                          if ( *v50 >= 0 )
                                          {
                                            v9 = (*(_QWORD *)v50)-- == 1;
                                            if ( v9 )
                                            {
                                              v719 = sub_18000CC40(v50);
                                              (*(void (__fastcall **)(int *))(v719 + 48))(v50);
                                            }
                                          }
                                          goto LABEL_1388;
                                        }
                                      }
                                    }
                                    sub_18000CE30(a1, &v862);
                                    v6 = v867;
                                    v10 = 420;
                                    v7 = v864;
                                    v8 = v865;
                                    goto LABEL_1392;
                                  }
                                }
                                sub_18000CE30(a1, &v862);
                                v6 = v867;
                                v10 = 416;
                                v7 = v864;
                                v8 = v865;
                                goto LABEL_1392;
                              }
                            }
                          }
LABEL_197:
                          v120 = a1;
                          goto LABEL_198;
                        }
                      }
                      sub_18000CE30(a1, &v862);
                      goto LABEL_797;
                    }
                    v348 = 353;
                    *(_DWORD *)(v4 + 40) = 353;
                    v422 = sub_180376770(a1, v419, v405);
                    v423 = (int *)v422;
                    if ( v422 == 0 )
                    {
                      sub_18000CE30(a1, &v862);
                      if ( *v411 >= 0 )
                      {
                        v9 = (*(_QWORD *)v411)-- == 1;
                        if ( v9 )
                        {
                          v424 = sub_18000CC40(v411);
                          (*(void (__fastcall **)(int *))(v424 + 48))(v411);
                        }
                      }
                      if ( *v417 >= 0 )
                      {
                        v9 = (*(_QWORD *)v417)-- == 1;
                        if ( v9 )
                        {
                          v425 = sub_18000CC40(v417);
                          (*(void (__fastcall **)(int *))(v425 + 48))(v417);
                        }
                      }
                      goto LABEL_718;
                    }
                    *(_DWORD *)(v4 + 40) = 353;
                    v949[0] = v417;
                    v949[1] = v422;
                    v426 = (int *)sub_180377750(a1, v411, v949);
                    if ( *v411 >= 0 )
                    {
                      v9 = (*(_QWORD *)v411)-- == 1;
                      if ( v9 )
                      {
                        v427 = sub_18000CC40(v411);
                        (*(void (__fastcall **)(int *))(v427 + 48))(v411);
                      }
                    }
                    if ( *v417 >= 0 )
                    {
                      v9 = (*(_QWORD *)v417)-- == 1;
                      if ( v9 )
                      {
                        v428 = sub_18000CC40(v417);
                        (*(void (__fastcall **)(int *))(v428 + 48))(v417);
                      }
                    }
                    if ( *v423 >= 0 )
                    {
                      v9 = (*(_QWORD *)v423)-- == 1;
                      if ( v9 )
                      {
                        v429 = sub_18000CC40(v423);
                        (*(void (__fastcall **)(int *))(v429 + 48))(v423);
                      }
                    }
                    if ( v426 != nullptr )
                    {
                      if ( *v426 >= 0 )
                      {
                        v9 = (*(_QWORD *)v426)-- == 1;
                        if ( v9 )
                        {
                          v479 = sub_18000CC40(v426);
                          (*(void (__fastcall **)(int *))(v479 + 48))(v426);
                        }
                      }
                      v480 = *v405;
                      v476 = v405;
                      v877 = v405;
                      v481 = v480 + 1;
                      if ( v481 != 0 )
                        *v405 = v481;
                      v6 = v867;
                      if ( *(int *)v867 >= 0 )
                      {
                        v9 = (*v867)-- == 1;
                        if ( v9 )
                        {
                          v482 = sub_18000CC40(v6);
                          (*(void (__fastcall **)(_QWORD *))(v482 + 48))(v6);
                        }
                      }
                      goto LABEL_824;
                    }
                    v401 = a1;
                  }
                  sub_18000CE30(v401, &v862);
                  goto LABEL_717;
                }
              }
              sub_18000CE30(a1, &v862);
LABEL_634:
              if ( *v355 < 0 )
                goto LABEL_638;
              v9 = (*(_QWORD *)v355)-- == 1;
              if ( !v9 )
                goto LABEL_638;
              v373 = sub_18000CC40(v355);
              v374 = v355;
              goto LABEL_637;
            }
            v10 = 293;
            while ( 1 )
            {
              v132 = v904;
              v133 = v865;
              v134 = v904;
              v916 = v7;
              v135 = v865;
              v917 = v6;
              v904 = v131;
              if ( v134 != nullptr && *v132 >= 0 )
              {
                v9 = (*(_QWORD *)v132)-- == 1;
                if ( v9 )
                {
                  v136 = sub_18000CC40(v132);
                  (*(void (__fastcall **)(int *))(v136 + 48))(v134);
                  v131 = v868;
                  v133 = v135;
                }
              }
              v137 = *v131;
              v138 = v133;
              v865 = v131;
              v139 = v137 + 1;
              if ( v139 != 0 )
                *v131 = v139;
              if ( v133 != nullptr && *v135 >= 0 )
              {
                v9 = (*(_QWORD *)v135)-- == 1;
                if ( v9 )
                {
                  v140 = sub_18000CC40(v133);
                  (*(void (__fastcall **)(int *))(v140 + 48))(v138);
                }
              }
              v7 = v864;
              if ( v864 == nullptr )
              {
// Original IDA line 5308 | qword_1803B78E0 -> novels.pipeline.short_fiction_runner#124 bytes [4005150, 4005159]: "missing"
                sub_180370FE0(a1, &v862, qword_1803B78E0);
                sub_18000D1D0(a1, v862);
                v10 = 269;
                goto LABEL_385;
              }
              v141 = sub_180001510(v864);
              if ( v141 == -1 )
              {
                sub_18000CE30(a1, &v862);
                v10 = 269;
                goto LABEL_385;
              }
              if ( v141 == 0 )
                goto LABEL_418;
              if ( v875 == nullptr )
              {
// Original IDA line 5324 | qword_1803B75A8 -> novels.pipeline.short_fiction_runner#21 bytes [4003611, 4003617]: "self"
                sub_180370FE0(a1, &v862, qword_1803B75A8);
                sub_18000D1D0(a1, v862);
                v10 = 271;
                goto LABEL_385;
              }
// Original IDA line 5329 | qword_1803B7520 -> novels.pipeline.short_fiction_runner#4 bytes [4003358, 4003365]: "_emit"
              v142 = (int *)sub_18037F0C0(a1, v875, qword_1803B7520);
              if ( v142 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 271;
                goto LABEL_385;
              }
// Original IDA line 5336 | qword_1803B78F8 -> novels.pipeline.short_fiction_runner#127 bytes [4005169, 4005192]: "补写缺失章节："
              v143 = (_DWORD *)qword_1803B78F8;
// Original IDA line 5337 | qword_1803B78E8 -> novels.pipeline.short_fiction_runner#125 bytes [4005159, 4005161]: 3
              v869 = (_QWORD *)qword_1803B78E8;
// Original IDA line 5338 | qword_1803B78F0 -> novels.pipeline.short_fiction_runner#126 bytes [4005161, 4005169]: "active"
              v905 = qword_1803B78F0;
              v144 = sub_180374A40(a1, 5);
              v145 = (int *)v144;
              if ( *v143 != -1 )
                ++*v143;
              sub_18000CC80(v144, 0, v143);
// Original IDA line 5344 | qword_1803B7900 -> novels.pipeline.short_fiction_runner#128 bytes [4005192, 4005196]: ", "
              v146 = qword_1803B7900;
              v147 = sub_18000D280(a1, v864);
              v148 = (int *)v147;
              if ( v147 == 0 )
                goto LABEL_408;
              v928 = sub_18036D750(v147);
              v149 = (int *)sub_18016BD40(a1, &v928);
              if ( *v148 >= 0 )
              {
                v9 = (*(_QWORD *)v148)-- == 1;
                if ( v9 )
                {
                  v150 = sub_18000CC40(v148);
                  (*(void (__fastcall **)(int *))(v150 + 48))(v148);
                }
              }
              v151 = sub_180374990(a1, v146, v149);
              if ( *v149 >= 0 )
              {
                v9 = (*(_QWORD *)v149)-- == 1;
                if ( v9 )
                {
                  v152 = sub_18000CC40(v149);
                  (*(void (__fastcall **)(int *))(v152 + 48))(v149);
                }
              }
              if ( v151 == 0 )
                goto LABEL_408;
              sub_18000CC80(v145, 1, v151);
// Original IDA line 5373 | qword_1803B7908 -> novels.pipeline.short_fiction_runner#129 bytes [4005196, 4005204]: "（第"
              v153 = qword_1803B7908;
// Original IDA line 5374 | qword_1803B7908 -> novels.pipeline.short_fiction_runner#129 bytes [4005196, 4005204]: "（第"
              if ( *(_DWORD *)qword_1803B7908 != -1 )
// Original IDA line 5375 | qword_1803B7908 -> novels.pipeline.short_fiction_runner#129 bytes [4005196, 4005204]: "（第"
                ++*(_DWORD *)qword_1803B7908;
              sub_18000CC80(v145, 2, v153);
// Original IDA line 5377 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
              v154 = sub_180375520(a1, v868, qword_1803B7598);
              if ( v154 == 0 )
              {
LABEL_408:
                v10 = 272;
                sub_18000CE30(a1, &v862);
                if ( *v142 >= 0 )
                {
                  v9 = (*(_QWORD *)v142)-- == 1;
                  if ( v9 )
                  {
                    v243 = sub_18000CC40(v142);
                    (*(void (__fastcall **)(int *))(v243 + 48))(v142);
                  }
                }
                if ( *v145 >= 0 )
                {
                  v9 = (*(_QWORD *)v145)-- == 1;
                  if ( v9 )
                  {
                    v244 = sub_18000CC40(v145);
                    (*(void (__fastcall **)(int *))(v244 + 48))(v145);
                  }
                }
                goto LABEL_383;
              }
              sub_18000CC80(v145, 3, v154);
// Original IDA line 5404 | qword_1803B7910 -> novels.pipeline.short_fiction_runner#130 bytes [4005204, 4005212]: "次）"
              v155 = qword_1803B7910;
// Original IDA line 5405 | qword_1803B7910 -> novels.pipeline.short_fiction_runner#130 bytes [4005204, 4005212]: "次）"
              if ( *(_DWORD *)qword_1803B7910 != -1 )
// Original IDA line 5406 | qword_1803B7910 -> novels.pipeline.short_fiction_runner#130 bytes [4005204, 4005212]: "次）"
                ++*(_DWORD *)qword_1803B7910;
              sub_18000CC80(v145, 4, v155);
// Original IDA line 5408 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
              v156 = (int *)PyUnicode_Join(qword_1803B7598, v145);
              if ( *v145 >= 0 )
              {
                v9 = (*(_QWORD *)v145)-- == 1;
                if ( v9 )
                {
                  v157 = sub_18000CC40(v145);
                  (*(void (__fastcall **)(int *))(v157 + 48))(v145);
                }
              }
              if ( v156 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                if ( *v142 >= 0 )
                {
                  v9 = (*(_QWORD *)v142)-- == 1;
                  if ( v9 )
                  {
                    v242 = sub_18000CC40(v142);
                    (*(void (__fastcall **)(int *))(v242 + 48))(v142);
                  }
                }
                v10 = 272;
                goto LABEL_383;
              }
              v158 = v869;
              v4 = v866;
              *(_DWORD *)(v866 + 40) = 271;
              v958[0] = v158;
              v958[1] = v905;
              v958[2] = v156;
// Original IDA line 5439 | qword_1803B7578 -> novels.pipeline.short_fiction_runner#15 bytes [4003498, 4003515]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4003500, 4003515], "value": "tool:progress"}]}; qword_1803B7918 -> novels.pipeline.short_fiction_runner#131 bytes [4005212, 4005243]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4005214, 4005226], "value": "stageIndex"}, {"tag": "a", "fileRange": [4005226, 4005234], "value": "status"}, {"tag": "a", "fileRange": [4005234, 4005243], "value": "message"}]}
              v159 = (int *)sub_18037D280(a1, v142, qword_1803B7578, v958, qword_1803B7918);
              if ( *v142 >= 0 )
              {
                v9 = (*(_QWORD *)v142)-- == 1;
                if ( v9 )
                {
                  v160 = sub_18000CC40(v142);
                  (*(void (__fastcall **)(int *))(v160 + 48))(v142);
                }
              }
              if ( *v156 >= 0 )
              {
                v9 = (*(_QWORD *)v156)-- == 1;
                if ( v9 )
                {
                  v161 = sub_18000CC40(v156);
                  (*(void (__fastcall **)(int *))(v161 + 48))(v156);
                }
              }
              if ( v159 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v6 = v867;
                v10 = 271;
                goto LABEL_385;
              }
              if ( *v159 >= 0 )
              {
                v9 = (*(_QWORD *)v159)-- == 1;
                if ( v9 )
                {
                  v162 = sub_18000CC40(v159);
                  (*(void (__fastcall **)(int *))(v162 + 48))(v159);
                }
              }
              v163 = sub_18015D740(a1);
              if ( v163 == 0 )
              {
// Original IDA line 5477 | qword_1803B7920 -> novels.pipeline.short_fiction_runner#132 bytes [4005243, 4005285]: "ShortFictionDraftContinuationPromptInput"
                sub_180373340(a1, &v862, qword_1803B7920);
                v6 = v867;
                v10 = 273;
                goto LABEL_385;
              }
              if ( v881 == nullptr )
              {
// Original IDA line 5484 | qword_1803B7928 -> novels.pipeline.short_fiction_runner#133 bytes [4005285, 4005296]: "direction"
                sub_180370FE0(a1, &v862, qword_1803B7928);
                sub_18000D1D0(a1, v862);
                v6 = v867;
                v10 = 274;
                goto LABEL_385;
              }
              v164 = v880;
              if ( v880 == nullptr )
              {
// Original IDA line 5493 | qword_1803B7930 -> novels.pipeline.short_fiction_runner#134 bytes [4005296, 4005310]: "outline_text"
                sub_180370FE0(a1, &v862, qword_1803B7930);
                sub_18000D1D0(a1, v862);
                v6 = v867;
                v10 = 275;
                goto LABEL_385;
              }
              if ( v876 == nullptr )
              {
// Original IDA line 5501 | qword_1803B7938 -> novels.pipeline.short_fiction_runner#135 bytes [4005310, 4005320]: "chapters"
                sub_180370FE0(a1, &v862, qword_1803B7938);
                sub_18000D1D0(a1, v862);
                v6 = v867;
                v10 = 276;
                goto LABEL_385;
              }
              if ( v878 == nullptr )
              {
// Original IDA line 5509 | qword_1803B7940 -> novels.pipeline.short_fiction_runner#136 bytes [4005320, 4005331]: "chars_per"
                sub_180370FE0(a1, &v862, qword_1803B7940);
                sub_18000D1D0(a1, v862);
                v6 = v867;
                v10 = 277;
                goto LABEL_385;
              }
              v165 = sub_18015F160(a1);
              if ( v165 == 0 )
              {
// Original IDA line 5518 | qword_1803B78D0 -> novels.pipeline.short_fiction_runner#122 bytes [4005072, 4005109]: "render_short_fiction_draft_markdown"
                sub_180373340(a1, &v862, qword_1803B78D0);
                v6 = v867;
                v10 = 278;
                goto LABEL_385;
              }
              v6 = v867;
              if ( v867 == nullptr )
              {
// Original IDA line 5526 | qword_1803B7948 -> novels.pipeline.short_fiction_runner#137 bytes [4005331, 4005341]: "draft_v1"
                sub_180370FE0(a1, &v862, qword_1803B7948);
                sub_18000D1D0(a1, v862);
                v10 = 278;
                goto LABEL_385;
              }
              v166 = v867;
              *(_DWORD *)(v4 + 40) = 278;
              v167 = (int *)sub_180376770(a1, v165, v166);
              if ( v167 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 278;
                goto LABEL_385;
              }
              v168 = v881;
              *(_DWORD *)(v4 + 40) = 273;
              v968[0] = v168;
              v968[2] = v876;
              v968[3] = v878;
              v968[5] = v864;
              v968[1] = v164;
              v968[4] = v167;
// Original IDA line 5548 | qword_1803B7950 -> novels.pipeline.short_fiction_runner#138 bytes [4005341, 4005449]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4005343, 4005354], "value": "direction"}, {"tag": "a", "fileRange": [4005354, 4005372], "value": "outline_markdown"}, {"tag": "a", "fileRange": [4005372, 4005387], "value": "chapter_cou... (truncated literal)
              v169 = sub_18037CCB0(a1, v163, v968, qword_1803B7950);
              v22 = *v167 < 0;
              v873 = (int *)v169;
              if ( !v22 )
              {
                v9 = (*(_QWORD *)v167)-- == 1;
                if ( v9 )
                {
                  v170 = sub_18000CC40(v167);
                  (*(void (__fastcall **)(int *))(v170 + 48))(v167);
                  v169 = (__int64)v873;
                }
              }
              if ( v169 == 0 )
              {
                sub_18000CE30(a1, &v862);
                v10 = 273;
                goto LABEL_385;
              }
              v171 = (int *)v882;
              v882 = (_QWORD *)v169;
              if ( v171 != nullptr && *v171 >= 0 )
              {
                v9 = (*(_QWORD *)v171)-- == 1;
                if ( v9 )
                {
                  v172 = sub_18000CC40(v171);
                  (*(void (__fastcall **)(int *))(v172 + 48))(v171);
                }
              }
// Original IDA line 5578 | qword_1803B7850 -> novels.pipeline.short_fiction_runner#106 bytes [4004772, 4004779]: "_chat"
              v173 = (int *)sub_18037F0C0(a1, v875, *(_QWORD *)&qword_1803B7850);
              if ( v173 == nullptr )
              {
                sub_18000CE30(a1, &v862);
                v10 = 281;
                goto LABEL_384;
              }
// Original IDA line 5585 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v174 = qword_1803B7860;
// Original IDA line 5586 | qword_1803B7868 -> novels.pipeline.short_fiction_runner#109 bytes [4004793, 4004801]: "system"
              v175 = qword_1803B7868;
// Original IDA line 5587 | qword_1803B7858 -> novels.pipeline.short_fiction_runner#107 bytes [4004779, 4004787]: "writer"
              v869 = (_QWORD *)qword_1803B7858;
              v176 = PyDict_NewPresized(2);
              PyDict_SetItem(v176, v174, v175);
// Original IDA line 5590 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              PyDict_SetItem(v176, qword_1803B7870, v908);
              v177 = (int *)sub_180372160(a1, 2);
              sub_180024750(v177, 0, v176);
// Original IDA line 5593 | qword_1803B7860 -> novels.pipeline.short_fiction_runner#108 bytes [4004787, 4004793]: "role"
              v178 = qword_1803B7860;
// Original IDA line 5594 | qword_1803B7878 -> novels.pipeline.short_fiction_runner#111 bytes [4004810, 4004816]: "user"
              v179 = qword_1803B7878;
              v180 = (int *)PyDict_NewPresized(2);
              PyDict_SetItem(v180, v178, v179);
// Original IDA line 5597 | qword_1803B7870 -> novels.pipeline.short_fiction_runner#110 bytes [4004801, 4004810]: "content"
              v181 = qword_1803B7870;
              v182 = sub_18015E060(a1);
              if ( v182 == 0 )
              {
// Original IDA line 5601 | qword_1803B7960 -> novels.pipeline.short_fiction_runner#140 bytes [4005461, 4005513]: "build_short_fiction_draft_continuation_user_prompt"
                sub_180373340(a1, &v862, qword_1803B7960);
                goto LABEL_374;
              }
              v183 = v873;
              *(_DWORD *)(v866 + 40) = 285;
              v184 = sub_180376770(a1, v182, v183);
              v185 = (int *)v184;
              if ( v184 == 0 )
              {
                sub_18000CE30(a1, &v862);
LABEL_374:
                v10 = 285;
                if ( *v180 >= 0 )
                {
                  v9 = (*(_QWORD *)v180)-- == 1;
                  if ( v9 )
                  {
                    v235 = sub_18000CC40(v180);
                    (*(void (__fastcall **)(int *))(v235 + 48))(v180);
                  }
                }
                if ( *v173 >= 0 )
                {
                  v9 = (*(_QWORD *)v173)-- == 1;
                  if ( v9 )
                  {
                    v236 = sub_18000CC40(v173);
                    (*(void (__fastcall **)(int *))(v236 + 48))(v173);
                  }
                }
                if ( *v177 >= 0 )
                {
                  v9 = (*(_QWORD *)v177)-- == 1;
                  if ( v9 )
                  {
                    v237 = sub_18000CC40(v177);
                    (*(void (__fastcall **)(int *))(v237 + 48))(v177);
                  }
                }
                goto LABEL_383;
              }
              PyDict_SetItem(v180, v181, v184);
              if ( *v185 >= 0 )
              {
                v9 = (*(_QWORD *)v185)-- == 1;
                if ( v9 )
                {
                  v186 = sub_18000CC40(v185);
                  (*(void (__fastcall **)(int *))(v186 + 48))(v185);
                }
              }
              sub_180024750(v177, 1, v180);
// Original IDA line 5653 | qword_1803B7888 -> novels.pipeline.short_fiction_runner#113 bytes [4004856, 4004870]: "撰写正文"
              v187 = qword_1803B7888;
              v188 = sub_18015EB00(a1);
              if ( v188 == 0 )
              {
// Original IDA line 5657 | qword_1803B7890 -> novels.pipeline.short_fiction_runner#114 bytes [4004870, 4004905]: "estimate_short_fiction_max_tokens"
                sub_180373340(a1, &v862, qword_1803B7890);
                goto LABEL_364;
              }
              v189 = (int *)sub_180375370(a1, v864);
              v190 = a1;
              if ( v189 == nullptr )
                goto LABEL_363;
              *(_DWORD *)(v866 + 40) = 288;
              v943[1] = v878;
              v943[0] = v189;
              v191 = (int *)sub_180377750(a1, v188, v943);
              if ( *v189 >= 0 )
              {
                v9 = (*(_QWORD *)v189)-- == 1;
                if ( v9 )
                {
                  v192 = sub_18000CC40(v189);
                  (*(void (__fastcall **)(int *))(v192 + 48))(v189);
                }
              }
              if ( v191 == nullptr )
              {
                v190 = a1;
LABEL_363:
                sub_18000CE30(v190, &v862);
LABEL_364:
                if ( *v173 >= 0 )
                {
                  v9 = (*(_QWORD *)v173)-- == 1;
                  if ( v9 )
                  {
                    v233 = sub_18000CC40(v173);
                    (*(void (__fastcall **)(int *))(v233 + 48))(v173);
                  }
                }
                if ( *v177 >= 0 )
                {
                  v9 = (*(_QWORD *)v177)-- == 1;
                  if ( v9 )
                  {
                    v234 = sub_18000CC40(v177);
                    (*(void (__fastcall **)(int *))(v234 + 48))(v177);
                  }
                }
                v10 = 288;
LABEL_383:
                v6 = v867;
                goto LABEL_384;
              }
// Original IDA line 5706 | qword_1803B7968 -> novels.pipeline.short_fiction_runner#141 bytes [4005513, 4005522]: 0.68
              v193 = qword_1803B7968;
              *(_DWORD *)(v866 + 40) = 281;
              v959[2] = v193;
              v944[0] = v869;
              v944[1] = v177;
              v959[0] = v187;
              v959[1] = v191;
// Original IDA line 5713 | qword_1803B78A0 -> novels.pipeline.short_fiction_runner#116 bytes [4004914, 4004954]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004916, 4004929], "value": "stage_label"}, {"tag": "a", "fileRange": [4004929, 4004941], "value": "max_tokens"}, {"tag": "a", "fileRange": [4004941, 4004954], "value": "temperature"}]}
              v194 = sub_18037D650(a1, v173, v944, v959, qword_1803B78A0);
              if ( *v173 >= 0 )
              {
                v9 = (*(_QWORD *)v173)-- == 1;
                if ( v9 )
                {
                  v195 = sub_18000CC40(v173);
                  (*(void (__fastcall **)(int *))(v195 + 48))(v173);
                }
              }
              if ( *v177 >= 0 )
              {
                v9 = (*(_QWORD *)v177)-- == 1;
                if ( v9 )
                {
                  v196 = sub_18000CC40(v177);
                  (*(void (__fastcall **)(int *))(v196 + 48))(v177);
                }
              }
              if ( *v191 >= 0 )
              {
                v9 = (*(_QWORD *)v191)-- == 1;
                if ( v9 )
                {
                  v197 = sub_18000CC40(v191);
                  (*(void (__fastcall **)(int *))(v197 + 48))(v191);
                }
              }
              if ( v194 == 0 )
              {
                sub_18000CE30(a1, &v862);
                v10 = 281;
                goto LABEL_383;
              }
              v198 = (int *)v883;
              v883 = (_QWORD *)v194;
              if ( v198 != nullptr && *v198 >= 0 )
              {
                v9 = (*(_QWORD *)v198)-- == 1;
                if ( v9 )
                {
                  v199 = sub_18000CC40(v198);
                  (*(void (__fastcall **)(int *))(v199 + 48))(v198);
                }
              }
              v6 = v867;
// Original IDA line 5759 | qword_1803B7970 -> novels.pipeline.short_fiction_runner#142 bytes [4005522, 4005535]: "raw_content"
              v200 = sub_18037F0C0(a1, v867, qword_1803B7970);
              v201 = (int *)v200;
              v202 = a1;
              if ( v200 == 0 )
                goto LABEL_360;
// Original IDA line 5764 | qword_1803B7620 -> novels.pipeline.short_fiction_runner#36 bytes [4003841, 4003848]: "strip"
              v203 = (int *)sub_18037F0C0(a1, v200, qword_1803B7620);
              if ( *v201 >= 0 )
              {
                v9 = (*(_QWORD *)v201)-- == 1;
                if ( v9 )
                {
                  v204 = sub_18000CC40(v201);
                  (*(void (__fastcall **)(int *))(v204 + 48))(v201);
                }
              }
              v202 = a1;
              if ( v203 == nullptr )
              {
LABEL_360:
                sub_18000CE30(v202, &v862);
                v10 = 291;
LABEL_384:
                v4 = v866;
                goto LABEL_385;
              }
              v4 = v866;
              *(_DWORD *)(v866 + 40) = 291;
              v205 = (int *)sub_180376210(a1, v203);
              if ( *v203 >= 0 )
              {
                v9 = (*(_QWORD *)v203)-- == 1;
                if ( v9 )
                {
                  v206 = sub_18000CC40(v203);
                  (*(void (__fastcall **)(int *))(v206 + 48))(v203);
                }
              }
              v207 = a1;
              if ( v205 == nullptr )
                goto LABEL_359;
// Original IDA line 5799 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
              v208 = sub_180375520(a1, v205, qword_1803B7598);
              if ( *v205 >= 0 )
              {
                v9 = (*(_QWORD *)v205)-- == 1;
                if ( v9 )
                {
                  v209 = sub_18000CC40(v205);
                  (*(void (__fastcall **)(int *))(v209 + 48))(v205);
                }
              }
              v207 = a1;
              if ( v208 == 0 )
              {
LABEL_359:
                sub_18000CE30(v207, &v862);
                v10 = 291;
                goto LABEL_385;
              }
              v210 = (int *)sub_180374A40(a1, 3);
              sub_18000CC80(v210, 0, v208);
// Original IDA line 5819 | qword_1803B7978 -> novels.pipeline.short_fiction_runner#143 bytes [4005535, 4005539]: "\n\n"
              v211 = qword_1803B7978;
// Original IDA line 5820 | qword_1803B7978 -> novels.pipeline.short_fiction_runner#143 bytes [4005535, 4005539]: "\n\n"
              if ( *(_DWORD *)qword_1803B7978 != -1 )
// Original IDA line 5821 | qword_1803B7978 -> novels.pipeline.short_fiction_runner#143 bytes [4005535, 4005539]: "\n\n"
                ++*(_DWORD *)qword_1803B7978;
              sub_18000CC80(v210, 1, v211);
// Original IDA line 5823 | qword_1803B7620 -> novels.pipeline.short_fiction_runner#36 bytes [4003841, 4003848]: "strip"
              v212 = sub_18037F0C0(a1, v194, qword_1803B7620);
              v213 = (int *)v212;
              if ( v212 == 0 )
                break;
              *(_DWORD *)(v4 + 40) = 291;
              v214 = (int *)sub_180376210(a1, v212);
              if ( *v213 >= 0 )
              {
                v9 = (*(_QWORD *)v213)-- == 1;
                if ( v9 )
                {
                  v215 = sub_18000CC40(v213);
                  (*(void (__fastcall **)(int *))(v215 + 48))(v213);
                }
              }
              if ( v214 == nullptr )
                break;
// Original IDA line 5840 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
              v216 = sub_180375520(a1, v214, qword_1803B7598);
              if ( *v214 >= 0 )
              {
                v9 = (*(_QWORD *)v214)-- == 1;
                if ( v9 )
                {
                  v217 = sub_18000CC40(v214);
                  (*(void (__fastcall **)(int *))(v217 + 48))(v214);
                }
              }
              if ( v216 == 0 )
                break;
              sub_18000CC80(v210, 2, v216);
// Original IDA line 5853 | qword_1803B7598 -> novels.pipeline.short_fiction_runner#19 bytes [4003601, 4003602]: ""
              v218 = PyUnicode_Join(qword_1803B7598, v210);
              if ( *v210 >= 0 )
              {
                v9 = (*(_QWORD *)v210)-- == 1;
                if ( v9 )
                {
                  v219 = sub_18000CC40(v210);
                  (*(void (__fastcall **)(int *))(v219 + 48))(v210);
                }
              }
              if ( v218 == 0 )
              {
                sub_18000CE30(a1, &v862);
                v10 = 291;
                goto LABEL_385;
              }
              v220 = (int *)v884;
              v884 = (_QWORD *)v218;
              if ( v220 != nullptr && *v220 >= 0 )
              {
                v9 = (*(_QWORD *)v220)-- == 1;
                if ( v9 )
                {
                  v221 = sub_18000CC40(v220);
                  (*(void (__fastcall **)(int *))(v221 + 48))(v220);
                }
              }
              v222 = sub_18015EE30(a1);
              if ( v222 == 0 )
              {
// Original IDA line 5883 | qword_1803B78A8 -> novels.pipeline.short_fiction_runner#117 bytes [4004954, 4004987]: "parse_short_fiction_batch_draft"
                sub_180373340(a1, &v862, qword_1803B78A8);
                v10 = 292;
                goto LABEL_385;
              }
              v223 = v876;
              *(_DWORD *)(v4 + 40) = 292;
              v929 = v223;
              v930 = v218;
// Original IDA line 5891 | qword_1803B78B0 -> novels.pipeline.short_fiction_runner#118 bytes [4004987, 4005008]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4004989, 4005008], "value": "expected_chapters"}]}
              v224 = sub_18037D030(a1, v222, &v930, &v929, qword_1803B78B0);
              v225 = v224;
              if ( v224 == 0 )
              {
                sub_18000CE30(a1, &v862);
                v10 = 292;
                goto LABEL_385;
              }
              v226 = v6;
              v867 = (_QWORD *)v224;
              v6 = (_QWORD *)v224;
              if ( *(int *)v917 >= 0 )
              {
                v9 = (*v917)-- == 1;
                if ( v9 )
                {
                  v227 = sub_18000CC40(v226);
                  (*(void (__fastcall **)(_QWORD *))(v227 + 48))(v226);
                }
              }
              v228 = sub_18015EC10(a1);
              if ( v228 == 0 )
              {
// Original IDA line 5914 | qword_1803B78B8 -> novels.pipeline.short_fiction_runner#119 bytes [4005008, 4005043]: "find_empty_short_fiction_chapters"
                sub_180373340(a1, &v862, qword_1803B78B8);
                goto LABEL_385;
              }
              *(_DWORD *)(v4 + 40) = 293;
              v229 = sub_180376770(a1, v228, v225);
              if ( v229 == 0 )
              {
                sub_18000CE30(a1, &v862);
                goto LABEL_385;
              }
              v230 = v864;
              v7 = (_QWORD *)v229;
              v864 = (_QWORD *)v229;
              if ( *(int *)v916 >= 0 )
              {
                v9 = (*v916)-- == 1;
                if ( v9 )
                {
                  v231 = sub_18000CC40(v230);
                  (*(void (__fastcall **)(_QWORD *))(v231 + 48))(v230);
                }
              }
              if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                goto LABEL_350;
              v129 = (int *)v903;
              v868 = (int *)sub_180038D80(v903);
              v131 = v868;
              if ( v868 == nullptr )
                goto LABEL_349;
            }
            v10 = 291;
            sub_18000CE30(a1, &v862);
            if ( *v210 >= 0 )
            {
              v9 = (*(_QWORD *)v210)-- == 1;
              if ( v9 )
              {
                v232 = sub_18000CC40(v210);
                (*(void (__fastcall **)(int *))(v232 + 48))(v210);
              }
            }
LABEL_385:
            v238 = v862;
            sub_18000CE70(&v862);
            v8 = v865;
            if ( v865 != nullptr && *v865 >= 0 )
            {
              v9 = (*(_QWORD *)v865)-- == 1;
              if ( v9 )
              {
                v239 = sub_18000CC40(v8);
                (*(void (__fastcall **)(int *))(v239 + 48))(v8);
              }
            }
            v240 = v903;
            if ( *(int *)v903 >= 0 )
            {
              v9 = (*v903)-- == 1;
              if ( v9 )
              {
                v241 = sub_18000CC40(v240);
                (*(void (__fastcall **)(_QWORD *))(v241 + 48))(v240);
              }
            }
            v7 = v864;
            v862 = v238;
LABEL_1392:
            v799 = *(_QWORD *)(v862 + 40);
            if ( v799 == 0 )
            {
              v800 = sub_18038E2A0(v4, v10);
              goto LABEL_1397;
            }
            if ( *(_QWORD *)(v799 + 24) != v4 )
            {
              v800 = sub_18038E2A0(v4, v10);
              *(_QWORD *)(v800 + 16) = v799;
              if ( *(_DWORD *)v799 != -1 )
                ++*(_DWORD *)v799;
LABEL_1397:
              sub_180001300(&v862, v800);
            }
            v861 = v901;
            v860 = v885;
            v859 = v887;
            v858 = v899;
            v857 = v877;
            v856 = v898;
            v855 = v894;
            v854 = v900;
            v853 = v884;
            v852 = v883;
            v851 = v882;
            v850 = v893;
            v849 = v896;
            v848 = v897;
            v847 = v911;
            v846 = v880;
            v845 = v910;
            v801 = v875;
            v844 = v920;
            v843 = v886;
            v842 = v913;
            v841 = v871;
            v840 = v888;
            v839 = v912;
            v838 = v878;
            v837 = v876;
            sub_18036D0A0(v4, "ooooooooooooooooooooooooooooooooooo", v875);
            v805 = qword_1803B7450;
            if ( v4 == qword_1803B7450 )
            {
              if ( *(int *)qword_1803B7450 >= 0 )
              {
                v9 = (*(_QWORD *)qword_1803B7450)-- == 1;
                if ( v9 )
                  (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *, int *, int *, int *, int *, int *, int *, int *, _QWORD *, int *, int *, int *, int *, _QWORD *, _QWORD *, int *, _QWORD *, _QWORD *, _QWORD *, int *, int *, int *, int *, int *, int *, int *, _QWORD, int *))(*(_QWORD *)(v805 + 8) + 48LL))(
                    v805,
                    *(_QWORD *)(*(_QWORD *)(v805 + 8) + 48LL),
                    v803,
                    v804,
                    v837,
                    v838,
                    v839,
                    v840,
                    v841,
                    v842,
                    v843,
                    v844,
                    v845,
                    v846,
                    v847,
                    v848,
                    v849,
                    v850,
                    v6,
                    v7,
                    v8,
                    v851,
                    v852,
                    v853,
                    v854,
                    v855,
                    v856,
                    v857,
                    v858,
                    v859,
                    v860,
                    0,
                    v861);
              }
              qword_1803B7450 = 0;
            }
            sub_180001D80(a1, v802);
            v806 = v886;
            v807 = v862;
            v862 = 0;
            if ( v886 != nullptr && *v886 >= 0 )
            {
              v9 = (*(_QWORD *)v886)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v806 + 1) + 48LL))();
            }
            v808 = v920;
            if ( v920 != nullptr && *v920 >= 0 )
            {
              v9 = (*(_QWORD *)v920)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v808 + 1) + 48LL))();
            }
            v809 = v910;
            if ( v910 != nullptr && *v910 >= 0 )
            {
              v9 = (*(_QWORD *)v910)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v809 + 1) + 48LL))();
            }
            v810 = v880;
            if ( v880 != nullptr && *(int *)v880 >= 0 )
            {
              v9 = (*v880)-- == 1;
              if ( v9 )
                (*(void (**)(void))(v810[1] + 48LL))();
            }
            v811 = v911;
            if ( v911 != nullptr && *v911 >= 0 )
            {
              v9 = (*(_QWORD *)v911)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v811 + 1) + 48LL))();
            }
            v812 = v897;
            if ( v897 != nullptr && *v897 >= 0 )
            {
              v9 = (*(_QWORD *)v897)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v812 + 1) + 48LL))();
            }
            v813 = v896;
            if ( v896 != nullptr && *v896 >= 0 )
            {
              v9 = (*(_QWORD *)v896)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v813 + 1) + 48LL))();
            }
            v814 = v893;
            if ( v893 != nullptr && *v893 >= 0 )
            {
              v9 = (*(_QWORD *)v893)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v814 + 1) + 48LL))();
            }
            if ( v6 != nullptr && *(int *)v6 >= 0 )
            {
              v9 = (*v6)-- == 1;
              if ( v9 )
                (*(void (__fastcall **)(_QWORD *))(v6[1] + 48LL))(v6);
            }
            if ( v7 != nullptr && *(int *)v7 >= 0 )
            {
              v9 = (*v7)-- == 1;
              if ( v9 )
                (*(void (__fastcall **)(_QWORD *))(v7[1] + 48LL))(v7);
            }
            if ( v8 != nullptr && *v8 >= 0 )
            {
              v9 = (*(_QWORD *)v8)-- == 1;
              if ( v9 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v8 + 1) + 48LL))(v8);
            }
            v815 = v882;
            if ( v882 != nullptr && *(int *)v882 >= 0 )
            {
              v9 = (*v882)-- == 1;
              if ( v9 )
                (*(void (**)(void))(v815[1] + 48LL))();
            }
            v816 = v883;
            if ( v883 != nullptr && *(int *)v883 >= 0 )
            {
              v9 = (*v883)-- == 1;
              if ( v9 )
                (*(void (**)(void))(v816[1] + 48LL))();
            }
            v817 = v884;
            if ( v884 != nullptr && *(int *)v884 >= 0 )
            {
              v9 = (*v884)-- == 1;
              if ( v9 )
                (*(void (**)(void))(v817[1] + 48LL))();
            }
            v818 = v900;
            if ( v900 != nullptr && *v900 >= 0 )
            {
              v9 = (*(_QWORD *)v900)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v818 + 1) + 48LL))();
            }
            v819 = v894;
            if ( v894 != nullptr && *v894 >= 0 )
            {
              v9 = (*(_QWORD *)v894)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v819 + 1) + 48LL))();
            }
            v820 = v898;
            if ( v898 != nullptr && *v898 >= 0 )
            {
              v9 = (*(_QWORD *)v898)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v820 + 1) + 48LL))();
            }
            v821 = v877;
            if ( v877 != nullptr && *v877 >= 0 )
            {
              v9 = (*(_QWORD *)v877)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v821 + 1) + 48LL))();
            }
            v822 = v899;
            if ( v899 != nullptr && *v899 >= 0 )
            {
              v9 = (*(_QWORD *)v899)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v822 + 1) + 48LL))();
            }
            v823 = v887;
            if ( v887 != nullptr && *v887 >= 0 )
            {
              v9 = (*(_QWORD *)v887)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v823 + 1) + 48LL))();
            }
            v824 = v885;
            if ( v885 != nullptr && *v885 >= 0 )
            {
              v9 = (*(_QWORD *)v885)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v824 + 1) + 48LL))();
            }
            v825 = v901;
            if ( v901 != nullptr && *v901 >= 0 )
            {
              v9 = (*(_QWORD *)v901)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v825 + 1) + 48LL))();
            }
            v826 = v889;
            if ( v889 != nullptr && *v889 >= 0 )
            {
              v9 = (*(_QWORD *)v889)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v826 + 1) + 48LL))();
            }
            v827 = v890;
            if ( v890 != nullptr && *v890 >= 0 )
            {
              v9 = (*(_QWORD *)v890)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v827 + 1) + 48LL))();
            }
            v828 = v891;
            if ( v891 != nullptr && *v891 >= 0 )
            {
              v9 = (*(_QWORD *)v891)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v828 + 1) + 48LL))();
            }
            v829 = v892;
            if ( v892 != nullptr && *v892 >= 0 )
            {
              v9 = (*(_QWORD *)v892)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v829 + 1) + 48LL))();
            }
            v862 = v807;
            if ( *v801 >= 0 )
            {
              v9 = (*(_QWORD *)v801)-- == 1;
              if ( v9 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v801 + 1) + 48LL))(v801);
            }
            v830 = v881;
            if ( *v881 >= 0 )
            {
              v9 = (*(_QWORD *)v881)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v830 + 1) + 48LL))();
            }
            v831 = v876;
            if ( *v876 >= 0 )
            {
              v9 = (*(_QWORD *)v876)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v831 + 1) + 48LL))();
            }
            v832 = v878;
            if ( *v878 >= 0 )
            {
              v9 = (*(_QWORD *)v878)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v832 + 1) + 48LL))();
            }
            v833 = v912;
            if ( *v912 >= 0 )
            {
              v9 = (*(_QWORD *)v912)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v833 + 1) + 48LL))();
            }
            v834 = v888;
            if ( *v888 >= 0 )
            {
              v9 = (*(_QWORD *)v888)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v834 + 1) + 48LL))();
            }
            v835 = v871;
            if ( *v871 >= 0 )
            {
              v9 = (*(_QWORD *)v871)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v835 + 1) + 48LL))();
            }
            v836 = v913;
            if ( *v913 >= 0 )
            {
              v9 = (*(_QWORD *)v913)-- == 1;
              if ( v9 )
                (*(void (**)(void))(*((_QWORD *)v836 + 1) + 48LL))();
            }
            sub_18000CE40(a1, &v862);
            return nullptr;
          }
        }
      }
      sub_18000CE30(a1, &v862);
      goto LABEL_204;
    }
  }
  sub_18000CE30(v92, &v862);
  v6 = v867;
  v4 = v81;
  v7 = v867;
  v8 = (int *)v867;
  goto LABEL_1392;
}
