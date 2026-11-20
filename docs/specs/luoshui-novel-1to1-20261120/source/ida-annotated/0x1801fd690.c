// Analyst annotations only; original export: source/ida/0x1801fd690.c
// qword comments derive module constant slot + original file range; not runtime values.
// write access to const memory has been detected, the output may be wrong!
// Original IDA line 2 | sub_1801FD690 -> AgentSessionService.confirm_action
int *__fastcall sub_1801FD690(__int64 a1, __int64 a2, _QWORD *a3)
{
  int *v5; // r12
  int *v6; // r14
  __int64 v7; // rbx
  bool v8; // zf
  __int64 v9; // rax
  __int64 v10; // r13
  int *v11; // r15
  __int64 v12; // rax
  int *v13; // rbx
  unsigned int v14; // edi
  int *v15; // r9
  __int64 v16; // r14
  __int64 v17; // rax
  __int64 v18; // rdx
  __int64 v19; // rax
  int *v20; // rdi
  int v21; // eax
  __int64 v22; // rdx
  int *v23; // r14
  __int64 v24; // rax
  int *v25; // rbx
  int *v26; // r14
  __int64 v27; // rax
  __int64 v28; // rax
  bool v29; // sf
  __int64 v30; // rax
  __int64 v31; // r15
  int *v32; // rbx
  __int64 v33; // rax
  int *v34; // rdi
  __int64 v35; // rax
  int v36; // r14d
  __int64 v37; // rax
  __int64 v38; // rax
  __int64 v39; // rax
  __int64 v40; // r15
  __int64 v41; // rax
  int v42; // eax
  __int64 v43; // rdi
  __int64 v44; // rax
  __int64 v45; // r15
  int v46; // eax
  __int64 v47; // rax
  int *v48; // rdi
  __int64 v49; // rcx
  int *v50; // rbx
  __int64 v51; // rax
  int *v52; // r14
  __int64 v53; // rax
  __int64 v54; // rax
  int *v55; // r15
  __int64 v56; // rax
  __int64 v57; // rax
  int *v58; // r14
  __int64 v59; // rax
  __int64 v60; // rax
  __int64 v61; // rax
  __int64 v62; // rax
  int *v63; // rbx
  __int64 v64; // r9
  int *v65; // r14
  __int64 v66; // rax
  __int64 v67; // rax
  __int64 v68; // rdi
  __int64 v69; // rbx
  int *v70; // r15
  __int64 v71; // r13
  int *v72; // r12
  int *v73; // rbx
  __int64 v74; // rax
  int v75; // eax
  __int64 v76; // rax
  __int64 v77; // rax
  __int64 v78; // r9
  __int64 v79; // rdx
  __int64 v80; // rax
  __int64 v81; // r13
  int *v82; // r12
  int *v83; // rbx
  __int64 v84; // rax
  __int64 v85; // rax
  __int64 v86; // r12
  __int64 v87; // rdx
  __int64 v88; // rax
  int *v89; // rbx
  __int64 v90; // rax
  __int64 v91; // rax
  __int64 v92; // rax
  _QWORD *v93; // rdx
  __int64 v94; // rax
  __int64 v95; // rbx
  __int64 v96; // rax
  __int64 v97; // rax
  int *v98; // rbx
  __int64 v99; // rax
  int *v100; // rbx
  __int64 v101; // rax
  int *v102; // rbx
  __int64 v103; // rax
  int *v104; // rbx
  __int64 v105; // rax
  int *v106; // rbx
  __int64 v107; // rax
  int *v108; // rbx
  __int64 v109; // rax
  int *v110; // rbx
  __int64 v111; // rax
  int *v112; // rbx
  __int64 v113; // rax
  int *v114; // rbx
  __int64 v115; // rax
  int *v116; // rbx
  __int64 v117; // rax
  int *v118; // rbx
  __int64 v119; // rax
  int *v120; // rbx
  __int64 v121; // rax
  int *v122; // rbx
  __int64 v123; // rax
  int *v124; // rbx
  __int64 v125; // rax
  int *v126; // rbx
  __int64 v127; // rax
  int *v128; // rbx
  __int64 v129; // rax
  int *v130; // rbx
  __int64 v131; // rax
  int *v132; // rbx
  __int64 v133; // rax
  _QWORD *v134; // rbx
  __int64 v135; // rax
  int *v136; // rbx
  __int64 v137; // rax
  int *v138; // rbx
  __int64 v139; // rax
  __int64 v140; // rax
  int *v141; // rbx
  __int64 v142; // rax
  int *v143; // rbx
  __int64 v144; // rax
  __int64 v146; // rax
  __int64 v147; // rax
  __int64 v148; // rbx
  __int64 v149; // r14
  int v150; // eax
  unsigned int v151; // ebx
  __int64 v152; // rax
  int *v153; // rbx
  int *v154; // rdi
  __int64 v155; // rax
  int *v156; // rax
  __int64 v157; // rax
  int *v158; // r13
  __int64 v159; // rax
  __int64 v160; // rdx
  __int64 v161; // rax
  int *v162; // rbx
  __int64 v163; // rcx
  char v164; // di
  __int64 v165; // rax
  __int64 v166; // rax
  int *v167; // rbx
  char v168; // di
  __int64 v169; // rax
  __int64 v170; // rax
  int *v171; // rbx
  __int64 v172; // rax
  int *v173; // r12
  __int64 v174; // rax
  _DWORD *v175; // rbx
  __int64 v176; // rax
  int *v177; // rdi
  __int64 v178; // rax
  int *v179; // rbx
  __int64 v180; // r15
  __int64 v181; // rax
  __int64 v182; // rax
  __int64 v183; // rax
  int *v184; // rcx
  __int64 v185; // rax
  __int64 v186; // r8
  __int64 v187; // rax
  int *v188; // rbx
  __int64 v189; // rax
  __int64 v190; // rax
  __int64 v191; // rax
  int *v192; // rdi
  __int64 v193; // rax
  __int64 v194; // rax
  __int64 v195; // rax
  __int64 v196; // rax
  __int64 v197; // rax
  __int64 v198; // rax
  __int64 v199; // rax
  __int64 v200; // rax
  int *v201; // rbx
  __int64 v202; // rcx
  __int64 v203; // rbx
  int *v204; // rdi
  __int64 v205; // rax
  int *v206; // r13
  __int64 v207; // rax
  int *v208; // rdi
  int *i; // r15
  int *v210; // rdx
  int *v211; // rcx
  int *v212; // rbx
  int *v213; // rdi
  __int64 v214; // rax
  int v215; // eax
  int *v216; // rbx
  int v217; // eax
  __int64 v218; // rax
  __int64 v219; // rbx
  __int64 v220; // rdi
  __int64 v221; // rdx
  __int64 v222; // rax
  __int64 v223; // rax
  int *v224; // rbx
  int *v225; // rax
  __int64 v226; // rax
  __int64 v227; // rax
  int *v228; // r13
  int *v229; // r15
  __int64 v230; // rax
  __int64 v231; // rax
  int *v232; // r12
  __int64 v233; // rax
  _QWORD *v234; // r12
  __int64 v235; // rax
  __int64 v236; // rax
  __int64 v237; // rax
  int *v238; // rdi
  __int64 v239; // rax
  __int64 v240; // rax
  __int64 v241; // rax
  int *v242; // rdi
  __int64 v243; // r9
  int *v244; // r14
  __int64 v245; // rax
  __int64 v246; // rax
  __int64 v247; // rdi
  __int64 v248; // r13
  __int64 v249; // r12
  _QWORD *v250; // rdx
  int *v251; // r15
  int *v252; // rdi
  __int64 v253; // rax
  __int64 v254; // r12
  __int64 v255; // rdx
  __int64 v256; // rax
  int *v257; // r15
  __int64 v258; // rax
  __int64 v259; // rax
  __int64 v260; // rbx
  __int64 v261; // rax
  __int64 v262; // rax
  int **v263; // rdi
  __int64 v264; // rax
  __int64 v265; // rax
  int v266; // eax
  int v267; // edx
  int v268; // r8d
  __int64 v269; // rax
  int *v270; // rdi
  __int64 v271; // rax
  int *v272; // r13
  __int64 v273; // rax
  int v274; // edx
  __int64 v275; // r12
  int v276; // r8d
  __int64 v277; // rax
  int *v278; // rdi
  __int64 v279; // rax
  int *v280; // r15
  __int64 v281; // rax
  __int64 v282; // rax
  int *v283; // r14
  int *v284; // rdi
  __int64 v285; // rax
  int v286; // eax
  __int64 v287; // rax
  __int64 v288; // rax
  __int64 v289; // rax
  int *v290; // r14
  __int64 v291; // rax
  int v292; // eax
  __int64 v293; // rax
  __int64 v294; // rax
  int *v295; // r14
  __int64 v296; // rax
  int v297; // eax
  __int64 v298; // rax
  __int64 v299; // rax
  int *v300; // r14
  __int64 v301; // rax
  int v302; // eax
  __int64 v303; // rax
  int *v304; // rbx
  __int64 v305; // rax
  int *v306; // rdi
  __int64 v307; // rax
  __int64 v308; // r12
  __int64 v309; // rax
  __int64 v310; // rcx
  int *v311; // r15
  __int64 v312; // rax
  __int64 v313; // rax
  __int64 v314; // rcx
  int v315; // eax
  __int64 v316; // rcx
  __int64 v317; // rax
  __int64 v318; // rax
  int *v319; // r15
  __int64 v320; // rax
  int v321; // eax
  __int64 v322; // rax
  int v323; // eax
  int *v324; // r15
  __int64 v325; // rax
  int *v326; // r14
  __int64 v327; // rax
  __int64 v328; // r15
  __int64 v329; // rax
  __int64 v330; // rax
  __int64 v331; // rax
  int *v332; // rdi
  __int64 v333; // r9
  int *v334; // r14
  __int64 v335; // rax
  __int64 v336; // rax
  __int64 v337; // rax
  __int64 v338; // r14
  __int64 v339; // rax
  int *v340; // rbx
  int *v341; // rdi
  __int64 v342; // rax
  int *v343; // rax
  __int64 v344; // rax
  int *v345; // r15
  __int64 v346; // rax
  __int64 v347; // rax
  int *v348; // r14
  __int64 v349; // rax
  int v350; // eax
  __int64 v351; // rax
  int *v352; // r14
  __int64 v353; // rax
  char v354; // di
  __int64 v355; // rax
  __int64 v356; // rax
  int *v357; // r14
  __int64 v358; // rax
  int v359; // r14d
  __int64 v360; // rax
  __int64 v361; // rax
  int *v362; // rdi
  int *v363; // r14
  __int64 v364; // rax
  int *v365; // rdi
  __int64 v366; // rax
  char v367; // r14
  __int64 v368; // rax
  __int64 v369; // rax
  int *v370; // r14
  __int64 v371; // rax
  int v372; // r14d
  __int64 v373; // rax
  __int64 v374; // rax
  int *v375; // rbx
  __int64 v376; // rcx
  __int64 v377; // rax
  int *v378; // rdi
  __int64 v379; // rax
  __int64 v380; // rbx
  int *v381; // r15
  int *v382; // r12
  __int64 v383; // rbx
  int *v384; // r13
  int *v385; // rcx
  int *v386; // r13
  int *v387; // rdi
  __int64 v388; // rax
  __int64 v389; // rax
  int *v390; // rdi
  int *v391; // rbx
  int *v392; // r15
  int *v393; // r12
  __int64 v394; // rax
  int *v395; // r14
  __int64 v396; // rax
  int v397; // eax
  int *v398; // rdi
  int v399; // eax
  __int64 v400; // rax
  __int64 v401; // rax
  int v402; // eax
  int *v403; // rbx
  int v404; // eax
  __int64 v405; // rax
  __int64 v406; // rax
  __int64 v407; // rbx
  int *v408; // rdi
  __int64 v409; // rbx
  int v410; // eax
  __int64 v411; // r8
  char v412; // al
  __int64 v413; // rcx
  __int64 v414; // rax
  __int64 v415; // rax
  _QWORD *v416; // rbx
  __int64 v417; // rax
  __int64 v418; // rax
  __int64 v419; // rax
  __int64 v420; // rax
  int *v421; // rbx
  __int64 v422; // rax
  __int64 v423; // rbx
  __int64 v424; // rax
  __int64 v425; // rax
  __int64 v426; // rax
  __int64 v427; // rbx
  __int64 v428; // rax
  int *v429; // r14
  __int64 v430; // rax
  __int64 v431; // rax
  _QWORD *v432; // r12
  int *v433; // r15
  __int64 v434; // rax
  __int64 v435; // rcx
  __int64 v436; // rax
  int *v437; // rcx
  int *v438; // r14
  __int64 v439; // rax
  __int64 v440; // rax
  __int64 v441; // rax
  int *v442; // rdi
  __int64 v443; // r9
  int *v444; // r14
  __int64 v445; // rax
  __int64 v446; // rax
  __int64 v447; // rdi
  __int64 v448; // r13
  __int64 v449; // rdi
  __int64 v450; // r12
  __int64 v451; // rdx
  __int64 v452; // rax
  int *v453; // rdi
  __int64 v454; // rax
  int *v455; // rcx
  __int64 v456; // rax
  __int64 v457; // rbx
  int *v458; // rdi
  __int64 v459; // rax
  int *v460; // rdi
  __int64 v461; // rax
  _QWORD *v462; // rdi
  __int64 v463; // rax
  int *v464; // rdi
  __int64 v465; // rax
  __int64 v466; // rax
  __int64 v467; // rax
  int *v468; // rbx
  int *v469; // r12
  __int64 v470; // rax
  int *v471; // r13
  __int64 v472; // rax
  _DWORD *v473; // rbx
  __int64 v474; // rax
  int *v475; // r14
  int *v476; // r15
  __int64 v477; // rax
  __int64 v478; // r14
  __int64 v479; // rax
  __int64 v480; // rax
  __int64 v481; // r8
  __int64 v482; // rax
  __int64 v483; // r14
  __int64 v484; // rax
  __int64 v485; // rax
  int *v486; // rcx
  __int64 v487; // rax
  int *v488; // r13
  __int64 v489; // rbx
  int *v490; // r15
  _DWORD *j; // r14
  int *v492; // rdx
  int *v493; // rcx
  _QWORD *v494; // rbx
  int *v495; // rdi
  __int64 v496; // rax
  int v497; // eax
  int *v498; // rbx
  int v499; // eax
  __int64 v500; // rax
  __int64 v501; // rbx
  __int64 v502; // rdi
  __int64 v503; // rax
  __int64 v504; // rax
  int *v505; // rbx
  _QWORD *v506; // rax
  __int64 v507; // rax
  __int64 v508; // rax
  int *v509; // rdi
  int *v510; // r13
  int *v511; // rdx
  int *v512; // r14
  __int64 v513; // rax
  __int64 v514; // rax
  __int64 v515; // rax
  __int64 v516; // rax
  int *v517; // rdi
  __int64 v518; // rax
  __int64 v519; // rax
  __int64 v520; // rax
  __int64 v521; // rax
  int *v522; // rdi
  __int64 v523; // r15
  __int64 v524; // r9
  int *v525; // r14
  __int64 v526; // rax
  __int64 v527; // rax
  __int64 v528; // rdi
  __int64 v529; // rbx
  __int64 v530; // r12
  __int64 v531; // rdx
  __int64 v532; // rax
  int *v533; // r15
  __int64 v534; // rax
  __int64 v535; // rax
  __int64 v536; // rbx
  __int64 v537; // rax
  __int64 v538; // rax
  int *v539; // rdi
  __int64 v540; // rax
  __int64 v541; // rax
  int v542; // eax
  __int64 v543; // rax
  int *v544; // rbx
  int *v545; // rdi
  __int64 v546; // rax
  __int64 v547; // rax
  int *v548; // r12
  __int64 v549; // rax
  __int64 v550; // rdx
  __int64 v551; // rax
  int *v552; // r14
  __int64 v553; // rax
  int v554; // eax
  __int64 v555; // rax
  int *v556; // r14
  __int64 v557; // rax
  __int64 v558; // r15
  char v559; // di
  __int64 v560; // rax
  __int64 v561; // rax
  int *v562; // r14
  __int64 v563; // rax
  int v564; // eax
  __int64 v565; // rax
  int *v566; // r14
  __int64 v567; // rax
  char v568; // di
  __int64 v569; // rax
  __int64 v570; // rax
  int *v571; // r14
  __int64 v572; // rax
  int v573; // eax
  __int64 v574; // rax
  int *v575; // r15
  __int64 v576; // rax
  __int64 v577; // rax
  int *v578; // r14
  __int64 v579; // rax
  int v580; // eax
  __int64 v581; // rbx
  __int64 v582; // rax
  int v583; // eax
  __int64 v584; // r14
  __int64 v585; // r15
  __int64 v586; // rax
  int *v587; // rbx
  int *v588; // rdi
  __int64 v589; // rax
  int *v590; // rbx
  __int64 v591; // rax
  int *v592; // rdi
  __int64 v593; // rax
  char v594; // bl
  __int64 v595; // rax
  __int64 v596; // rax
  int *v597; // rbx
  __int64 v598; // rax
  __int64 v599; // rax
  _DWORD *v600; // rbx
  __int64 v601; // rax
  int *v602; // r15
  __int64 v603; // rax
  int *v604; // r12
  int *v605; // r13
  __int64 v606; // rax
  __int64 v607; // r12
  __int64 v608; // rax
  __int64 v609; // rax
  __int64 v610; // rax
  __int64 v611; // r8
  int *v612; // r12
  __int64 v613; // rax
  _QWORD *v614; // rax
  int *v615; // r15
  __int64 v616; // rax
  __int64 v617; // rax
  int *v618; // r14
  __int64 v619; // rax
  __int64 v620; // rax
  __int64 v621; // rax
  __int64 v622; // rax
  int *v623; // rbx
  __int64 v624; // rax
  __int64 v625; // rdi
  __int64 v626; // rbx
  __int64 v627; // r12
  __int64 v628; // rax
  __int64 v629; // rax
  int *v630; // rdi
  __int64 v631; // rax
  __int64 v632; // rax
  int *v633; // rdi
  __int64 v634; // rax
  __int64 v635; // rax
  __int64 v636; // rax
  __int64 v637; // rax
  int *v638; // rdi
  __int64 v639; // r15
  __int64 v640; // r9
  int *v641; // r14
  __int64 v642; // rax
  __int64 v643; // rax
  __int64 v644; // rdi
  __int64 v645; // rbx
  __int64 v646; // r12
  __int64 v647; // rdx
  __int64 v648; // rax
  int *v649; // r15
  __int64 v650; // rax
  int v651; // eax
  __int64 v652; // rax
  int *v653; // rbx
  int *v654; // rdi
  __int64 v655; // rax
  int *v656; // r14
  __int64 v657; // rax
  int *v658; // r15
  __int64 v659; // rax
  __int64 v660; // rax
  __int64 v661; // rax
  int *v662; // r12
  __int64 v663; // rax
  __int64 v664; // rax
  __int64 v665; // rax
  __int64 v666; // rax
  int *v667; // rbx
  __int64 v668; // rax
  __int64 v669; // rax
  _DWORD *v670; // rbx
  __int64 v671; // rax
  __int64 v672; // rax
  int *v673; // r12
  int *v674; // r13
  __int64 v675; // rax
  __int64 v676; // r12
  __int64 v677; // rax
  __int64 v678; // r8
  int *v679; // r12
  __int64 v680; // rax
  _QWORD *v681; // rax
  int *v682; // r15
  __int64 v683; // rax
  __int64 v684; // rax
  int *v685; // r14
  __int64 v686; // rax
  __int64 v687; // rax
  __int64 v688; // rax
  __int64 v689; // rax
  int *v690; // rdi
  __int64 v691; // r12
  __int64 v692; // r9
  int *v693; // r14
  __int64 v694; // rax
  __int64 v695; // rax
  __int64 v696; // rdi
  __int64 v697; // rbx
  __int64 v698; // rdx
  __int64 v699; // rax
  int *v700; // rbx
  int v701; // edi
  __int64 v702; // rax
  __int64 v703; // rax
  __int64 v704; // rax
  int *v705; // rbx
  int *v706; // rdi
  __int64 v707; // rax
  int *v708; // r13
  __int64 v709; // rax
  int *v710; // r12
  __int64 v711; // rax
  __int64 v712; // rax
  __int64 v713; // rax
  int v714; // eax
  __int64 v715; // rax
  int *v716; // rbx
  int v717; // edi
  __int64 v718; // rax
  __int64 v719; // r12
  __int64 v720; // rax
  int *v721; // rbx
  int *v722; // rdi
  __int64 v723; // rax
  int *v724; // r14
  __int64 v725; // rax
  int *v726; // rax
  __int64 v727; // rax
  int *v728; // r15
  __int64 v729; // rax
  __int64 v730; // rax
  __int64 v731; // rdx
  __int64 v732; // rax
  int *v733; // rbx
  int *v734; // rdi
  __int64 v735; // rax
  __int64 v736; // r12
  __int64 v737; // rax
  int *v738; // r14
  __int64 v739; // rax
  __int64 v740; // r13
  __int64 v741; // rax
  __int64 v742; // rax
  __int64 v743; // rdx
  __int64 v744; // r14
  int *v745; // rdi
  __int64 v746; // rax
  __int64 v747; // rax
  __int64 v748; // rdx
  __int64 v749; // rax
  int *v750; // rbx
  __int64 v751; // rax
  int *v752; // r15
  __int64 v753; // rax
  __int64 v754; // rdx
  int *v755; // r12
  __int64 v756; // rax
  __int64 v757; // rax
  int *v758; // rbx
  __int64 v759; // rax
  __int64 v760; // rax
  __int64 v761; // rax
  __int64 v762; // rax
  __int64 v763; // rax
  __int64 v764; // rax
  __int64 v765; // rax
  int *v766; // r13
  __int64 v767; // rax
  __int64 v768; // rax
  __int64 v769; // rax
  __int64 v770; // rax
  __int64 v771; // rax
  __int64 v772; // rcx
  __int64 v773; // rax
  __int64 v774; // rax
  __int64 v775; // rax
  int v776; // eax
  __int64 v777; // rax
  __int64 v778; // rax
  __int64 v779; // rax
  int *v780; // r14
  __int64 v781; // rax
  __int64 v782; // rax
  int *v783; // r13
  __int64 v784; // rax
  _QWORD *v785; // rdx
  __int64 v786; // rax
  int v787; // r13d
  __int64 v788; // rax
  __int64 v789; // rax
  __int64 v790; // rax
  int *v791; // rbx
  __int64 v792; // rcx
  __int64 v793; // rbx
  int *v794; // rdi
  __int64 v795; // rax
  __int64 v796; // rax
  int *v797; // rbx
  __int64 v798; // rax
  _DWORD *v799; // rdi
  int *k; // r13
  int *v801; // rdx
  int *v802; // rcx
  int *v803; // rbx
  int *v804; // rdi
  __int64 v805; // rax
  int v806; // eax
  int *v807; // rbx
  int v808; // eax
  __int64 v809; // rax
  __int64 v810; // rbx
  __int64 v811; // rdi
  __int64 v812; // rdx
  __int64 v813; // rax
  __int64 v814; // rax
  int *v815; // rbx
  int *v816; // rax
  __int64 v817; // rax
  __int64 v818; // rax
  int *v819; // rdi
  int *v820; // r13
  __int64 v821; // rax
  __int64 v822; // rax
  __int64 v823; // rax
  _QWORD *v824; // rdi
  __int64 v825; // rax
  int *v826; // rdi
  __int64 v827; // rax
  __int64 v828; // rax
  __int64 v829; // rax
  __int64 v830; // rax
  int *v831; // rdi
  __int64 v832; // r12
  __int64 v833; // r9
  int *v834; // r14
  __int64 v835; // rax
  __int64 v836; // rax
  __int64 v837; // rdi
  __int64 v838; // rbx
  __int64 v839; // rax
  int *v840; // rbx
  int *v841; // rdi
  __int64 v842; // rax
  __int64 v843; // rax
  int *v844; // r12
  __int64 v845; // rax
  int *v846; // r13
  __int64 v847; // rax
  __int64 v848; // rax
  __int64 v849; // rax
  __int64 v850; // r12
  __int64 v851; // rdx
  __int64 v852; // rax
  int *v853; // rdi
  __int64 v854; // rax
  __int64 v855; // rbx
  int *v856; // rdi
  __int64 v857; // rax
  _QWORD *v858; // rdi
  __int64 v859; // rax
  int **v860; // rdi
  __int64 v861; // rax
  __int64 v862; // rax
  int v863; // eax
  __int64 v864; // rax
  int *v865; // rbx
  int v866; // edi
  __int64 v867; // rax
  __int64 v868; // rax
  int *v869; // rdi
  __int64 v870; // r15
  __int64 v871; // rax
  __int64 v872; // rax
  int *v873; // rbx
  int *v874; // rdi
  __int64 v875; // rax
  int *v876; // r14
  __int64 v877; // rax
  int v878; // eax
  __int64 v879; // rax
  __int64 v880; // rax
  int *v881; // r15
  int *v882; // rax
  __int64 v883; // rax
  int *v884; // r12
  __int64 v885; // rax
  __int64 v886; // rax
  __int64 v887; // rax
  __int64 v888; // rax
  int *v889; // rbx
  int *v890; // rdi
  __int64 v891; // rax
  __int64 v892; // rax
  int *v893; // r14
  __int64 v894; // rax
  __int64 v895; // r13
  __int64 v896; // rax
  __int64 v897; // rax
  __int64 v898; // rax
  int *v899; // r14
  __int64 v900; // r15
  int *v901; // rdi
  __int64 v902; // rax
  int v903; // eax
  __int64 v904; // rax
  __int64 v905; // rbx
  __int64 v906; // rax
  __int64 v907; // r12
  __int64 v908; // rax
  int *v909; // r14
  __int64 v910; // rax
  int *v911; // r15
  __int64 v912; // rax
  int *v913; // r14
  __int64 v914; // rax
  __int64 v915; // rax
  int v916; // edi
  __int64 v917; // rax
  _DWORD *v918; // r15
  _QWORD *v919; // r12
  int *v920; // r14
  __int64 v921; // rax
  int v922; // eax
  __int64 v923; // rax
  __int64 v924; // r15
  __int64 v925; // rax
  int *v926; // rdi
  int *v927; // r14
  __int64 v928; // rax
  __int64 v929; // rax
  __int64 v930; // rax
  __int64 v931; // rax
  __int64 v932; // rax
  int *v933; // rbx
  __int64 v934; // rax
  int *v935; // rbx
  __int64 v936; // rax
  int *v937; // rdi
  __int64 v938; // rax
  __int64 v939; // rax
  __int64 v940; // rax
  __int64 v941; // rax
  __int64 v942; // rax
  int v943; // eax
  __int64 v944; // rcx
  __int64 v945; // rax
  int *v946; // rbx
  __int64 v947; // rcx
  __int64 v948; // rbx
  int *v949; // rdi
  __int64 v950; // rax
  int *v951; // r12
  __int64 v952; // rax
  _DWORD *v953; // r13
  _DWORD *v954; // r15
  int *v955; // rdx
  int *v956; // rcx
  _QWORD *v957; // rbx
  int *v958; // rdi
  __int64 v959; // rax
  int v960; // eax
  int *v961; // rbx
  int v962; // eax
  __int64 v963; // rax
  __int64 v964; // rbx
  __int64 v965; // rdi
  __int64 v966; // rax
  __int64 v967; // rax
  int *v968; // rbx
  _QWORD *v969; // rax
  __int64 v970; // rax
  __int64 v971; // rax
  int *v972; // rdi
  int *v973; // r12
  _QWORD *v974; // rdx
  int *v975; // r15
  __int64 v976; // rax
  __int64 v977; // rax
  _QWORD *v978; // rdi
  __int64 v979; // rax
  __int64 v980; // rax
  int *v981; // rdi
  __int64 v982; // rax
  __int64 v983; // rax
  __int64 v984; // rax
  __int64 v985; // rax
  int *v986; // rdi
  __int64 v987; // r12
  __int64 v988; // r9
  int *v989; // r14
  __int64 v990; // rax
  __int64 v991; // rax
  __int64 v992; // rdi
  __int64 v993; // rbx
  __int64 v994; // rax
  int *v995; // rbx
  int *v996; // rdi
  __int64 v997; // rax
  __int64 v998; // rax
  int *v999; // r12
  __int64 v1000; // rax
  int *v1001; // r13
  __int64 v1002; // rax
  __int64 v1003; // rax
  __int64 v1004; // rax
  __int64 v1005; // rbx
  __int64 v1006; // rax
  __int64 v1007; // rax
  int *v1008; // rdi
  __int64 v1009; // rax
  __int64 v1010; // rax
  int v1011; // eax
  __int64 v1012; // rax
  int *v1013; // rbx
  int v1014; // edi
  __int64 v1015; // rax
  __int64 v1016; // rax
  int *v1017; // rdi
  __int64 v1018; // r15
  __int64 v1019; // rax
  int v1020; // eax
  __int64 v1021; // rax
  int *v1022; // rbx
  int *v1023; // rdi
  __int64 v1024; // rax
  int *v1025; // r14
  __int64 v1026; // rax
  int *v1027; // r15
  __int64 v1028; // rax
  __int64 v1029; // rax
  int *v1030; // r13
  __int64 v1031; // rax
  __int64 v1032; // rax
  __int64 v1033; // rax
  __int64 v1034; // rax
  __int64 v1035; // rcx
  __int64 v1036; // rax
  __int64 v1037; // rax
  __int64 v1038; // rax
  int v1039; // eax
  __int64 v1040; // rax
  __int64 v1041; // rax
  __int64 v1042; // rax
  __int64 v1043; // rax
  __int64 v1044; // rax
  __int64 v1045; // rax
  __int64 v1046; // rax
  __int64 v1047; // rax
  __int64 v1048; // rax
  __int64 v1049; // rax
  __int64 v1050; // rax
  _QWORD *v1051; // rcx
  __int64 v1052; // rax
  __int64 v1053; // rcx
  __int64 v1054; // rax
  __int64 v1055; // rax
  int v1056; // eax
  __int64 v1057; // rax
  __int64 v1058; // rax
  __int64 v1059; // rax
  __int64 v1060; // rax
  __int64 v1061; // rax
  __int64 v1062; // rax
  __int64 v1063; // rax
  _QWORD *v1064; // rcx
  __int64 v1065; // rax
  int v1066; // eax
  __int64 v1067; // rax
  __int64 v1068; // rax
  __int64 v1069; // rax
  int *v1070; // r12
  __int64 v1071; // rax
  __int64 v1072; // rax
  __int64 v1073; // rax
  __int64 v1074; // rax
  __int64 v1075; // rax
  __int64 v1076; // rax
  int *v1077; // r15
  __int64 v1078; // r12
  __int64 v1079; // rax
  int *v1080; // rbx
  int *v1081; // rdi
  __int64 v1082; // rax
  int *v1083; // r14
  __int64 v1084; // rax
  __int64 v1085; // r13
  __int64 v1086; // rax
  __int64 v1087; // rax
  __int64 v1088; // rax
  int *v1089; // rbx
  __int64 v1090; // rax
  __int64 v1091; // rax
  _DWORD *v1092; // rbx
  __int64 v1093; // rax
  __int64 v1094; // rax
  int *v1095; // rbx
  __int64 v1096; // r15
  __int64 v1097; // rax
  __int64 v1098; // r8
  __int64 v1099; // rax
  int *v1100; // rbx
  __int64 v1101; // r15
  __int64 v1102; // rax
  __int64 v1103; // r8
  __int64 v1104; // rax
  int *v1105; // rbx
  __int64 v1106; // rax
  __int64 v1107; // rax
  __int64 v1108; // rax
  int *v1109; // rdi
  __int64 v1110; // rax
  __int64 v1111; // rax
  __int64 v1112; // rax
  __int64 v1113; // rax
  __int64 v1114; // rax
  __int64 v1115; // rax
  __int64 v1116; // rax
  __int64 v1117; // rax
  int *v1118; // rbx
  __int64 v1119; // rcx
  __int64 v1120; // rbx
  int *v1121; // rdi
  __int64 v1122; // rax
  int *v1123; // r13
  __int64 v1124; // rax
  _DWORD *v1125; // rdi
  int *v1126; // r15
  int *v1127; // rdx
  int *v1128; // rcx
  int *v1129; // rbx
  int *v1130; // rdi
  __int64 v1131; // rax
  int v1132; // eax
  int *v1133; // rbx
  int v1134; // eax
  __int64 v1135; // rax
  __int64 v1136; // rbx
  __int64 v1137; // rdi
  __int64 v1138; // rdx
  __int64 v1139; // rax
  __int64 v1140; // rax
  int **v1141; // rbx
  int *v1142; // rax
  __int64 v1143; // rax
  __int64 v1144; // rax
  int *v1145; // r13
  int *v1146; // r15
  __int64 v1147; // rax
  __int64 v1148; // rax
  _QWORD *v1149; // r12
  __int64 v1150; // rax
  int *v1151; // r12
  __int64 v1152; // rax
  __int64 v1153; // rax
  int *v1154; // rdi
  __int64 v1155; // rax
  __int64 v1156; // rax
  __int64 v1157; // rax
  __int64 v1158; // rax
  int *v1159; // rdi
  __int64 v1160; // r12
  __int64 v1161; // r9
  int *v1162; // r14
  __int64 v1163; // rax
  __int64 v1164; // rax
  __int64 v1165; // rdi
  __int64 v1166; // rbx
  __int64 v1167; // rax
  int *v1168; // rbx
  int *v1169; // rdi
  __int64 v1170; // rax
  __int64 v1171; // rax
  int *v1172; // r12
  __int64 v1173; // rax
  int *v1174; // r13
  __int64 v1175; // rax
  __int64 v1176; // rax
  __int64 v1177; // rax
  __int64 v1178; // rbx
  __int64 v1179; // rax
  __int64 v1180; // rax
  int **v1181; // rdi
  __int64 v1182; // rax
  __int64 v1183; // rax
  int v1184; // eax
  __int64 v1185; // rax
  int *v1186; // rbx
  int v1187; // edi
  __int64 v1188; // rax
  __int64 v1189; // rax
  int *v1190; // r14
  __int64 v1191; // rax
  int v1192; // eax
  __int64 v1193; // rax
  __int64 v1194; // rax
  int *v1195; // r14
  __int64 v1196; // rax
  int v1197; // eax
  __int64 v1198; // rax
  __int64 v1199; // r15
  __int64 v1200; // rax
  __int64 v1201; // rax
  int *v1202; // rbx
  __int64 v1203; // rax
  __int64 v1204; // rax
  __int64 v1205; // r12
  __int64 v1206; // r13
  __int64 v1207; // rax
  __int64 v1208; // rcx
  int *v1209; // r15
  int *v1210; // r12
  __int64 v1211; // rax
  int *v1212; // rax
  __int64 v1213; // rax
  int *v1214; // r13
  __int64 v1215; // rax
  __int64 v1216; // rax
  __int64 v1217; // rax
  __int64 v1218; // rax
  int *v1219; // rbx
  __int64 v1220; // rcx
  int *v1221; // rdi
  __int64 v1222; // rax
  __int64 v1223; // rax
  int *v1224; // r14
  __int64 v1225; // rax
  __int64 v1226; // r12
  __int64 v1227; // rax
  __int64 v1228; // rax
  _DWORD *v1229; // rbx
  __int64 v1230; // rax
  __int64 v1231; // rax
  int *v1232; // rbx
  __int64 v1233; // r14
  __int64 v1234; // rax
  __int64 v1235; // r8
  __int64 v1236; // rax
  int *v1237; // rbx
  __int64 v1238; // r14
  __int64 v1239; // rax
  __int64 v1240; // r8
  __int64 v1241; // r15
  __int64 v1242; // rax
  int *v1243; // r14
  __int64 v1244; // rax
  int v1245; // r14d
  __int64 v1246; // rax
  __int64 v1247; // rax
  _DWORD *v1248; // rbx
  __int64 v1249; // rax
  int *v1250; // rbx
  __int64 v1251; // rcx
  int *v1252; // r14
  __int64 v1253; // rax
  int *v1254; // r15
  __int64 v1255; // rax
  __int64 v1256; // rax
  __int64 v1257; // rbx
  __int64 v1258; // rax
  __int64 v1259; // r8
  int *v1260; // rbx
  __int64 v1261; // rax
  __int64 v1262; // rax
  __int64 v1263; // r15
  __int64 v1264; // rax
  int *v1265; // rbx
  __int64 v1266; // rax
  int *v1267; // r14
  __int64 v1268; // rax
  __int64 v1269; // r13
  int *v1270; // r15
  __int64 v1271; // rax
  __int64 v1272; // rax
  int *v1273; // r12
  __int64 v1274; // rax
  __int64 v1275; // rax
  __int64 v1276; // rax
  int *v1277; // r13
  __int64 v1278; // rax
  __int64 v1279; // rax
  __int64 v1280; // rax
  int *v1281; // r14
  __int64 v1282; // rax
  __int64 v1283; // rax
  __int64 v1284; // rax
  __int64 v1285; // rax
  int *v1286; // rdi
  __int64 v1287; // r12
  __int64 v1288; // r9
  int *v1289; // r14
  __int64 v1290; // rax
  __int64 v1291; // rax
  __int64 v1292; // rdi
  __int64 v1293; // rbx
  __int64 v1294; // rax
  int *v1295; // rbx
  int *v1296; // rdi
  __int64 v1297; // rax
  __int64 v1298; // rax
  int *v1299; // r12
  __int64 v1300; // rax
  int *v1301; // r13
  __int64 v1302; // rax
  __int64 v1303; // rax
  __int64 v1304; // rax
  __int64 v1305; // rdi
  __int64 v1306; // r15
  _DWORD *v1307; // rbx
  __int64 v1308; // rax
  int *v1309; // rdi
  __int64 v1310; // rax
  __int64 v1311; // rax
  int *v1312; // rbx
  __int64 v1313; // rax
  __int64 v1314; // rax
  __int64 v1315; // rax
  __int64 v1316; // rax
  int *v1317; // r12
  int v1318; // edx
  int v1319; // r8d
  __int64 v1320; // rax
  int *v1321; // r14
  __int64 v1322; // rax
  __int64 v1323; // r15
  __int64 v1324; // rax
  __int64 v1325; // rbx
  __int64 v1326; // rbx
  __int64 v1327; // rax
  __int64 v1328; // rax
  __int64 v1329; // rax
  int *v1330; // rbx
  int v1331; // r14d
  __int64 v1332; // rax
  __int64 v1333; // rax
  __int64 v1334; // rax
  __int64 v1335; // r12
  __int64 v1336; // r14
  int v1337; // ecx
  __int64 v1338; // rax
  int *v1339; // rdx
  int *v1340; // rax
  int *v1341; // rbx
  __int64 v1342; // rax
  int *v1343; // rbx
  unsigned int v1344; // r14d
  int v1345; // eax
  __int64 v1346; // rax
  __int64 v1347; // r13
  _DWORD *v1348; // r14
  __int64 v1349; // rax
  int *v1350; // rdi
  __int64 v1351; // rax
  __int64 v1352; // rax
  __int64 v1353; // rax
  int *v1354; // r15
  __int64 v1355; // rax
  __int64 v1356; // rax
  int *v1357; // rdi
  __int64 v1358; // rax
  __int64 v1359; // rax
  __int64 v1360; // rax
  __int64 v1361; // rax
  __int64 v1362; // rax
  __int64 v1363; // rax
  __int64 v1364; // rbx
  __int64 v1365; // rax
  __int64 v1366; // rax
  __int64 v1367; // rbx
  __int64 v1368; // rax
  __int64 v1369; // rax
  __int64 v1370; // rax
  __int64 v1371; // rax
  int *v1372; // rdi
  __int64 v1373; // rdx
  __int64 v1374; // rbx
  __int64 v1375; // rax
  __int64 v1376; // rbx
  __int64 v1377; // rax
  __int64 v1378; // rax
  int *v1379; // r14
  __int64 v1380; // rax
  int *v1381; // r14
  __int64 v1382; // rax
  int *v1383; // r14
  __int64 v1384; // rax
  int *v1385; // r14
  __int64 v1386; // rax
  __int64 v1387; // rax
  int *v1388; // rdi
  __int64 v1389; // rax
  int *v1390; // rdi
  __int64 v1391; // rax
  int *v1392; // rdi
  __int64 v1393; // rax
  int *v1394; // rdi
  __int64 v1395; // rax
  int *v1396; // rdi
  __int64 v1397; // rax
  int *v1398; // rdi
  __int64 v1399; // rax
  int *v1400; // rdi
  __int64 v1401; // rax
  int *v1402; // rdi
  __int64 v1403; // rax
  int *v1404; // rdi
  __int64 v1405; // rax
  int *v1406; // rdi
  __int64 v1407; // rax
  int *v1408; // rdi
  __int64 v1409; // rax
  int *v1410; // rdi
  __int64 v1411; // rax
  int *v1412; // rdi
  __int64 v1413; // rax
  _QWORD *v1414; // rdi
  __int64 v1415; // rax
  int *v1416; // rdi
  __int64 v1417; // rax
  int *v1418; // rdi
  __int64 v1419; // rax
  int *v1420; // rdi
  __int64 v1421; // rax
  int *v1422; // rdi
  __int64 v1423; // rax
  int *v1424; // rdi
  __int64 v1425; // rax
  __int64 v1426; // rax
  int *v1427; // rbx
  __int64 v1428; // rax
  int *v1429; // rbx
  __int64 v1430; // rax
  __int64 v1431; // rax
  __int64 v1432; // [rsp+20h] [rbp-160h]
  __int64 v1433; // [rsp+20h] [rbp-160h]
  __int64 v1434; // [rsp+20h] [rbp-160h]
  __int64 v1435; // [rsp+28h] [rbp-158h]
  __int64 v1436; // [rsp+28h] [rbp-158h]
  __int64 v1437; // [rsp+28h] [rbp-158h]
  __int64 v1438; // [rsp+100h] [rbp-80h] BYREF
  int *v1439; // [rsp+108h] [rbp-78h]
  __int64 v1440; // [rsp+110h] [rbp-70h]
  __int64 v1441; // [rsp+118h] [rbp-68h]
  _QWORD *v1442; // [rsp+120h] [rbp-60h]
  int *v1443; // [rsp+128h] [rbp-58h]
  int *v1444; // [rsp+130h] [rbp-50h]
  __int64 v1445; // [rsp+138h] [rbp-48h]
  _QWORD *v1446; // [rsp+140h] [rbp-40h]
  int *v1447; // [rsp+148h] [rbp-38h]
  int *v1448; // [rsp+150h] [rbp-30h]
  int *v1449; // [rsp+158h] [rbp-28h]
  _QWORD *v1450; // [rsp+160h] [rbp-20h]
  int *v1451; // [rsp+168h] [rbp-18h]
  int *v1452; // [rsp+170h] [rbp-10h]
  int *v1453; // [rsp+178h] [rbp-8h]
  _QWORD *v1454; // [rsp+180h] [rbp+0h]
  int *v1455; // [rsp+188h] [rbp+8h]
  int *v1456; // [rsp+190h] [rbp+10h]
  int *v1457; // [rsp+198h] [rbp+18h]
  int *v1458; // [rsp+1A0h] [rbp+20h]
  int *v1459; // [rsp+1A8h] [rbp+28h]
  int *v1460; // [rsp+1B0h] [rbp+30h]
  int *v1461; // [rsp+1B8h] [rbp+38h]
  int *v1462; // [rsp+1C0h] [rbp+40h]
  __int64 v1463; // [rsp+1C8h] [rbp+48h]
  _QWORD *v1464; // [rsp+1D0h] [rbp+50h]
  int *v1465; // [rsp+1D8h] [rbp+58h]
  int *v1466; // [rsp+1E0h] [rbp+60h]
  int *v1467; // [rsp+1E8h] [rbp+68h]
  int *v1468; // [rsp+1F0h] [rbp+70h]
  int *v1469; // [rsp+1F8h] [rbp+78h]
  int *v1470; // [rsp+200h] [rbp+80h]
  __int64 v1471; // [rsp+208h] [rbp+88h] BYREF
  __int64 v1472; // [rsp+210h] [rbp+90h] BYREF
  int *v1473; // [rsp+218h] [rbp+98h]
  int *v1474; // [rsp+220h] [rbp+A0h]
  int *v1475; // [rsp+228h] [rbp+A8h]
  int *v1476; // [rsp+230h] [rbp+B0h]
  int *v1477; // [rsp+238h] [rbp+B8h]
  int *v1478; // [rsp+240h] [rbp+C0h]
  int *v1479; // [rsp+248h] [rbp+C8h]
  int *v1480; // [rsp+250h] [rbp+D0h]
  int *v1481; // [rsp+258h] [rbp+D8h]
  int *v1482; // [rsp+260h] [rbp+E0h]
  int *v1483; // [rsp+268h] [rbp+E8h]
  int *v1484; // [rsp+270h] [rbp+F0h]
  int *v1485; // [rsp+278h] [rbp+F8h]
  int *v1486; // [rsp+280h] [rbp+100h]
  int *v1487; // [rsp+288h] [rbp+108h] BYREF
  int *v1488; // [rsp+290h] [rbp+110h] BYREF
  __int64 v1489; // [rsp+298h] [rbp+118h] BYREF
  __int64 v1490; // [rsp+2A0h] [rbp+120h] BYREF
  int *v1491; // [rsp+2A8h] [rbp+128h] BYREF
  int *v1492; // [rsp+2B0h] [rbp+130h] BYREF
  int *v1493; // [rsp+2B8h] [rbp+138h] BYREF
  __int64 v1494; // [rsp+2C0h] [rbp+140h] BYREF
  int *v1495; // [rsp+2C8h] [rbp+148h] BYREF
  int *v1496; // [rsp+2D0h] [rbp+150h] BYREF
  __int64 v1497; // [rsp+2D8h] [rbp+158h] BYREF
  int *v1498; // [rsp+2E0h] [rbp+160h] BYREF
  int *v1499; // [rsp+2E8h] [rbp+168h] BYREF
  int *v1500; // [rsp+2F0h] [rbp+170h] BYREF
  int *v1501; // [rsp+2F8h] [rbp+178h] BYREF
  int *v1502; // [rsp+300h] [rbp+180h] BYREF
  int *v1503; // [rsp+308h] [rbp+188h] BYREF
  _QWORD *v1504; // [rsp+310h] [rbp+190h] BYREF
  int *v1505; // [rsp+318h] [rbp+198h] BYREF
  int *v1506; // [rsp+320h] [rbp+1A0h] BYREF
  __int64 v1507; // [rsp+328h] [rbp+1A8h] BYREF
  _QWORD v1508[2]; // [rsp+330h] [rbp+1B0h] BYREF
  _QWORD v1509[2]; // [rsp+340h] [rbp+1C0h] BYREF
  _QWORD v1510[2]; // [rsp+350h] [rbp+1D0h] BYREF
  _QWORD v1511[2]; // [rsp+360h] [rbp+1E0h] BYREF
  _QWORD v1512[2]; // [rsp+370h] [rbp+1F0h] BYREF
  _QWORD v1513[2]; // [rsp+380h] [rbp+200h] BYREF
  _QWORD v1514[2]; // [rsp+390h] [rbp+210h] BYREF
  _QWORD v1515[2]; // [rsp+3A0h] [rbp+220h] BYREF
  _QWORD v1516[2]; // [rsp+3B0h] [rbp+230h] BYREF
  _QWORD v1517[2]; // [rsp+3C0h] [rbp+240h] BYREF
  _QWORD v1518[2]; // [rsp+3D0h] [rbp+250h] BYREF
  _QWORD v1519[2]; // [rsp+3E0h] [rbp+260h] BYREF
  _QWORD v1520[2]; // [rsp+3F0h] [rbp+270h] BYREF
  _QWORD v1521[2]; // [rsp+400h] [rbp+280h] BYREF
  _QWORD v1522[2]; // [rsp+410h] [rbp+290h] BYREF
  _QWORD v1523[2]; // [rsp+420h] [rbp+2A0h] BYREF
  _QWORD v1524[2]; // [rsp+430h] [rbp+2B0h] BYREF
  _QWORD v1525[2]; // [rsp+440h] [rbp+2C0h] BYREF
  _QWORD v1526[2]; // [rsp+450h] [rbp+2D0h] BYREF
  _QWORD v1527[2]; // [rsp+460h] [rbp+2E0h] BYREF
  _QWORD v1528[3]; // [rsp+470h] [rbp+2F0h] BYREF
  _QWORD v1529[3]; // [rsp+488h] [rbp+308h] BYREF
  _QWORD v1530[3]; // [rsp+4A0h] [rbp+320h] BYREF
  _QWORD v1531[3]; // [rsp+4B8h] [rbp+338h] BYREF
  _QWORD v1532[3]; // [rsp+4D0h] [rbp+350h] BYREF
  _QWORD v1533[3]; // [rsp+4E8h] [rbp+368h] BYREF
  _QWORD v1534[5]; // [rsp+500h] [rbp+380h] BYREF
  _QWORD v1535[5]; // [rsp+528h] [rbp+3A8h] BYREF
  _QWORD v1536[5]; // [rsp+550h] [rbp+3D0h] BYREF

  v1441 = sub_18036D7E0(*a3);
  v1482 = (int *)a3[1];
  v1456 = (int *)a3[2];
  v1445 = sub_18036D6D0();
  v5 = nullptr;
  v1443 = nullptr;
  v1452 = nullptr;
  v6 = (int *)v1445;
  v1473 = nullptr;
  v1474 = (int *)sub_18036D6D0();
  v1438 = 0;
  v1466 = nullptr;
  v1461 = nullptr;
  v1475 = nullptr;
  v1476 = nullptr;
  v1477 = nullptr;
  v1478 = nullptr;
  v1479 = nullptr;
  v1457 = nullptr;
  v1480 = nullptr;
  v1483 = nullptr;
  v1467 = nullptr;
  v1462 = nullptr;
  v1460 = nullptr;
  v1468 = nullptr;
  v1469 = nullptr;
  v1464 = nullptr;
  v1481 = nullptr;
  v1470 = nullptr;
  v1484 = nullptr;
  v1485 = nullptr;
  v1486 = nullptr;
  v1455 = nullptr;
  v1447 = nullptr;
  v1465 = nullptr;
  v1458 = nullptr;
  v1444 = nullptr;
  v1459 = nullptr;
  v1454 = nullptr;
  v1449 = nullptr;
  v1446 = nullptr;
  v1450 = nullptr;
  v1448 = nullptr;
  v1442 = nullptr;
  v1439 = nullptr;
  if ( (unsigned __int8)sub_1800DEB30(qword_1803BED28) != 0 )
  {
    v7 = qword_1803BED28;
    if ( qword_1803BED28 != 0 && *(int *)qword_1803BED28 >= 0 )
    {
      v8 = (*(_QWORD *)qword_1803BED28)-- == 1;
      if ( v8 )
      {
        v9 = sub_18000CC40(v7);
        (*(void (__fastcall **)(__int64))(v9 + 48))(v7);
      }
    }
    v10 = sub_18036CD10(a1, *(_QWORD *)&qword_1803BDE20, qword_1803CA5C0, 240);
    qword_1803BED28 = v10;
  }
  else
  {
    v10 = qword_1803BED28;
  }
  v1440 = v10;
  sub_180001D00(a1, v10);
  v11 = (int *)v1441;
// Original IDA line 1603 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
  v12 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
  v13 = (int *)v12;
  if ( v12 == 0 )
  {
    sub_18000CE30(a1, &v1438);
    v14 = 244;
    goto LABEL_2631;
  }
  v15 = v1482;
  v14 = 244;
  *(_DWORD *)(v10 + 40) = 244;
// Original IDA line 1614 | qword_1803BDF80 -> novels.services.agent_session#14 bytes [4307115, 4307129]: "load_session"
  v16 = sub_18037E6C0(a1, v12, qword_1803BDF80, v15);
  if ( *v13 >= 0 )
  {
    v8 = (*(_QWORD *)v13)-- == 1;
    if ( v8 )
    {
      v17 = sub_18000CC40(v13);
      (*(void (__fastcall **)(int *))(v17 + 48))(v13);
    }
  }
  if ( v16 == 0 )
    goto LABEL_101;
  v18 = v16;
  v6 = (int *)v1445;
  sub_1800DEB50(v1445, v18);
// Original IDA line 1629 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
  v19 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258);
  v1451 = (int *)v19;
  v5 = (int *)v19;
  if ( v19 == 0 )
  {
    sub_18000CE30(a1, &v1438);
    v5 = v1443;
    v14 = 245;
    goto LABEL_2631;
  }
  v1443 = (int *)v19;
  v20 = (int *)v19;
  v21 = sub_180001510(v19);
  if ( v21 == -1 )
  {
    sub_18000CE30(a1, &v1438);
    v14 = 246;
    goto LABEL_2631;
  }
  if ( v21 == 0 )
  {
// Original IDA line 1650 | qword_1803BE2B8 -> novels.services.agent_session#117 bytes [4308680, 4308727]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4308682, 4308691], "value": "success"}, {"tag": "a", "fileRange": [4308691, 4308700], "value": "message"}], "values": [{"tag": "F", "fileRange": [4308700, 4308701], "value": false}, {"tag... (truncated literal)
    v23 = (int *)sub_18036FB80(a1, qword_1803BE2B8);
    goto LABEL_151;
  }
// Original IDA line 1653 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
  v24 = sub_18037F0C0(a1, v5, qword_1803BDFD8);
  v25 = (int *)v24;
  if ( v24 == 0 )
  {
    sub_18000CE30(a1, &v1438);
    v14 = 248;
    goto LABEL_2631;
  }
  v14 = 248;
  *(_DWORD *)(v10 + 40) = 248;
// Original IDA line 1663 | qword_1803BE2C0 -> novels.services.agent_session#118 bytes [4308727, 4308738]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4308729, 4308737], "value": "action"}, {"tag": "s", "fileRange": [4308737, 4308738], "value": ""}]}
  v26 = (int *)sub_180378210(a1, v24, qword_1803BE2C0);
  if ( *v25 >= 0 )
  {
    v8 = (*(_QWORD *)v25)-- == 1;
    if ( v8 )
    {
      v27 = sub_18000CC40(v25);
      (*(void (__fastcall **)(int *))(v27 + 48))(v25);
    }
  }
  if ( v26 == nullptr )
    goto LABEL_101;
  v28 = sub_1803749A0(v26);
  v29 = *v26 < 0;
  v1463 = v28;
  if ( !v29 )
  {
    v8 = (*(_QWORD *)v26)-- == 1;
    if ( v8 )
    {
      v30 = sub_18000CC40(v26);
      (*(void (__fastcall **)(int *))(v30 + 48))(v26);
      v28 = v1463;
    }
  }
  if ( v28 == 0 )
    goto LABEL_101;
  v1452 = (int *)v28;
  v31 = sub_1801F4280(a1);
  if ( v31 == 0 )
  {
// Original IDA line 1694 | qword_1803BE1B0 -> novels.services.agent_session#84 bytes [4308170, 4308196]: "validate_proposed_action"
    sub_180373340(a1, &v1438, qword_1803BE1B0);
    v11 = (int *)v1441;
    v14 = 250;
LABEL_2630:
    v6 = (int *)v1445;
LABEL_2631:
    v1370 = sub_18000CE80(&v1438);
    if ( v1370 != 0 )
    {
      if ( *(_QWORD *)(v1370 + 24) == v10 )
      {
LABEL_2636:
        v1372 = v1461;
        sub_18036D0A0(v10, "coocooocoooooooooooooooooooooo", v11);
        v1374 = qword_1803BED28;
        if ( v10 == qword_1803BED28 )
        {
          if ( *(int *)qword_1803BED28 >= 0 )
          {
            v8 = (*(_QWORD *)qword_1803BED28)-- == 1;
            if ( v8 )
            {
              v1375 = sub_18000CC40(v1374);
              (*(void (__fastcall **)(__int64))(v1375 + 48))(v1374);
            }
          }
          qword_1803BED28 = 0;
        }
        sub_180001D80(a1, v1373);
        v1376 = v1438;
        sub_18000CE70(&v1438);
        if ( *v6 >= 0 )
        {
          v8 = (*(_QWORD *)v6)-- == 1;
          if ( v8 )
          {
            v1377 = sub_18000CC40(v6);
            (*(void (__fastcall **)(int *))(v1377 + 48))(v6);
          }
        }
        if ( v5 != nullptr && *v5 >= 0 )
        {
          v8 = (*(_QWORD *)v5)-- == 1;
          if ( v8 )
          {
            v1378 = sub_18000CC40(v5);
            (*(void (__fastcall **)(int *))(v1378 + 48))(v5);
          }
        }
        v1379 = v1452;
        if ( v1452 != nullptr && *v1452 >= 0 )
        {
          v8 = (*(_QWORD *)v1452)-- == 1;
          if ( v8 )
          {
            v1380 = sub_18000CC40(v1379);
            (*(void (__fastcall **)(int *))(v1380 + 48))(v1379);
          }
        }
        v1381 = v1473;
        if ( v1473 != nullptr && *v1473 >= 0 )
        {
          v8 = (*(_QWORD *)v1473)-- == 1;
          if ( v8 )
          {
            v1382 = sub_18000CC40(v1381);
            (*(void (__fastcall **)(int *))(v1382 + 48))(v1381);
          }
        }
        v1383 = v1474;
        if ( *v1474 >= 0 )
        {
          v8 = (*(_QWORD *)v1474)-- == 1;
          if ( v8 )
          {
            v1384 = sub_18000CC40(v1383);
            (*(void (__fastcall **)(int *))(v1384 + 48))(v1383);
          }
        }
        v1385 = v1466;
        if ( v1466 != nullptr && *v1466 >= 0 )
        {
          v8 = (*(_QWORD *)v1466)-- == 1;
          if ( v8 )
          {
            v1386 = sub_18000CC40(v1385);
            (*(void (__fastcall **)(int *))(v1386 + 48))(v1385);
          }
        }
        if ( v1372 != nullptr && *v1372 >= 0 )
        {
          v8 = (*(_QWORD *)v1372)-- == 1;
          if ( v8 )
          {
            v1387 = sub_18000CC40(v1372);
            (*(void (__fastcall **)(int *))(v1387 + 48))(v1372);
          }
        }
        v1388 = v1475;
        if ( v1475 != nullptr && *v1475 >= 0 )
        {
          v8 = (*(_QWORD *)v1475)-- == 1;
          if ( v8 )
          {
            v1389 = sub_18000CC40(v1388);
            (*(void (__fastcall **)(int *))(v1389 + 48))(v1388);
          }
        }
        v1390 = v1476;
        if ( v1476 != nullptr && *v1476 >= 0 )
        {
          v8 = (*(_QWORD *)v1476)-- == 1;
          if ( v8 )
          {
            v1391 = sub_18000CC40(v1390);
            (*(void (__fastcall **)(int *))(v1391 + 48))(v1390);
          }
        }
        v1392 = v1477;
        if ( v1477 != nullptr && *v1477 >= 0 )
        {
          v8 = (*(_QWORD *)v1477)-- == 1;
          if ( v8 )
          {
            v1393 = sub_18000CC40(v1392);
            (*(void (__fastcall **)(int *))(v1393 + 48))(v1392);
          }
        }
        v1394 = v1478;
        if ( v1478 != nullptr && *v1478 >= 0 )
        {
          v8 = (*(_QWORD *)v1478)-- == 1;
          if ( v8 )
          {
            v1395 = sub_18000CC40(v1394);
            (*(void (__fastcall **)(int *))(v1395 + 48))(v1394);
          }
        }
        v1396 = v1479;
        if ( v1479 != nullptr && *v1479 >= 0 )
        {
          v8 = (*(_QWORD *)v1479)-- == 1;
          if ( v8 )
          {
            v1397 = sub_18000CC40(v1396);
            (*(void (__fastcall **)(int *))(v1397 + 48))(v1396);
          }
        }
        v1398 = v1457;
        if ( v1457 != nullptr && *v1457 >= 0 )
        {
          v8 = (*(_QWORD *)v1457)-- == 1;
          if ( v8 )
          {
            v1399 = sub_18000CC40(v1398);
            (*(void (__fastcall **)(int *))(v1399 + 48))(v1398);
          }
        }
        v1400 = v1480;
        if ( v1480 != nullptr && *v1480 >= 0 )
        {
          v8 = (*(_QWORD *)v1480)-- == 1;
          if ( v8 )
          {
            v1401 = sub_18000CC40(v1400);
            (*(void (__fastcall **)(int *))(v1401 + 48))(v1400);
          }
        }
        v1402 = v1483;
        if ( v1483 != nullptr && *v1483 >= 0 )
        {
          v8 = (*(_QWORD *)v1483)-- == 1;
          if ( v8 )
          {
            v1403 = sub_18000CC40(v1402);
            (*(void (__fastcall **)(int *))(v1403 + 48))(v1402);
          }
        }
        v1404 = v1467;
        if ( v1467 != nullptr && *v1467 >= 0 )
        {
          v8 = (*(_QWORD *)v1467)-- == 1;
          if ( v8 )
          {
            v1405 = sub_18000CC40(v1404);
            (*(void (__fastcall **)(int *))(v1405 + 48))(v1404);
          }
        }
        v1406 = v1462;
        if ( v1462 != nullptr && *v1462 >= 0 )
        {
          v8 = (*(_QWORD *)v1462)-- == 1;
          if ( v8 )
          {
            v1407 = sub_18000CC40(v1406);
            (*(void (__fastcall **)(int *))(v1407 + 48))(v1406);
          }
        }
        v1408 = v1460;
        if ( v1460 != nullptr && *v1460 >= 0 )
        {
          v8 = (*(_QWORD *)v1460)-- == 1;
          if ( v8 )
          {
            v1409 = sub_18000CC40(v1408);
            (*(void (__fastcall **)(int *))(v1409 + 48))(v1408);
          }
        }
        v1410 = v1468;
        if ( v1468 != nullptr && *v1468 >= 0 )
        {
          v8 = (*(_QWORD *)v1468)-- == 1;
          if ( v8 )
          {
            v1411 = sub_18000CC40(v1410);
            (*(void (__fastcall **)(int *))(v1411 + 48))(v1410);
          }
        }
        v1412 = v1469;
        if ( v1469 != nullptr && *v1469 >= 0 )
        {
          v8 = (*(_QWORD *)v1469)-- == 1;
          if ( v8 )
          {
            v1413 = sub_18000CC40(v1412);
            (*(void (__fastcall **)(int *))(v1413 + 48))(v1412);
          }
        }
        v1414 = v1464;
        if ( v1464 != nullptr && *(int *)v1464 >= 0 )
        {
          v8 = (*v1464)-- == 1;
          if ( v8 )
          {
            v1415 = sub_18000CC40(v1414);
            (*(void (__fastcall **)(_QWORD *))(v1415 + 48))(v1414);
          }
        }
        v1416 = v1481;
        if ( v1481 != nullptr && *v1481 >= 0 )
        {
          v8 = (*(_QWORD *)v1481)-- == 1;
          if ( v8 )
          {
            v1417 = sub_18000CC40(v1416);
            (*(void (__fastcall **)(int *))(v1417 + 48))(v1416);
          }
        }
        v1418 = v1470;
        if ( v1470 != nullptr && *v1470 >= 0 )
        {
          v8 = (*(_QWORD *)v1470)-- == 1;
          if ( v8 )
          {
            v1419 = sub_18000CC40(v1418);
            (*(void (__fastcall **)(int *))(v1419 + 48))(v1418);
          }
        }
        v1420 = v1484;
        if ( v1484 != nullptr && *v1484 >= 0 )
        {
          v8 = (*(_QWORD *)v1484)-- == 1;
          if ( v8 )
          {
            v1421 = sub_18000CC40(v1420);
            (*(void (__fastcall **)(int *))(v1421 + 48))(v1420);
          }
        }
        v1422 = v1485;
        if ( v1485 != nullptr && *v1485 >= 0 )
        {
          v8 = (*(_QWORD *)v1485)-- == 1;
          if ( v8 )
          {
            v1423 = sub_18000CC40(v1422);
            (*(void (__fastcall **)(int *))(v1423 + 48))(v1422);
          }
        }
        v1424 = v1486;
        if ( v1486 != nullptr && *v1486 >= 0 )
        {
          v8 = (*(_QWORD *)v1486)-- == 1;
          if ( v8 )
          {
            v1425 = sub_18000CC40(v1424);
            (*(void (__fastcall **)(int *))(v1425 + 48))(v1424);
          }
        }
        v1438 = v1376;
        if ( *v11 >= 0 )
        {
          v8 = (*(_QWORD *)v11)-- == 1;
          if ( v8 )
          {
            v1426 = sub_18000CC40(v11);
            (*(void (__fastcall **)(int *))(v1426 + 48))(v11);
          }
        }
        v1427 = v1482;
        if ( *v1482 >= 0 )
        {
          v8 = (*(_QWORD *)v1482)-- == 1;
          if ( v8 )
          {
            v1428 = sub_18000CC40(v1427);
            (*(void (__fastcall **)(int *))(v1428 + 48))(v1427);
          }
        }
        v1429 = v1456;
        if ( *v1456 >= 0 )
        {
          v8 = (*(_QWORD *)v1456)-- == 1;
          if ( v8 )
          {
            v1430 = sub_18000CC40(v1429);
            (*(void (__fastcall **)(int *))(v1430 + 48))(v1429);
          }
        }
        sub_18000CE40(a1, &v1438);
        return nullptr;
      }
      v1371 = sub_18000CC90(v1370, v10, v14);
    }
    else
    {
      v1371 = sub_18038E2A0(v10, v14);
    }
    sub_180001300(&v1438, v1371);
    goto LABEL_2636;
  }
  v6 = (int *)v1445;
// Original IDA line 2025 | qword_1803BDFA0 -> novels.services.agent_session#18 bytes [4307199, 4307205]: "kind"
  v32 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFA0);
  if ( v32 == nullptr )
  {
    sub_18000CE30(a1, &v1438);
LABEL_35:
    v11 = (int *)v1441;
    v14 = 250;
    goto LABEL_2631;
  }
// Original IDA line 2034 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
  v33 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
  v34 = (int *)v33;
  if ( v33 == 0 )
  {
    sub_18000CE30(a1, &v1438);
    if ( *v32 >= 0 )
    {
      v8 = (*(_QWORD *)v32)-- == 1;
      if ( v8 )
      {
        v35 = sub_18000CC40(v32);
        (*(void (__fastcall **)(int *))(v35 + 48))(v32);
      }
    }
    goto LABEL_35;
  }
  v36 = sub_180001510(v33);
  if ( *v34 >= 0 )
  {
    v8 = (*(_QWORD *)v34)-- == 1;
    if ( v8 )
    {
      v37 = sub_18000CC40(v34);
      (*(void (__fastcall **)(int *))(v37 + 48))(v34);
    }
  }
  if ( v36 == -1 )
  {
    sub_18000CE30(a1, &v1438);
    if ( *v32 >= 0 )
    {
      v8 = (*(_QWORD *)v32)-- == 1;
      if ( v8 )
      {
        v38 = sub_18000CC40(v32);
        (*(void (__fastcall **)(int *))(v38 + 48))(v32);
      }
    }
    v14 = 250;
    goto LABEL_2629;
  }
  v39 = Py_FalseStruct;
  v14 = 250;
  if ( v36 != 0 )
    v39 = Py_TrueStruct;
  *(_DWORD *)(v10 + 40) = 250;
  v1490 = v39;
  v1508[0] = v32;
  v1508[1] = v5;
// Original IDA line 2083 | qword_1803BE140 -> novels.services.agent_session#70 bytes [4307950, 4307962]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4307952, 4307962], "value": "has_book"}]}
  v40 = sub_18037D650(a1, v31, v1508, &v1490, qword_1803BE140);
  if ( *v32 >= 0 )
  {
    v8 = (*(_QWORD *)v32)-- == 1;
    if ( v8 )
    {
      v41 = sub_18000CC40(v32);
      (*(void (__fastcall **)(int *))(v41 + 48))(v32);
    }
  }
  if ( v40 == 0 )
    goto LABEL_54;
  v1473 = (int *)v40;
  v42 = sub_180001510(v40);
  if ( v42 == -1 )
  {
    sub_18000CE30(a1, &v1438);
    v14 = 251;
    goto LABEL_2629;
  }
  if ( v42 == 0 )
  {
    v44 = sub_1801F3630(a1);
    if ( v44 == 0 )
    {
// Original IDA line 2108 | qword_1803BE1D8 -> novels.services.agent_session#89 bytes [4308243, 4308261]: "ASSISTED_ACTIONS"
      sub_180373340(a1, &v1438, qword_1803BE1D8);
      v14 = 254;
      goto LABEL_2629;
    }
    v45 = v1463;
    v46 = PySequence_Contains(v44, v1463);
    if ( v46 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v14 = 254;
      goto LABEL_2629;
    }
    if ( v46 == 1 )
    {
      v6 = (int *)v1445;
// Original IDA line 2123 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
      if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
      {
        sub_18000CE30(a1, &v1438);
        v11 = (int *)v1441;
        v14 = 255;
        goto LABEL_2631;
      }
// Original IDA line 2130 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
      v47 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
      v48 = (int *)v47;
      v49 = a1;
      if ( v47 != 0 )
      {
// Original IDA line 2135 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
        v50 = (int *)sub_18037F0C0(a1, v47, qword_1803BE0B8);
        if ( *v48 >= 0 )
        {
          v8 = (*(_QWORD *)v48)-- == 1;
          if ( v8 )
          {
            v51 = sub_18000CC40(v48);
            (*(void (__fastcall **)(int *))(v51 + 48))(v48);
          }
        }
        v49 = a1;
        if ( v50 != nullptr )
        {
// Original IDA line 2148 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
          v52 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
          if ( v52 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v50 >= 0 )
            {
              v8 = (*(_QWORD *)v50)-- == 1;
              if ( v8 )
              {
                v53 = sub_18000CC40(v50);
                (*(void (__fastcall **)(int *))(v53 + 48))(v50);
              }
            }
            v14 = 256;
            goto LABEL_2629;
          }
// Original IDA line 2164 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"
          v54 = qword_1803BE2E0;
          v14 = 256;
          *(_DWORD *)(v10 + 40) = 256;
          v1509[1] = v54;
          v1509[0] = v45;
// Original IDA line 2169 | qword_1803BE2E8 -> novels.services.agent_session#123 bytes [4308782, 4308823]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4308784, 4308800], "value": "propose_action"}, {"tag": "u", "fileRange": [4308800, 4308823], "value": "已打开导入入口"}]}; qword_1803BE2F0 -> novels.services.agent_session#124 bytes [4308823, 4308841]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4308825, 4308833], "value": "action"}, {"tag": "a", "fileRange": [4308833, 4308841], "value": "status"}]}
          v55 = (int *)sub_18037D8C0(a1, v52, qword_1803BE2E8, v1509, qword_1803BE2F0);
          if ( *v52 >= 0 )
          {
            v8 = (*(_QWORD *)v52)-- == 1;
            if ( v8 )
            {
              v56 = sub_18000CC40(v52);
              (*(void (__fastcall **)(int *))(v56 + 48))(v52);
            }
          }
          if ( v55 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v50 >= 0 )
            {
              v8 = (*(_QWORD *)v50)-- == 1;
              if ( v8 )
              {
                v57 = sub_18000CC40(v50);
                (*(void (__fastcall **)(int *))(v57 + 48))(v50);
              }
            }
            goto LABEL_2629;
          }
          *(_DWORD *)(v10 + 40) = 256;
          v58 = (int *)sub_180376770(a1, v50, v55);
          if ( *v50 >= 0 )
          {
            v8 = (*(_QWORD *)v50)-- == 1;
            if ( v8 )
            {
              v59 = sub_18000CC40(v50);
              (*(void (__fastcall **)(int *))(v59 + 48))(v50);
            }
          }
          if ( *v55 >= 0 )
          {
            v8 = (*(_QWORD *)v55)-- == 1;
            if ( v8 )
            {
              v60 = sub_18000CC40(v55);
              (*(void (__fastcall **)(int *))(v60 + 48))(v55);
            }
          }
          if ( v58 == nullptr )
          {
LABEL_54:
            sub_18000CE30(a1, &v1438);
LABEL_2629:
            v11 = (int *)v1441;
            goto LABEL_2630;
          }
          if ( *v58 >= 0 )
          {
            v8 = (*(_QWORD *)v58)-- == 1;
            if ( v8 )
            {
              v61 = sub_18000CC40(v58);
              (*(void (__fastcall **)(int *))(v61 + 48))(v58);
            }
          }
          v11 = (int *)v1441;
// Original IDA line 2231 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
          v62 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
          v63 = (int *)v62;
          if ( v62 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v14 = 262;
            goto LABEL_2630;
          }
          v14 = 262;
          v64 = *(_QWORD *)(v1445 + 16);
          *(_DWORD *)(v10 + 40) = 262;
// Original IDA line 2242 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
          v65 = (int *)sub_18037E6C0(a1, v62, qword_1803BE0D8, v64);
          if ( *v63 >= 0 )
          {
            v8 = (*(_QWORD *)v63)-- == 1;
            if ( v8 )
            {
              v66 = sub_18000CC40(v63);
              (*(void (__fastcall **)(int *))(v66 + 48))(v63);
            }
          }
          if ( v65 == nullptr )
          {
LABEL_101:
            sub_18000CE30(a1, &v1438);
            goto LABEL_2630;
          }
          if ( *v65 >= 0 )
          {
            v8 = (*(_QWORD *)v65)-- == 1;
            if ( v8 )
            {
              v67 = sub_18000CC40(v65);
              (*(void (__fastcall **)(int *))(v67 + 48))(v65);
            }
          }
// Original IDA line 2267 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
          v68 = qword_1803BE2C8;
          v23 = (int *)PyDict_NewPresized(2);
          PyDict_SetItem(v23, v68, Py_TrueStruct);
// Original IDA line 2270 | qword_1803BE300 -> novels.services.agent_session#126 bytes [4308847, 4308863]: "assistedAction"
          v69 = qword_1803BE300;
// Original IDA line 2271 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
          v1442 = (_QWORD *)qword_1803BE2F8;
          v70 = (int *)PyDict_NewPresized(4);
          PyDict_SetItem(v70, v69, v1463);
// Original IDA line 2274 | qword_1803BE1F0 -> novels.services.agent_session#92 bytes [4308288, 4308301]: "targetRoute"
          v71 = qword_1803BE1F0;
// Original IDA line 2275 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v72 = (int *)sub_18037F0C0(a1, v5, qword_1803BDFD8);
          if ( v72 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            v10 = v1440;
            v14 = 267;
            goto LABEL_136;
          }
          v14 = 267;
          *(_DWORD *)(v1440 + 40) = 267;
// Original IDA line 2285 | qword_1803BE308 -> novels.services.agent_session#127 bytes [4308863, 4308878]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4308865, 4308878], "value": "targetRoute"}]}
          v73 = (int *)sub_1803770D0(a1, v72, qword_1803BE308);
          if ( *v72 >= 0 )
          {
            v8 = (*(_QWORD *)v72)-- == 1;
            if ( v8 )
            {
              v74 = sub_18000CC40(v72);
              (*(void (__fastcall **)(int *))(v74 + 48))(v72);
            }
          }
          if ( v73 == nullptr )
            goto LABEL_111;
          v75 = sub_180001510(v73);
          if ( v75 == -1 )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v73 < 0 )
              goto LABEL_112;
            v8 = (*(_QWORD *)v73)-- == 1;
            if ( !v8 )
              goto LABEL_112;
            v76 = sub_18000CC40(v73);
            (*(void (__fastcall **)(int *))(v76 + 48))(v73);
            v10 = v1440;
LABEL_136:
            if ( *v70 >= 0 )
            {
              v8 = (*(_QWORD *)v70)-- == 1;
              if ( v8 )
              {
                v90 = sub_18000CC40(v70);
                (*(void (__fastcall **)(int *))(v90 + 48))(v70);
              }
            }
            if ( *v23 >= 0 )
            {
              v8 = (*(_QWORD *)v23)-- == 1;
              if ( v8 )
              {
                v91 = sub_18000CC40(v23);
                (*(void (__fastcall **)(int *))(v91 + 48))(v23);
              }
            }
            goto LABEL_2628;
          }
          if ( v75 != 1 )
          {
            if ( *v73 >= 0 )
            {
              v8 = (*(_QWORD *)v73)-- == 1;
              if ( v8 )
              {
                v77 = sub_18000CC40(v73);
                (*(void (__fastcall **)(int *))(v77 + 48))(v73);
              }
            }
            v78 = v1463;
            v79 = *(_QWORD *)(v1441 + 16);
            *(_DWORD *)(v1440 + 40) = 267;
// Original IDA line 2344 | qword_1803BE1E0 -> novels.services.agent_session#90 bytes [4308261, 4308280]: "_route_for_action"
            v73 = (int *)sub_18037E6C0(a1, v79, *(_QWORD *)&qword_1803BE1E0, v78);
            if ( v73 == nullptr )
              goto LABEL_111;
          }
          PyDict_SetItem(v70, v71, v73);
          if ( *v73 >= 0 )
          {
            v8 = (*(_QWORD *)v73)-- == 1;
            if ( v8 )
            {
              v80 = sub_18000CC40(v73);
              (*(void (__fastcall **)(int *))(v80 + 48))(v73);
            }
          }
// Original IDA line 2358 | qword_1803BE310 -> novels.services.agent_session#128 bytes [4308878, 4308891]: "instruction"
          v81 = qword_1803BE310;
// Original IDA line 2359 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v82 = (int *)sub_18037F0C0(a1, v1451, qword_1803BDFD8);
          if ( v82 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            v10 = v1440;
            v14 = 268;
            goto LABEL_136;
          }
          v14 = 268;
          *(_DWORD *)(v1440 + 40) = 268;
// Original IDA line 2369 | qword_1803BE318 -> novels.services.agent_session#129 bytes [4308891, 4308907]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4308893, 4308906], "value": "instruction"}, {"tag": "s", "fileRange": [4308906, 4308907], "value": ""}]}
          v83 = (int *)sub_180378210(a1, v82, qword_1803BE318);
          if ( *v82 >= 0 )
          {
            v8 = (*(_QWORD *)v82)-- == 1;
            if ( v8 )
            {
              v84 = sub_18000CC40(v82);
              (*(void (__fastcall **)(int *))(v84 + 48))(v82);
            }
          }
          if ( v83 == nullptr )
          {
LABEL_111:
            sub_18000CE30(a1, &v1438);
LABEL_112:
            v10 = v1440;
            goto LABEL_136;
          }
          PyDict_SetItem(v70, v81, v83);
          if ( *v83 >= 0 )
          {
            v8 = (*(_QWORD *)v83)-- == 1;
            if ( v8 )
            {
              v85 = sub_18000CC40(v83);
              (*(void (__fastcall **)(int *))(v85 + 48))(v83);
            }
          }
          v14 = 269;
// Original IDA line 2398 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
          v86 = qword_1803BE1B8;
          v10 = v1440;
          v87 = *(_QWORD *)(v1445 + 16);
          *(_DWORD *)(v1440 + 40) = 269;
// Original IDA line 2402 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
          v88 = sub_18037E560(a1, v87, qword_1803BE278);
          v89 = (int *)v88;
          if ( v88 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            goto LABEL_136;
          }
          PyDict_SetItem(v70, v86, v88);
          if ( *v89 >= 0 )
          {
            v8 = (*(_QWORD *)v89)-- == 1;
            if ( v8 )
            {
              v92 = sub_18000CC40(v89);
              (*(void (__fastcall **)(int *))(v92 + 48))(v89);
            }
          }
          v93 = v1442;
          goto LABEL_146;
        }
      }
      sub_18000CE30(v49, &v1438);
      v11 = (int *)v1441;
      v14 = 256;
      goto LABEL_2631;
    }
    v146 = sub_180370DC0(a1, v5);
    if ( v146 == 0 )
    {
      sub_18000CE30(a1, &v1438);
      v14 = 273;
      goto LABEL_2629;
    }
    sub_1800DEB50(v1474, v146);
// Original IDA line 2436 | qword_1803BE320 -> novels.services.agent_session#130 bytes [4308907, 4308955]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4308909, 4308920], "value": "tool_name"}, {"tag": "a", "fileRange": [4308920, 4308925], "value": "exc"}, {"tag": "a", "fileRange": [4308925, 4308933], "value": "return"}], "values": [{"ta... (truncated literal)
    v147 = sub_18036FB80(a1, qword_1803BE320);
    v1528[0] = v1474;
    if ( *v1474 != -1 )
      ++*v1474;
    v148 = v1441;
    v1528[1] = v1441;
    if ( *(_DWORD *)v1441 != -1 )
      ++*(_DWORD *)v1441;
    v149 = v1445;
    v1528[2] = v1445;
    if ( *(_DWORD *)v1445 != -1 )
      ++*(_DWORD *)v1445;
    v1466 = (int *)sub_180217290(a1, v147, v1528);
// Original IDA line 2449 | qword_1803BE338 -> novels.services.agent_session#133 bytes [4309039, 4309052]: "create_book"
    v150 = sub_1803836D0(v45, qword_1803BE338);
    if ( v150 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 289;
      goto LABEL_2516;
    }
    if ( v150 == 1 )
    {
// Original IDA line 2458 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
      v152 = sub_18037F0C0(a1, *(_QWORD *)(v148 + 16), qword_1803BDF18);
      v153 = (int *)v152;
      if ( v152 != 0 )
      {
// Original IDA line 2462 | qword_1803BE338 -> novels.services.agent_session#133 bytes [4309039, 4309052]: "create_book"
        v154 = (int *)sub_18037F0C0(a1, v152, qword_1803BE338);
        if ( *v153 >= 0 )
        {
          v8 = (*(_QWORD *)v153)-- == 1;
          if ( v8 )
          {
            v155 = sub_18000CC40(v153);
            (*(void (__fastcall **)(int *))(v155 + 48))(v153);
          }
        }
        if ( v154 != nullptr )
        {
          v156 = v1456;
          v151 = 290;
          *(_DWORD *)(v10 + 40) = 290;
          v1491 = v156;
          v1492 = v5;
// Original IDA line 2479 | qword_1803BE340 -> novels.services.agent_session#134 bytes [4309052, 4309061]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309054, 4309061], "value": "model"}]}
          v157 = sub_18037D030(a1, v154, &v1492, &v1491, qword_1803BE340);
          v29 = *v154 < 0;
          v158 = (int *)v157;
          v1450 = (_QWORD *)v157;
          if ( !v29 )
          {
            v8 = (*(_QWORD *)v154)-- == 1;
            if ( v8 )
            {
              v159 = sub_18000CC40(v154);
              (*(void (__fastcall **)(int *))(v159 + 48))(v154);
            }
          }
          if ( v158 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
LABEL_272:
            v10 = v1440;
            goto LABEL_2516;
          }
          v160 = *(_QWORD *)(v149 + 16);
          v1461 = v158;
// Original IDA line 2501 | qword_1803BDFA0 -> novels.services.agent_session#18 bytes [4307199, 4307205]: "kind"; qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
          if ( (unsigned __int8)sub_18037F3A0(a1, v160, qword_1803BDFA0, qword_1803BE348) == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v10 = v1440;
            v151 = 291;
            goto LABEL_2516;
          }
// Original IDA line 2508 | qword_1803BE000 -> novels.services.agent_session#30 bytes [4307327, 4307331]: "id"
          v161 = sub_18037F0C0(a1, v158, qword_1803BE000);
          v162 = (int *)v161;
          v163 = a1;
          if ( v161 != 0 )
          {
// Original IDA line 2513 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
            v164 = sub_18037F3A0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0, v161);
            if ( *v162 >= 0 )
            {
              v8 = (*(_QWORD *)v162)-- == 1;
              if ( v8 )
              {
                v165 = sub_18000CC40(v162);
                (*(void (__fastcall **)(int *))(v165 + 48))(v162);
              }
            }
            v163 = a1;
            if ( v164 != 0 )
            {
// Original IDA line 2526 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
              v166 = sub_18037F0C0(a1, v158, qword_1803BE350);
              v167 = (int *)v166;
              if ( v166 != 0 )
              {
// Original IDA line 2530 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                v168 = sub_18037F3A0(a1, *(_QWORD *)(v149 + 16), qword_1803BE350, v166);
                if ( *v167 >= 0 )
                {
                  v8 = (*(_QWORD *)v167)-- == 1;
                  if ( v8 )
                  {
                    v169 = sub_18000CC40(v167);
                    (*(void (__fastcall **)(int *))(v169 + 48))(v167);
                  }
                }
                if ( v168 != 0 )
                {
// Original IDA line 2542 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                  if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v149 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
                  {
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 294;
                    goto LABEL_2516;
                  }
// Original IDA line 2549 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                  v170 = sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BE070);
                  v171 = (int *)v170;
                  if ( v170 != 0 )
                  {
// Original IDA line 2553 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                    v23 = (int *)sub_18037F0C0(a1, v170, qword_1803BE0B8);
                    if ( *v171 >= 0 )
                    {
                      v8 = (*(_QWORD *)v171)-- == 1;
                      if ( v8 )
                      {
                        v172 = sub_18000CC40(v171);
                        (*(void (__fastcall **)(int *))(v172 + 48))(v171);
                      }
                    }
                    if ( v23 != nullptr )
                    {
// Original IDA line 2565 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                      v173 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                      if ( v173 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v174 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v174 + 48))(v23);
                          }
                        }
                        v10 = v1440;
                        v151 = 295;
                        goto LABEL_2516;
                      }
// Original IDA line 2582 | qword_1803BE360 -> novels.services.agent_session#138 bytes [4309085, 4309105]: "已创建作品《"
                      v175 = (_DWORD *)qword_1803BE360;
// Original IDA line 2583 | qword_1803BE358 -> novels.services.agent_session#137 bytes [4309074, 4309085]: "architect"
                      v1442 = (_QWORD *)qword_1803BE358;
                      v176 = sub_180374A40(a1, 3);
                      v177 = (int *)v176;
                      if ( *v175 != -1 )
                        ++*v175;
                      sub_18000CC80(v176, 0, v175);
// Original IDA line 2589 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                      v178 = sub_18037F0C0(a1, v158, qword_1803BE350);
                      v179 = (int *)v178;
                      if ( v178 != 0 )
                      {
// Original IDA line 2593 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                        v180 = sub_180375520(a1, v178, qword_1803BDFE8);
                        if ( *v179 >= 0 )
                        {
                          v8 = (*(_QWORD *)v179)-- == 1;
                          if ( v8 )
                          {
                            v181 = sub_18000CC40(v179);
                            (*(void (__fastcall **)(int *))(v181 + 48))(v179);
                          }
                        }
                        if ( v180 != 0 )
                        {
                          sub_18000CC80(v177, 1, v180);
// Original IDA line 2606 | qword_1803BE368 -> novels.services.agent_session#139 bytes [4309105, 4309131]: "》并完成基础设定"
                          v186 = qword_1803BE368;
// Original IDA line 2607 | qword_1803BE368 -> novels.services.agent_session#139 bytes [4309105, 4309131]: "》并完成基础设定"
                          if ( *(_DWORD *)qword_1803BE368 != -1 )
// Original IDA line 2608 | qword_1803BE368 -> novels.services.agent_session#139 bytes [4309105, 4309131]: "》并完成基础设定"
                            ++*(_DWORD *)qword_1803BE368;
                          sub_18000CC80(v177, 2, v186);
// Original IDA line 2610 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          v187 = PyUnicode_Join(qword_1803BDFE8, v177);
                          v29 = *v177 < 0;
                          v188 = (int *)v187;
                          v1448 = (int *)v187;
                          if ( !v29 )
                          {
                            v8 = (*(_QWORD *)v177)-- == 1;
                            if ( v8 )
                            {
                              v189 = sub_18000CC40(v177);
                              (*(void (__fastcall **)(int *))(v189 + 48))(v177);
                            }
                          }
                          if ( v188 == nullptr )
                          {
                            sub_18000CE30(a1, &v1438);
                            if ( *v23 >= 0 )
                            {
                              v8 = (*(_QWORD *)v23)-- == 1;
                              if ( v8 )
                              {
                                v190 = sub_18000CC40(v23);
                                (*(void (__fastcall **)(int *))(v190 + 48))(v23);
                              }
                            }
                            if ( *v173 >= 0 )
                            {
                              v8 = (*(_QWORD *)v173)-- == 1;
                              if ( v8 )
                              {
                                v191 = sub_18000CC40(v173);
                                (*(void (__fastcall **)(int *))(v191 + 48))(v173);
                              }
                            }
                            v10 = v1440;
                            v151 = 297;
                            goto LABEL_2516;
                          }
// Original IDA line 2648 | qword_1803BE000 -> novels.services.agent_session#30 bytes [4307327, 4307331]: "id"
                          v1453 = (int *)sub_18037F0C0(a1, v158, qword_1803BE000);
                          v192 = v1453;
                          if ( v1453 == nullptr )
                          {
                            sub_18000CE30(a1, &v1438);
                            if ( *v23 >= 0 )
                            {
                              v8 = (*(_QWORD *)v23)-- == 1;
                              if ( v8 )
                              {
                                v193 = sub_18000CC40(v23);
                                (*(void (__fastcall **)(int *))(v193 + 48))(v23);
                              }
                            }
                            if ( *v173 >= 0 )
                            {
                              v8 = (*(_QWORD *)v173)-- == 1;
                              if ( v8 )
                              {
                                v194 = sub_18000CC40(v173);
                                (*(void (__fastcall **)(int *))(v194 + 48))(v173);
                              }
                            }
                            if ( *v188 >= 0 )
                            {
                              v8 = (*(_QWORD *)v188)-- == 1;
                              if ( v8 )
                              {
                                v195 = sub_18000CC40(v188);
                                (*(void (__fastcall **)(int *))(v195 + 48))(v188);
                              }
                            }
                            v10 = v1440;
                            v151 = 298;
                            goto LABEL_2516;
                          }
// Original IDA line 2684 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                          v1446 = (_QWORD *)sub_18037F0C0(a1, v158, qword_1803BE350);
                          if ( v1446 == nullptr )
                          {
                            sub_18000CE30(a1, &v1438);
                            if ( *v23 >= 0 )
                            {
                              v8 = (*(_QWORD *)v23)-- == 1;
                              if ( v8 )
                              {
                                v196 = sub_18000CC40(v23);
                                (*(void (__fastcall **)(int *))(v196 + 48))(v23);
                              }
                            }
                            if ( *v173 >= 0 )
                            {
                              v8 = (*(_QWORD *)v173)-- == 1;
                              if ( v8 )
                              {
                                v197 = sub_18000CC40(v173);
                                (*(void (__fastcall **)(int *))(v197 + 48))(v173);
                              }
                            }
                            if ( *v188 >= 0 )
                            {
                              v8 = (*(_QWORD *)v188)-- == 1;
                              if ( v8 )
                              {
                                v198 = sub_18000CC40(v188);
                                (*(void (__fastcall **)(int *))(v198 + 48))(v188);
                              }
                            }
                            if ( *v192 >= 0 )
                            {
                              v8 = (*(_QWORD *)v192)-- == 1;
                              if ( v8 )
                              {
                                v199 = sub_18000CC40(v192);
                                (*(void (__fastcall **)(int *))(v199 + 48))(v192);
                              }
                            }
                            v10 = v1440;
                            v151 = 299;
                            goto LABEL_2516;
                          }
// Original IDA line 2728 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                          v200 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                          v201 = (int *)v200;
                          v202 = a1;
                          if ( v200 != 0 )
                          {
// Original IDA line 2733 | qword_1803BE370 -> novels.services.agent_session#140 bytes [4309131, 4309146]: "STAGES_CREATE"
                            v204 = (int *)sub_18037F0C0(a1, v200, qword_1803BE370);
                            if ( *v201 >= 0 )
                            {
                              v8 = (*(_QWORD *)v201)-- == 1;
                              if ( v8 )
                              {
                                v205 = sub_18000CC40(v201);
                                (*(void (__fastcall **)(int *))(v205 + 48))(v201);
                              }
                            }
                            v202 = a1;
                            if ( v204 != nullptr )
                            {
                              v206 = (int *)sub_180001930(a1, v204);
                              if ( *v204 >= 0 )
                              {
                                v8 = (*(_QWORD *)v204)-- == 1;
                                if ( v8 )
                                {
                                  v207 = sub_18000CC40(v204);
                                  (*(void (__fastcall **)(int *))(v207 + 48))(v204);
                                }
                              }
                              v202 = a1;
                              if ( v206 != nullptr )
                              {
                                v1439 = (int *)sub_180372160(a1, 0);
                                v208 = v1439;
                                for ( i = (int *)sub_18000D430(v206); i != nullptr; i = (int *)sub_18000D430(v206) )
                                {
                                  v210 = v1449;
                                  v211 = v1455;
                                  v212 = v1449;
                                  v1449 = i;
                                  v213 = v1455;
                                  if ( v212 != nullptr && *v210 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v210)-- == 1;
                                    if ( v8 )
                                    {
                                      v214 = sub_18000CC40(v210);
                                      (*(void (__fastcall **)(int *))(v214 + 48))(v212);
                                      v211 = v213;
                                    }
                                  }
                                  v215 = *i;
                                  v216 = v211;
                                  v1455 = i;
                                  v217 = v215 + 1;
                                  if ( v217 != 0 )
                                    *i = v217;
                                  if ( v211 != nullptr && *v213 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v213)-- == 1;
                                    if ( v8 )
                                    {
                                      v218 = sub_18000CC40(v211);
                                      (*(void (__fastcall **)(int *))(v218 + 48))(v216);
                                    }
                                  }
// Original IDA line 2793 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                                  v219 = qword_1803BE378;
                                  v220 = PyDict_NewPresized(2);
                                  PyDict_SetItem(v220, v219, i);
// Original IDA line 2796 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"; qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                                  PyDict_SetItem(v220, qword_1803BE380, qword_1803BE2E0);
                                  v221 = v220;
                                  v208 = v1439;
                                  if ( (unsigned __int8)sub_180372560(v1439, v221) == 0
                                    || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                  {
                                    goto LABEL_452;
                                  }
                                }
                                if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
                                {
LABEL_452:
                                  sub_18000CE30(a1, &v1438);
                                  v260 = v1438;
                                  sub_18000CE70(&v1438);
                                  if ( *v206 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v206)-- == 1;
                                    if ( v8 )
                                    {
                                      v261 = sub_18000CC40(v206);
                                      (*(void (__fastcall **)(int *))(v261 + 48))(v206);
                                    }
                                  }
                                  if ( *v208 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v208)-- == 1;
                                    if ( v8 )
                                    {
                                      v262 = sub_18000CC40(v208);
                                      (*(void (__fastcall **)(int *))(v262 + 48))(v208);
                                    }
                                  }
                                  v263 = (int **)v1455;
                                  if ( v1455 != nullptr && *v1455 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1455)-- == 1;
                                    v1449 = *v263;
                                    if ( v8 )
                                    {
                                      v264 = sub_18000CC40(v263);
                                      (*(void (__fastcall **)(int **))(v264 + 48))(v263);
                                    }
                                  }
                                  v1438 = v260;
                                  sub_18000CE70(&v1438);
                                  if ( v263 != nullptr && *(int *)v263 >= 0 )
                                  {
                                    v8 = *v263 == (int *)1;
                                    *v263 = (int *)((char *)*v263 - 1);
                                    v1449 = *v263;
                                    if ( v8 )
                                    {
                                      v265 = sub_18000CC40(v263);
                                      (*(void (__fastcall **)(int **))(v265 + 48))(v263);
                                    }
                                  }
                                  v1438 = v260;
                                  v10 = v1440;
                                  v151 = 300;
                                  goto LABEL_2516;
                                }
                                if ( *v208 != -1 )
                                  ++*v208;
                                if ( *v206 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v206)-- == 1;
                                  if ( v8 )
                                  {
                                    v222 = sub_18000CC40(v206);
                                    (*(void (__fastcall **)(int *))(v222 + 48))(v206);
                                  }
                                }
                                if ( *v208 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v208)-- == 1;
                                  if ( v8 )
                                  {
                                    v223 = sub_18000CC40(v208);
                                    (*(void (__fastcall **)(int *))(v223 + 48))(v208);
                                  }
                                }
                                v224 = v1455;
                                if ( v1455 != nullptr && *v1455 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1455)-- == 1;
                                  v225 = *(int **)v224;
                                  if ( v8 )
                                  {
                                    v226 = sub_18000CC40(v224);
                                    (*(void (__fastcall **)(int *))(v226 + 48))(v224);
                                    LODWORD(v225) = *v224;
                                  }
                                  if ( (int)v225 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v224)-- == 1;
                                    v1449 = *(int **)v224;
                                    if ( v8 )
                                    {
                                      v227 = sub_18000CC40(v224);
                                      (*(void (__fastcall **)(int *))(v227 + 48))(v224);
                                    }
                                  }
                                }
                                v228 = v1448;
                                v151 = 295;
                                *(_DWORD *)(v1440 + 40) = 295;
                                v1510[0] = v1442;
                                v1529[0] = v1453;
                                v1529[1] = v1446;
                                v1510[1] = v228;
                                v1529[2] = v208;
// Original IDA line 2908 | qword_1803BE388 -> novels.services.agent_session#143 bytes [4309160, 4309189]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309162, 4309170], "value": "bookId"}, {"tag": "a", "fileRange": [4309170, 4309181], "value": "bookTitle"}, {"tag": "a", "fileRange": [4309181, 4309189], "value": "stages"}]}
                                v229 = (int *)sub_18037D650(a1, v173, v1510, v1529, qword_1803BE388);
                                if ( *v173 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v173)-- == 1;
                                  if ( v8 )
                                  {
                                    v230 = sub_18000CC40(v173);
                                    (*(void (__fastcall **)(int *))(v230 + 48))(v173);
                                  }
                                }
                                if ( *v228 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v228)-- == 1;
                                  if ( v8 )
                                  {
                                    v231 = sub_18000CC40(v228);
                                    (*(void (__fastcall **)(int *))(v231 + 48))(v228);
                                  }
                                }
                                v232 = v1453;
                                if ( *v1453 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1453)-- == 1;
                                  if ( v8 )
                                  {
                                    v233 = sub_18000CC40(v232);
                                    (*(void (__fastcall **)(int *))(v233 + 48))(v232);
                                  }
                                }
                                v234 = v1446;
                                if ( *(int *)v1446 >= 0 )
                                {
                                  v8 = (*v1446)-- == 1;
                                  if ( v8 )
                                  {
                                    v235 = sub_18000CC40(v234);
                                    (*(void (__fastcall **)(_QWORD *))(v235 + 48))(v234);
                                  }
                                }
                                if ( *v208 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v208)-- == 1;
                                  if ( v8 )
                                  {
                                    v236 = sub_18000CC40(v208);
                                    (*(void (__fastcall **)(int *))(v236 + 48))(v208);
                                  }
                                }
                                if ( v229 == nullptr )
                                {
                                  sub_18000CE30(a1, &v1438);
                                  if ( *v23 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v23)-- == 1;
                                    if ( v8 )
                                    {
LABEL_415:
                                      v237 = sub_18000CC40(v23);
                                      (*(void (__fastcall **)(int *))(v237 + 48))(v23);
                                      v10 = v1440;
                                      goto LABEL_2516;
                                    }
                                  }
LABEL_277:
                                  v10 = v1440;
                                  goto LABEL_2516;
                                }
                                v10 = v1440;
                                *(_DWORD *)(v1440 + 40) = 295;
                                v238 = (int *)sub_180376770(a1, v23, v229);
                                if ( *v23 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v23)-- == 1;
                                  if ( v8 )
                                  {
                                    v239 = sub_18000CC40(v23);
                                    (*(void (__fastcall **)(int *))(v239 + 48))(v23);
                                  }
                                }
                                if ( *v229 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v229)-- == 1;
                                  if ( v8 )
                                  {
                                    v240 = sub_18000CC40(v229);
                                    (*(void (__fastcall **)(int *))(v240 + 48))(v229);
                                  }
                                }
                                if ( v238 != nullptr )
                                {
                                  if ( *v238 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v238)-- == 1;
                                    if ( v8 )
                                    {
                                      v241 = sub_18000CC40(v238);
                                      (*(void (__fastcall **)(int *))(v241 + 48))(v238);
                                    }
                                  }
// Original IDA line 3007 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                                  v242 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                                  if ( v242 == nullptr )
                                  {
                                    sub_18000CE30(a1, &v1438);
                                    v151 = 305;
                                    goto LABEL_2516;
                                  }
                                  v151 = 305;
                                  v243 = *(_QWORD *)(v1445 + 16);
                                  *(_DWORD *)(v10 + 40) = 305;
// Original IDA line 3017 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                                  v244 = (int *)sub_18037E6C0(a1, v242, qword_1803BE0D8, v243);
                                  if ( *v242 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v242)-- == 1;
                                    if ( v8 )
                                    {
                                      v245 = sub_18000CC40(v242);
                                      (*(void (__fastcall **)(int *))(v245 + 48))(v242);
                                    }
                                  }
                                  if ( v244 != nullptr )
                                  {
                                    if ( *v244 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v244)-- == 1;
                                      if ( v8 )
                                      {
                                        v246 = sub_18000CC40(v244);
                                        (*(void (__fastcall **)(int *))(v246 + 48))(v244);
                                      }
                                    }
// Original IDA line 3038 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                                    v247 = qword_1803BE2C8;
                                    v23 = (int *)PyDict_NewPresized(2);
                                    PyDict_SetItem(v23, v247, Py_TrueStruct);
                                    v151 = 309;
// Original IDA line 3042 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                                    v248 = qword_1803BE2F8;
// Original IDA line 3043 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
                                    v249 = qword_1803BE348;
                                    v250 = v1450;
                                    *(_DWORD *)(v1440 + 40) = 309;
// Original IDA line 3046 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
                                    v251 = (int *)sub_18037E560(a1, v250, qword_1803BE278);
                                    if ( v251 == nullptr )
                                    {
                                      sub_18000CE30(a1, &v1438);
                                      goto LABEL_445;
                                    }
                                    v252 = (int *)PyDict_NewPresized(3);
                                    PyDict_SetItem(v252, v249, v251);
                                    if ( *v251 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v251)-- == 1;
                                      if ( v8 )
                                      {
                                        v253 = sub_18000CC40(v251);
                                        (*(void (__fastcall **)(int *))(v253 + 48))(v251);
                                      }
                                    }
                                    v151 = 310;
// Original IDA line 3064 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
                                    v254 = qword_1803BE1B8;
                                    v255 = *(_QWORD *)(v1445 + 16);
                                    *(_DWORD *)(v1440 + 40) = 310;
// Original IDA line 3067 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
                                    v256 = sub_18037E560(a1, v255, qword_1803BE278);
                                    v257 = (int *)v256;
                                    if ( v256 == 0 )
                                    {
                                      sub_18000CE30(a1, &v1438);
                                      if ( *v252 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v252)-- == 1;
                                        if ( v8 )
                                        {
                                          v258 = sub_18000CC40(v252);
                                          (*(void (__fastcall **)(int *))(v258 + 48))(v252);
                                        }
                                      }
LABEL_445:
                                      if ( *v23 < 0 )
                                        goto LABEL_272;
                                      v8 = (*(_QWORD *)v23)-- == 1;
                                      if ( !v8 )
                                        goto LABEL_272;
                                      goto LABEL_415;
                                    }
                                    PyDict_SetItem(v252, v254, v256);
                                    if ( *v257 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v257)-- == 1;
                                      if ( v8 )
                                      {
                                        v259 = sub_18000CC40(v257);
                                        (*(void (__fastcall **)(int *))(v259 + 48))(v257);
                                      }
                                    }
// Original IDA line 3099 | qword_1803BE390 -> novels.services.agent_session#144 bytes [4309189, 4309203]: "switchToBook"
                                    PyDict_SetItem(v252, qword_1803BE390, Py_TrueStruct);
LABEL_971:
                                    PyDict_SetItem(v23, v248, v252);
                                    if ( *v252 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v252)-- == 1;
                                      if ( v8 )
                                      {
                                        v535 = sub_18000CC40(v252);
                                        (*(void (__fastcall **)(int *))(v535 + 48))(v252);
                                      }
                                    }
                                    goto LABEL_149;
                                  }
                                }
LABEL_2475:
                                sub_18000CE30(a1, &v1438);
                                goto LABEL_2516;
                              }
                            }
                          }
                          sub_18000CE30(v202, &v1438);
                          v203 = v1438;
                          sub_18000CE70(&v1438);
                          v1438 = v203;
                          v10 = v1440;
                          v151 = 300;
                          goto LABEL_2516;
                        }
                      }
                      v151 = 297;
LABEL_309:
                      sub_18000CE30(a1, &v1438);
                      if ( *v23 >= 0 )
                      {
                        v8 = (*(_QWORD *)v23)-- == 1;
                        if ( v8 )
                        {
                          v182 = sub_18000CC40(v23);
                          (*(void (__fastcall **)(int *))(v182 + 48))(v23);
                        }
                      }
                      if ( *v173 >= 0 )
                      {
                        v8 = (*(_QWORD *)v173)-- == 1;
                        if ( v8 )
                        {
                          v183 = sub_18000CC40(v173);
                          v184 = v173;
LABEL_315:
                          (*(void (__fastcall **)(int *))(v183 + 48))(v184);
                          goto LABEL_316;
                        }
                      }
                      goto LABEL_316;
                    }
                  }
                  sub_18000CE30(a1, &v1438);
                  v10 = v1440;
                  v151 = 295;
                  goto LABEL_2516;
                }
              }
              sub_18000CE30(a1, &v1438);
              v10 = v1440;
              v151 = 293;
              goto LABEL_2516;
            }
          }
          sub_18000CE30(v163, &v1438);
          v151 = 292;
          goto LABEL_277;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 290;
      goto LABEL_2516;
    }
// Original IDA line 3177 | qword_1803BE398 -> novels.services.agent_session#145 bytes [4309203, 4309214]: "short_run"
    v266 = sub_1803836D0(v45, qword_1803BE398);
    if ( v266 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 315;
      goto LABEL_2516;
    }
    if ( v266 == 1 )
    {
      v151 = 316;
// Original IDA line 3187 | qword_1803BE3A0 -> novels.services.agent_session#146 bytes [4309214, 4309245]: "pipeline.short_fiction_runner"
      v267 = qword_1803BE3A0;
      v268 = qword_1803CA5B8;
// Original IDA line 3189 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
      v1435 = qword_1803BE028;
// Original IDA line 3190 | qword_1803BE3A8 -> novels.services.agent_session#147 bytes [4309245, 4309261]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309247, 4309261], "value": "SHORT_STAGES"}]}
      v1432 = qword_1803BE3A8;
      *(_DWORD *)(v10 + 40) = 316;
      v269 = sub_1803717E0(a1, v267, v268, Py_NoneStruct, v1432, v1435);
      v270 = (int *)v269;
      if ( v269 == 0 )
        goto LABEL_2475;
      if ( (unsigned int)sub_18000CC50(v269, PyModule_Type) != 0 )
// Original IDA line 3197 | qword_1803BE3B0 -> novels.services.agent_session#148 bytes [4309261, 4309275]: "SHORT_STAGES"
        v271 = sub_180371D80(a1, (_DWORD)v270, qword_1803CA5B8, qword_1803BE3B0, qword_1803AB250);
      else
// Original IDA line 3199 | qword_1803BE3B0 -> novels.services.agent_session#148 bytes [4309261, 4309275]: "SHORT_STAGES"
        v271 = sub_180371890(a1, v270, qword_1803BE3B0);
      v29 = *v270 < 0;
      v272 = (int *)v271;
      v1449 = (int *)v271;
      if ( !v29 )
      {
        v8 = (*(_QWORD *)v270)-- == 1;
        if ( v8 )
        {
          v273 = sub_18000CC40(v270);
          (*(void (__fastcall **)(int *))(v273 + 48))(v270);
        }
      }
      if ( v272 == nullptr )
      {
LABEL_2465:
        sub_18000CE30(a1, &v1438);
        v10 = v1440;
        goto LABEL_2516;
      }
      v151 = 317;
// Original IDA line 3220 | qword_1803BE3B8 -> novels.services.agent_session#149 bytes [4309275, 4309300]: "storage.file_repository"
      v274 = qword_1803BE3B8;
      v275 = v1440;
      v276 = qword_1803CA5B8;
// Original IDA line 3223 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
      v1436 = qword_1803BE028;
// Original IDA line 3224 | qword_1803BE3C0 -> novels.services.agent_session#150 bytes [4309300, 4309318]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309302, 4309318], "value": "FileRepository"}]}
      v1433 = qword_1803BE3C0;
      *(_DWORD *)(v1440 + 40) = 317;
      v1475 = v272;
      v277 = sub_1803717E0(a1, v274, v276, Py_NoneStruct, v1433, v1436);
      v278 = (int *)v277;
      if ( v277 == 0 )
        goto LABEL_696;
      if ( (unsigned int)sub_18000CC50(v277, PyModule_Type) != 0 )
// Original IDA line 3232 | qword_1803BE3C8 -> novels.services.agent_session#151 bytes [4309318, 4309334]: "FileRepository"
        v279 = sub_180371D80(a1, (_DWORD)v278, qword_1803CA5B8, qword_1803BE3C8, qword_1803AB250);
      else
// Original IDA line 3234 | qword_1803BE3C8 -> novels.services.agent_session#151 bytes [4309318, 4309334]: "FileRepository"
        v279 = sub_180371890(a1, v278, qword_1803BE3C8);
      v280 = (int *)v279;
      if ( *v278 >= 0 )
      {
        v8 = (*(_QWORD *)v278)-- == 1;
        if ( v8 )
        {
          v281 = sub_18000CC40(v278);
          (*(void (__fastcall **)(int *))(v281 + 48))(v278);
        }
      }
      if ( v280 == nullptr )
        goto LABEL_696;
      v1476 = v280;
// Original IDA line 3248 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
      v282 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
      v283 = (int *)v282;
      if ( v282 == 0 )
      {
        sub_18000CE30(a1, &v1438);
        v151 = 320;
        v10 = v275;
        goto LABEL_2516;
      }
      v151 = 320;
      *(_DWORD *)(v275 + 40) = 320;
// Original IDA line 3259 | qword_1803BE3D0 -> novels.services.agent_session#152 bytes [4309334, 4309347]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309336, 4309347], "value": "direction"}]}
      v284 = (int *)sub_1803770D0(a1, v282, qword_1803BE3D0);
      if ( *v283 >= 0 )
      {
        v8 = (*(_QWORD *)v283)-- == 1;
        if ( v8 )
        {
          v285 = sub_18000CC40(v283);
          (*(void (__fastcall **)(int *))(v285 + 48))(v283);
        }
      }
      if ( v284 == nullptr )
      {
LABEL_696:
        sub_18000CE30(a1, &v1438);
        v10 = v275;
        goto LABEL_2516;
      }
      v286 = sub_180001510(v284);
      if ( v286 == -1 )
        goto LABEL_495;
      if ( v286 != 1 )
      {
        if ( *v284 >= 0 )
        {
          v8 = (*(_QWORD *)v284)-- == 1;
          if ( v8 )
          {
            v288 = sub_18000CC40(v284);
            (*(void (__fastcall **)(int *))(v288 + 48))(v284);
          }
        }
// Original IDA line 3290 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
        v289 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
        v290 = (int *)v289;
        if ( v289 == 0 )
        {
          sub_18000CE30(a1, &v1438);
          v151 = 321;
          v10 = v275;
          goto LABEL_2516;
        }
        v151 = 321;
        *(_DWORD *)(v275 + 40) = 321;
// Original IDA line 3301 | qword_1803BE3D8 -> novels.services.agent_session#153 bytes [4309347, 4309362]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309349, 4309362], "value": "instruction"}]}
        v284 = (int *)sub_1803770D0(a1, v289, qword_1803BE3D8);
        if ( *v290 >= 0 )
        {
          v8 = (*(_QWORD *)v290)-- == 1;
          if ( v8 )
          {
            v291 = sub_18000CC40(v290);
            (*(void (__fastcall **)(int *))(v291 + 48))(v290);
          }
        }
        if ( v284 == nullptr )
          goto LABEL_696;
        v292 = sub_180001510(v284);
        if ( v292 == -1 )
          goto LABEL_495;
        if ( v292 != 1 )
        {
          if ( *v284 >= 0 )
          {
            v8 = (*(_QWORD *)v284)-- == 1;
            if ( v8 )
            {
              v293 = sub_18000CC40(v284);
              (*(void (__fastcall **)(int *))(v293 + 48))(v284);
            }
          }
// Original IDA line 3327 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v294 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
          v295 = (int *)v294;
          if ( v294 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 322;
            v10 = v275;
            goto LABEL_2516;
          }
          v151 = 322;
          *(_DWORD *)(v275 + 40) = 322;
// Original IDA line 3338 | qword_1803BE3E0 -> novels.services.agent_session#154 bytes [4309362, 4309373]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309364, 4309373], "value": "summary"}]}
          v284 = (int *)sub_1803770D0(a1, v294, qword_1803BE3E0);
          if ( *v295 >= 0 )
          {
            v8 = (*(_QWORD *)v295)-- == 1;
            if ( v8 )
            {
              v296 = sub_18000CC40(v295);
              (*(void (__fastcall **)(int *))(v296 + 48))(v295);
            }
          }
          if ( v284 == nullptr )
            goto LABEL_696;
          v297 = sub_180001510(v284);
          if ( v297 == -1 )
            goto LABEL_495;
          if ( v297 != 1 )
          {
            if ( *v284 >= 0 )
            {
              v8 = (*(_QWORD *)v284)-- == 1;
              if ( v8 )
              {
                v298 = sub_18000CC40(v284);
                (*(void (__fastcall **)(int *))(v298 + 48))(v284);
              }
            }
// Original IDA line 3364 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
            v299 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
            v300 = (int *)v299;
            if ( v299 == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 323;
              v10 = v275;
              goto LABEL_2516;
            }
            v151 = 323;
            *(_DWORD *)(v275 + 40) = 323;
// Original IDA line 3375 | qword_1803BE3E8 -> novels.services.agent_session#155 bytes [4309373, 4309382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309375, 4309382], "value": "title"}]}
            v284 = (int *)sub_1803770D0(a1, v299, qword_1803BE3E8);
            if ( *v300 >= 0 )
            {
              v8 = (*(_QWORD *)v300)-- == 1;
              if ( v8 )
              {
                v301 = sub_18000CC40(v300);
                (*(void (__fastcall **)(int *))(v301 + 48))(v300);
              }
            }
            if ( v284 == nullptr )
              goto LABEL_696;
            v302 = sub_180001510(v284);
            if ( v302 == -1 )
              goto LABEL_495;
            if ( v302 != 1 )
            {
              if ( *v284 >= 0 )
              {
                v8 = (*(_QWORD *)v284)-- == 1;
                if ( v8 )
                {
                  v303 = sub_18000CC40(v284);
                  (*(void (__fastcall **)(int *))(v303 + 48))(v284);
                }
              }
// Original IDA line 3401 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              v284 = (int *)qword_1803BDFE8;
// Original IDA line 3402 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 3403 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                ++*(_DWORD *)qword_1803BDFE8;
            }
          }
        }
      }
      v304 = (int *)sub_1803749A0(v284);
      if ( *v284 >= 0 )
      {
        v8 = (*(_QWORD *)v284)-- == 1;
        if ( v8 )
        {
          v305 = sub_18000CC40(v284);
          (*(void (__fastcall **)(int *))(v305 + 48))(v284);
        }
      }
      if ( v304 != nullptr )
      {
// Original IDA line 3420 | qword_1803BDFF0 -> novels.services.agent_session#28 bytes [4307300, 4307307]: "strip"
        v306 = (int *)sub_18037F0C0(a1, v304, qword_1803BDFF0);
        if ( *v304 >= 0 )
        {
          v8 = (*(_QWORD *)v304)-- == 1;
          if ( v8 )
          {
            v307 = sub_18000CC40(v304);
            (*(void (__fastcall **)(int *))(v307 + 48))(v304);
          }
        }
        if ( v306 != nullptr )
        {
          v151 = 319;
          *(_DWORD *)(v275 + 40) = 319;
          v308 = sub_180376210(a1, v306);
          if ( *v306 >= 0 )
          {
            v8 = (*(_QWORD *)v306)-- == 1;
            if ( v8 )
            {
              v309 = sub_18000CC40(v306);
              (*(void (__fastcall **)(int *))(v309 + 48))(v306);
            }
          }
          v310 = a1;
          if ( v308 == 0 )
            goto LABEL_2060;
          v1477 = (int *)v308;
// Original IDA line 3448 | qword_1803BE3F0 -> novels.services.agent_session#156 bytes [4309382, 4309391]: "slugify"
          v177 = (int *)sub_18037F0C0(a1, v280, qword_1803BE3F0);
          if ( v177 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            v10 = v1440;
            v151 = 326;
            goto LABEL_2516;
          }
// Original IDA line 3456 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v311 = (int *)sub_18037F0C0(a1, v1443, qword_1803BDFD8);
          if ( v311 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v312 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v312 + 48))(v177);
              }
            }
            v10 = v1440;
            v151 = 327;
            goto LABEL_2516;
          }
          v151 = 327;
          *(_DWORD *)(v1440 + 40) = 327;
// Original IDA line 3475 | qword_1803BE3F8 -> novels.services.agent_session#157 bytes [4309391, 4309402]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309393, 4309402], "value": "storyId"}]}
          v23 = (int *)sub_1803770D0(a1, v311, qword_1803BE3F8);
          if ( *v311 >= 0 )
          {
            v8 = (*(_QWORD *)v311)-- == 1;
            if ( v8 )
            {
              v313 = sub_18000CC40(v311);
              (*(void (__fastcall **)(int *))(v313 + 48))(v311);
            }
          }
          if ( v23 == nullptr )
            goto LABEL_565;
          v315 = sub_180001510(v23);
          if ( v315 == -1 )
            goto LABEL_571;
          if ( v315 == 1 )
            goto LABEL_597;
          if ( *v23 >= 0 )
          {
            v8 = (*(_QWORD *)v23)-- == 1;
            if ( v8 )
            {
              v318 = sub_18000CC40(v23);
              (*(void (__fastcall **)(int *))(v318 + 48))(v23);
            }
          }
// Original IDA line 3501 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v319 = (int *)sub_18037F0C0(a1, v1443, qword_1803BDFD8);
          v314 = a1;
          if ( v319 == nullptr )
            goto LABEL_566;
          *(_DWORD *)(v1440 + 40) = 327;
// Original IDA line 3506 | qword_1803BE3E8 -> novels.services.agent_session#155 bytes [4309373, 4309382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309375, 4309382], "value": "title"}]}
          v23 = (int *)sub_1803770D0(a1, v319, qword_1803BE3E8);
          if ( *v319 >= 0 )
          {
            v8 = (*(_QWORD *)v319)-- == 1;
            if ( v8 )
            {
              v320 = sub_18000CC40(v319);
              (*(void (__fastcall **)(int *))(v320 + 48))(v319);
            }
          }
          if ( v23 == nullptr )
            goto LABEL_565;
          v321 = sub_180001510(v23);
          if ( v321 == -1 )
          {
LABEL_571:
            v316 = a1;
LABEL_572:
            sub_18000CE30(v316, &v1438);
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v317 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v317 + 48))(v177);
              }
            }
            goto LABEL_575;
          }
          if ( v321 == 1 )
          {
LABEL_597:
            v324 = (int *)sub_1803749A0(v23);
            if ( *v23 >= 0 )
            {
              v8 = (*(_QWORD *)v23)-- == 1;
              if ( v8 )
              {
                v325 = sub_18000CC40(v23);
                (*(void (__fastcall **)(int *))(v325 + 48))(v23);
              }
            }
            v314 = a1;
            if ( v324 == nullptr )
              goto LABEL_566;
// Original IDA line 3552 | qword_1803BE408 -> novels.services.agent_session#159 bytes [4309410, 4309415]: {"type": ":", "items": [{"tag": "n", "fileRange": [4309411, 4309412], "value": null}, {"tag": "l", "fileRange": [4309412, 4309414], "value": 40}, {"tag": "n", "fileRange": [4309414, 4309415], "value": null}]}
            v326 = (int *)sub_180001B80(a1, v324, qword_1803BE408);
            if ( *v324 >= 0 )
            {
              v8 = (*(_QWORD *)v324)-- == 1;
              if ( v8 )
              {
                v327 = sub_18000CC40(v324);
                (*(void (__fastcall **)(int *))(v327 + 48))(v324);
              }
            }
            v314 = a1;
            if ( v326 == nullptr )
              goto LABEL_566;
            v275 = v1440;
            v151 = 326;
            *(_DWORD *)(v1440 + 40) = 326;
            v328 = sub_180376770(a1, v177, v326);
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v329 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v329 + 48))(v177);
              }
            }
            if ( *v326 >= 0 )
            {
              v8 = (*(_QWORD *)v326)-- == 1;
              if ( v8 )
              {
                v330 = sub_18000CC40(v326);
                (*(void (__fastcall **)(int *))(v330 + 48))(v326);
              }
            }
            if ( v328 == 0 )
              goto LABEL_696;
            v1478 = (int *)v328;
// Original IDA line 3590 | qword_1803BE410 -> novels.services.agent_session#160 bytes [4309415, 4309429]: "shortStoryId"
            if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE410, v328) == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 329;
              v10 = v275;
              goto LABEL_2516;
            }
// Original IDA line 3597 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
            v331 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
            v332 = (int *)v331;
            if ( v331 == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 330;
              v10 = v275;
              goto LABEL_2516;
            }
            v333 = *(_QWORD *)(v1445 + 16);
            v151 = 330;
            *(_DWORD *)(v275 + 40) = 330;
// Original IDA line 3609 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
            v334 = (int *)sub_18037E6C0(a1, v331, qword_1803BE0D8, v333);
            if ( *v332 >= 0 )
            {
              v8 = (*(_QWORD *)v332)-- == 1;
              if ( v8 )
              {
                v335 = sub_18000CC40(v332);
                (*(void (__fastcall **)(int *))(v335 + 48))(v332);
              }
            }
            if ( v334 == nullptr )
              goto LABEL_696;
            if ( *v334 >= 0 )
            {
              v8 = (*(_QWORD *)v334)-- == 1;
              if ( v8 )
              {
                v336 = sub_18000CC40(v334);
                (*(void (__fastcall **)(int *))(v336 + 48))(v334);
              }
            }
            v337 = sub_180370DC0(a1, v1443);
            v338 = v337;
            if ( v337 == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 331;
              v10 = v275;
              goto LABEL_2516;
            }
            v1479 = (int *)v337;
// Original IDA line 3640 | qword_1803BE418 -> novels.services.agent_session#161 bytes [4309429, 4309438]: "storyId"
            PyDict_SetItem(v337, qword_1803BE418, v328);
// Original IDA line 3641 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
            v339 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
            v340 = (int *)v339;
            if ( v339 == 0 )
              goto LABEL_627;
// Original IDA line 3645 | qword_1803BE420 -> novels.services.agent_session#162 bytes [4309438, 4309457]: "run_short_fiction"
            v341 = (int *)sub_18037F0C0(a1, v339, qword_1803BE420);
            if ( *v340 >= 0 )
            {
              v8 = (*(_QWORD *)v340)-- == 1;
              if ( v8 )
              {
                v342 = sub_18000CC40(v340);
                (*(void (__fastcall **)(int *))(v342 + 48))(v340);
              }
            }
            if ( v341 == nullptr )
            {
LABEL_627:
              sub_18000CE30(a1, &v1438);
              v151 = 333;
              v10 = v275;
              goto LABEL_2516;
            }
            v343 = v1456;
            v151 = 333;
            *(_DWORD *)(v275 + 40) = 333;
            v1493 = v343;
            v1494 = v338;
// Original IDA line 3668 | qword_1803BE340 -> novels.services.agent_session#134 bytes [4309052, 4309061]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309054, 4309061], "value": "model"}]}
            v344 = sub_18037D030(a1, v341, &v1494, &v1493, qword_1803BE340);
            v29 = *v341 < 0;
            v345 = (int *)v344;
            v1454 = (_QWORD *)v344;
            if ( !v29 )
            {
              v8 = (*(_QWORD *)v341)-- == 1;
              if ( v8 )
              {
                v346 = sub_18000CC40(v341);
                (*(void (__fastcall **)(int *))(v346 + 48))(v341);
              }
            }
            if ( v345 == nullptr )
              goto LABEL_696;
            v1457 = v345;
// Original IDA line 3684 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
            v347 = sub_18037F0C0(a1, v345, qword_1803BDFD8);
            v348 = (int *)v347;
            if ( v347 == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 334;
              v10 = v275;
              goto LABEL_2516;
            }
            v151 = 334;
            *(_DWORD *)(v275 + 40) = 334;
// Original IDA line 3695 | qword_1803BE3F8 -> novels.services.agent_session#157 bytes [4309391, 4309402]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309393, 4309402], "value": "storyId"}]}
            v284 = (int *)sub_1803770D0(a1, v347, qword_1803BE3F8);
            if ( *v348 >= 0 )
            {
              v8 = (*(_QWORD *)v348)-- == 1;
              if ( v8 )
              {
                v349 = sub_18000CC40(v348);
                (*(void (__fastcall **)(int *))(v349 + 48))(v348);
              }
            }
            if ( v284 == nullptr )
              goto LABEL_696;
            v350 = sub_180001510(v284);
            if ( v350 != -1 )
            {
              if ( v350 != 1 )
              {
                if ( *v284 >= 0 )
                {
                  v8 = (*(_QWORD *)v284)-- == 1;
                  if ( v8 )
                  {
                    v351 = sub_18000CC40(v284);
                    (*(void (__fastcall **)(int *))(v351 + 48))(v284);
                  }
                }
// Original IDA line 3721 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                v284 = (int *)qword_1803BDFE8;
// Original IDA line 3722 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 3723 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                  ++*(_DWORD *)qword_1803BDFE8;
              }
              v352 = (int *)sub_1803749A0(v284);
              if ( *v284 >= 0 )
              {
                v8 = (*(_QWORD *)v284)-- == 1;
                if ( v8 )
                {
                  v353 = sub_18000CC40(v284);
                  (*(void (__fastcall **)(int *))(v353 + 48))(v284);
                }
              }
              if ( v352 == nullptr )
                goto LABEL_696;
// Original IDA line 3737 | qword_1803BE410 -> novels.services.agent_session#160 bytes [4309415, 4309429]: "shortStoryId"
              v354 = sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE410, v352);
              if ( *v352 >= 0 )
              {
                v8 = (*(_QWORD *)v352)-- == 1;
                if ( v8 )
                {
                  v355 = sub_18000CC40(v352);
                  (*(void (__fastcall **)(int *))(v355 + 48))(v352);
                }
              }
              if ( v354 == 0 )
                goto LABEL_696;
// Original IDA line 3749 | qword_1803BE0D0 -> novels.services.agent_session#56 bytes [4307684, 4307693]: "isDraft"
              if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE0D0, Py_FalseStruct) == 0 )
              {
                sub_18000CE30(a1, &v1438);
                v151 = 335;
                v10 = v275;
                goto LABEL_2516;
              }
// Original IDA line 3756 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
              v356 = sub_18037F0C0(a1, v345, qword_1803BDFD8);
              v357 = (int *)v356;
              if ( v356 == 0 )
              {
                sub_18000CE30(a1, &v1438);
                v151 = 336;
                v10 = v275;
                goto LABEL_2516;
              }
              v151 = 336;
              *(_DWORD *)(v275 + 40) = 336;
// Original IDA line 3767 | qword_1803BE3E8 -> novels.services.agent_session#155 bytes [4309373, 4309382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309375, 4309382], "value": "title"}]}
              v284 = (int *)sub_1803770D0(a1, v356, qword_1803BE3E8);
              if ( *v357 >= 0 )
              {
                v8 = (*(_QWORD *)v357)-- == 1;
                if ( v8 )
                {
                  v358 = sub_18000CC40(v357);
                  (*(void (__fastcall **)(int *))(v358 + 48))(v357);
                }
              }
              if ( v284 == nullptr )
                goto LABEL_696;
              v359 = sub_180001510(v284);
              if ( v359 != -1 )
              {
                if ( *v284 >= 0 )
                {
                  v8 = (*(_QWORD *)v284)-- == 1;
                  if ( v8 )
                  {
                    v360 = sub_18000CC40(v284);
                    (*(void (__fastcall **)(int *))(v360 + 48))(v284);
                  }
                }
                if ( v359 != 0 )
                {
// Original IDA line 3793 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                  v361 = sub_18037F0C0(a1, v345, qword_1803BDFD8);
                  v362 = (int *)v361;
                  if ( v361 == 0 )
                  {
                    sub_18000CE30(a1, &v1438);
                    v151 = 337;
                    v10 = v275;
                    goto LABEL_2516;
                  }
                  v151 = 337;
                  *(_DWORD *)(v275 + 40) = 337;
// Original IDA line 3804 | qword_1803BE3E8 -> novels.services.agent_session#155 bytes [4309373, 4309382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309375, 4309382], "value": "title"}]}
                  v363 = (int *)sub_1803770D0(a1, v361, qword_1803BE3E8);
                  if ( *v362 >= 0 )
                  {
                    v8 = (*(_QWORD *)v362)-- == 1;
                    if ( v8 )
                    {
                      v364 = sub_18000CC40(v362);
                      (*(void (__fastcall **)(int *))(v364 + 48))(v362);
                    }
                  }
                  if ( v363 == nullptr )
                    goto LABEL_696;
                  v365 = (int *)sub_1803749A0(v363);
                  if ( *v363 >= 0 )
                  {
                    v8 = (*(_QWORD *)v363)-- == 1;
                    if ( v8 )
                    {
                      v366 = sub_18000CC40(v363);
                      (*(void (__fastcall **)(int *))(v366 + 48))(v363);
                    }
                  }
                  if ( v365 == nullptr )
                    goto LABEL_696;
// Original IDA line 3828 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                  v367 = sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE350, v365);
                  if ( *v365 >= 0 )
                  {
                    v8 = (*(_QWORD *)v365)-- == 1;
                    if ( v8 )
                    {
                      v368 = sub_18000CC40(v365);
                      (*(void (__fastcall **)(int *))(v368 + 48))(v365);
                    }
                  }
                  if ( v367 == 0 )
                    goto LABEL_696;
                }
// Original IDA line 3841 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
                {
                  sub_18000CE30(a1, &v1438);
                  v151 = 338;
                  v10 = v275;
                  goto LABEL_2516;
                }
// Original IDA line 3848 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                v369 = sub_18037F0C0(a1, v345, qword_1803BDFD8);
                v370 = (int *)v369;
                if ( v369 == 0 )
                {
                  sub_18000CE30(a1, &v1438);
                  v151 = 339;
                  v10 = v275;
                  goto LABEL_2516;
                }
                v151 = 339;
                *(_DWORD *)(v275 + 40) = 339;
// Original IDA line 3859 | qword_1803BE428 -> novels.services.agent_session#163 bytes [4309457, 4309481]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309459, 4309481], "value": "awaitingConfirmation"}]}
                v284 = (int *)sub_1803770D0(a1, v369, qword_1803BE428);
                if ( *v370 >= 0 )
                {
                  v8 = (*(_QWORD *)v370)-- == 1;
                  if ( v8 )
                  {
                    v371 = sub_18000CC40(v370);
                    (*(void (__fastcall **)(int *))(v371 + 48))(v370);
                  }
                }
                if ( v284 == nullptr )
                  goto LABEL_696;
                v372 = sub_180001510(v284);
                if ( v372 != -1 )
                {
                  if ( *v284 >= 0 )
                  {
                    v8 = (*(_QWORD *)v284)-- == 1;
                    if ( v8 )
                    {
                      v373 = sub_18000CC40(v284);
                      (*(void (__fastcall **)(int *))(v373 + 48))(v284);
                    }
                  }
                  if ( v372 != 0 )
                  {
                    *(_DWORD *)(v275 + 40) = 342;
                    v374 = sub_180376770(a1, PyEnum_Type, v272);
                    v375 = (int *)v374;
                    v376 = a1;
                    if ( v374 != 0 )
                    {
                      v377 = sub_180001930(a1, v374);
                      v29 = *v375 < 0;
                      v378 = (int *)v377;
                      v1449 = (int *)v377;
                      if ( !v29 )
                      {
                        v8 = (*(_QWORD *)v375)-- == 1;
                        if ( v8 )
                        {
                          v379 = sub_18000CC40(v375);
                          (*(void (__fastcall **)(int *))(v379 + 48))(v375);
                        }
                      }
                      v376 = a1;
                      if ( v378 != nullptr )
                      {
                        v1453 = (int *)sub_180372160(a1, 0);
                        v381 = v1453;
                        v382 = v1447;
                        v383 = sub_18000D430(v378);
                        v384 = v1447;
                        if ( v383 != 0 )
                        {
                          while ( 1 )
                          {
                            v385 = (int *)v1446;
                            v1439 = v384;
                            v386 = v382;
                            v1446 = (_QWORD *)v383;
                            v387 = v385;
                            if ( v385 != nullptr && *v385 >= 0 )
                            {
                              v8 = (*(_QWORD *)v385)-- == 1;
                              if ( v8 )
                              {
                                v388 = sub_18000CC40(v385);
                                (*(void (__fastcall **)(int *))(v388 + 48))(v387);
                              }
                            }
                            v389 = sub_180001A60(v383);
                            v390 = (int *)v389;
                            if ( v389 == 0 )
                            {
                              sub_18000CE30(a1, &v1438);
                              v427 = v1438;
                              sub_18000CE70(&v1438);
                              v1438 = v427;
                              goto LABEL_837;
                            }
                            v391 = (int *)sub_18002A560(a1, &v1438, v389, 0, 2);
                            v392 = nullptr;
                            v393 = nullptr;
                            if ( v391 == nullptr )
                              break;
                            v392 = v391;
                            v394 = sub_18002A560(a1, &v1438, v390, 1, 2);
                            v395 = (int *)v394;
                            if ( v394 == 0 )
                              break;
                            v393 = (int *)v394;
                            if ( (unsigned __int8)sub_180065DB0(a1, &v1438, v390, 2) == 0 )
                              break;
                            if ( *v390 >= 0 )
                            {
                              v8 = (*(_QWORD *)v390)-- == 1;
                              if ( v8 )
                              {
                                v396 = sub_18000CC40(v390);
                                (*(void (__fastcall **)(int *))(v396 + 48))(v390);
                              }
                            }
                            v397 = *v391;
                            v382 = v391;
                            v398 = v1447;
                            v1447 = v391;
                            v399 = v397 + 1;
                            if ( v399 != 0 )
                              *v391 = v399;
                            if ( v398 != nullptr && *v386 >= 0 )
                            {
                              v8 = (*(_QWORD *)v386)-- == 1;
                              if ( v8 )
                              {
                                v400 = sub_18000CC40(v398);
                                (*(void (__fastcall **)(int *))(v400 + 48))(v398);
                              }
                            }
                            if ( *v391 >= 0 )
                            {
                              v8 = (*(_QWORD *)v391)-- == 1;
                              if ( v8 )
                              {
                                v401 = sub_18000CC40(v391);
                                (*(void (__fastcall **)(int *))(v401 + 48))(v391);
                              }
                            }
                            v402 = *v395;
                            v384 = v395;
                            v403 = v1465;
                            v1465 = v395;
                            v404 = v402 + 1;
                            if ( v404 != 0 )
                              *v395 = v404;
                            if ( v403 != nullptr && *v1439 >= 0 )
                            {
                              v8 = (*(_QWORD *)v1439)-- == 1;
                              if ( v8 )
                              {
                                v405 = sub_18000CC40(v403);
                                (*(void (__fastcall **)(int *))(v405 + 48))(v403);
                              }
                            }
                            if ( *v395 >= 0 )
                            {
                              v8 = (*(_QWORD *)v395)-- == 1;
                              if ( v8 )
                              {
                                v406 = sub_18000CC40(v395);
                                (*(void (__fastcall **)(int *))(v406 + 48))(v395);
                              }
                            }
// Original IDA line 4012 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                            v407 = qword_1803BE378;
                            v408 = (int *)PyDict_NewPresized(2);
                            PyDict_SetItem(v408, v407, v395);
// Original IDA line 4015 | qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                            v409 = qword_1803BE380;
// Original IDA line 4016 | qword_1803BE430 -> novels.services.agent_session#164 bytes [4309481, 4309483]: 3
                            v410 = sub_180384830(v382, qword_1803BE430);
                            if ( v410 == -1 )
                            {
LABEL_769:
                              sub_18000CE30(a1, &v1438);
                              if ( *v408 >= 0 )
                              {
                                v8 = (*(_QWORD *)v408)-- == 1;
                                if ( v8 )
                                {
                                  v422 = sub_18000CC40(v408);
                                  (*(void (__fastcall **)(int *))(v422 + 48))(v408);
                                }
                              }
                              goto LABEL_837;
                            }
                            if ( v410 == 1 )
                            {
// Original IDA line 4034 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"
                              v411 = qword_1803BE2E0;
                            }
                            else
                            {
// Original IDA line 4038 | qword_1803BE430 -> novels.services.agent_session#164 bytes [4309481, 4309483]: 3
                              if ( (unsigned int)sub_180383A30(v382, qword_1803BE430) == -1 )
                                goto LABEL_769;
// Original IDA line 4040 | qword_1803BE438 -> novels.services.agent_session#165 bytes [4309483, 4309492]: "pending"
                              v411 = qword_1803BE438;
                            }
                            PyDict_SetItem(v408, v409, v411);
                            v381 = v1453;
                            v412 = sub_180372560(v1453, v408);
                            v413 = a1;
                            if ( v412 == 0 )
                              goto LABEL_836;
                            if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                              goto LABEL_835;
                            v378 = v1449;
                            v383 = sub_18000D430(v1449);
                            if ( v383 == 0 )
                              goto LABEL_746;
                          }
                          v423 = v1438;
                          sub_18000CE70(&v1438);
                          if ( *v390 >= 0 )
                          {
                            v8 = (*(_QWORD *)v390)-- == 1;
                            if ( v8 )
                            {
                              v424 = sub_18000CC40(v390);
                              (*(void (__fastcall **)(int *))(v424 + 48))(v390);
                            }
                          }
                          v1438 = v423;
                          sub_18000CE70(&v1438);
                          if ( v392 != nullptr && *v392 >= 0 )
                          {
                            v8 = (*(_QWORD *)v392)-- == 1;
                            if ( v8 )
                            {
                              v425 = sub_18000CC40(v392);
                              (*(void (__fastcall **)(int *))(v425 + 48))(v392);
                            }
                          }
                          if ( v393 != nullptr && *v393 >= 0 )
                          {
                            v8 = (*(_QWORD *)v393)-- == 1;
                            if ( v8 )
                            {
                              v426 = sub_18000CC40(v393);
                              (*(void (__fastcall **)(int *))(v426 + 48))(v393);
                            }
                          }
                          v382 = v1447;
                          v1438 = v423;
                          goto LABEL_837;
                        }
LABEL_746:
                        if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
                        {
LABEL_835:
                          v413 = a1;
LABEL_836:
                          sub_18000CE30(v413, &v1438);
LABEL_837:
                          v457 = v1438;
                          sub_18000CE70(&v1438);
                          v458 = v1449;
                          if ( *v1449 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1449)-- == 1;
                            if ( v8 )
                            {
                              v459 = sub_18000CC40(v458);
                              (*(void (__fastcall **)(int *))(v459 + 48))(v458);
                            }
                          }
                          v460 = v1453;
                          if ( *v1453 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1453)-- == 1;
                            if ( v8 )
                            {
                              v461 = sub_18000CC40(v460);
                              (*(void (__fastcall **)(int *))(v461 + 48))(v460);
                            }
                          }
                          v462 = v1446;
                          if ( v1446 != nullptr && *(int *)v1446 >= 0 )
                          {
                            v8 = (*v1446)-- == 1;
                            if ( v8 )
                            {
                              v463 = sub_18000CC40(v462);
                              (*(void (__fastcall **)(_QWORD *))(v463 + 48))(v462);
                            }
                          }
                          v1438 = v457;
                          sub_18000CE70(&v1438);
                          v464 = v1465;
                          if ( v382 != nullptr && *v382 >= 0 )
                          {
                            v8 = (*(_QWORD *)v382)-- == 1;
                            if ( v8 )
                            {
                              v465 = sub_18000CC40(v382);
                              (*(void (__fastcall **)(int *))(v465 + 48))(v382);
                            }
                          }
                          if ( v464 != nullptr && *v464 >= 0 )
                          {
                            v8 = (*(_QWORD *)v464)-- == 1;
                            if ( v8 )
                            {
                              v466 = sub_18000CC40(v464);
                              (*(void (__fastcall **)(int *))(v466 + 48))(v464);
                            }
                          }
                          v1438 = v457;
                          v10 = v1440;
                          v151 = 340;
                          goto LABEL_2516;
                        }
                        if ( *v381 != -1 )
                          ++*v381;
                        if ( *v378 >= 0 )
                        {
                          v8 = (*(_QWORD *)v378)-- == 1;
                          if ( v8 )
                          {
                            v414 = sub_18000CC40(v378);
                            (*(void (__fastcall **)(int *))(v414 + 48))(v378);
                          }
                        }
                        if ( *v381 >= 0 )
                        {
                          v8 = (*(_QWORD *)v381)-- == 1;
                          if ( v8 )
                          {
                            v415 = sub_18000CC40(v381);
                            (*(void (__fastcall **)(int *))(v415 + 48))(v381);
                          }
                        }
                        v416 = v1446;
                        if ( v1446 != nullptr && *(int *)v1446 >= 0 )
                        {
                          v8 = (*v1446)-- == 1;
                          if ( v8 )
                          {
                            v417 = sub_18000CC40(v416);
                            (*(void (__fastcall **)(_QWORD *))(v417 + 48))(v416);
                          }
                        }
                        if ( v382 != nullptr && *v382 >= 0 )
                        {
                          v8 = (*(_QWORD *)v382)-- == 1;
                          if ( v8 )
                          {
                            v418 = sub_18000CC40(v382);
                            (*(void (__fastcall **)(int *))(v418 + 48))(v382);
                          }
                        }
                        if ( v384 != nullptr && *v384 >= 0 )
                        {
                          v8 = (*(_QWORD *)v384)-- == 1;
                          if ( v8 )
                          {
                            v419 = sub_18000CC40(v384);
                            (*(void (__fastcall **)(int *))(v419 + 48))(v384);
                          }
                        }
                        v1480 = v381;
// Original IDA line 4205 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                        v420 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
                        v421 = (int *)v420;
                        if ( v420 != 0 )
                        {
// Original IDA line 4209 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                          v177 = (int *)sub_18037F0C0(a1, v420, qword_1803BE0B8);
                          if ( *v421 >= 0 )
                          {
                            v8 = (*(_QWORD *)v421)-- == 1;
                            if ( v8 )
                            {
                              v428 = sub_18000CC40(v421);
                              (*(void (__fastcall **)(int *))(v428 + 48))(v421);
                            }
                          }
                          if ( v177 != nullptr )
                          {
// Original IDA line 4221 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                            v429 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                            if ( v429 == nullptr )
                            {
                              sub_18000CE30(a1, &v1438);
                              if ( *v177 >= 0 )
                              {
                                v8 = (*(_QWORD *)v177)-- == 1;
                                if ( v8 )
                                {
                                  v430 = sub_18000CC40(v177);
                                  (*(void (__fastcall **)(int *))(v430 + 48))(v177);
                                }
                              }
                              v10 = v1440;
                              v151 = 344;
                              goto LABEL_2516;
                            }
// Original IDA line 4238 | qword_1803BE440 -> novels.services.agent_session#166 bytes [4309492, 4309515]: "awaiting_confirmation"
                            v431 = qword_1803BE440;
                            v10 = v1440;
                            v151 = 344;
                            v432 = v1454;
                            *(_DWORD *)(v1440 + 40) = 344;
                            v1530[1] = v431;
                            v1530[0] = v432;
                            v1530[2] = v381;
// Original IDA line 4246 | qword_1803BE448 -> novels.services.agent_session#167 bytes [4309515, 4309589]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309517, 4309536], "value": "short_fiction_run"}, {"tag": "u", "fileRange": [4309536, 4309589], "value": "大纲已生成，等待确认后继续撰写正文"}]}; qword_1803BE450 -> novels.services.agent_session#168 bytes [4309589, 4309616]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309591, 4309600], "value": "details"}, {"tag": "a", "fileRange": [4309600, 4309608], "value": "status"}, {"tag": "a", "fileRange": [4309608, 4309616], "value": "stages"}]}
                            v433 = (int *)sub_18037D8C0(a1, v429, qword_1803BE448, v1530, qword_1803BE450);
                            if ( *v429 >= 0 )
                            {
                              v8 = (*(_QWORD *)v429)-- == 1;
                              if ( v8 )
                              {
                                v434 = sub_18000CC40(v429);
                                (*(void (__fastcall **)(int *))(v434 + 48))(v429);
                              }
                            }
                            v435 = a1;
                            if ( v433 == nullptr )
                              goto LABEL_798;
                            *(_DWORD *)(v10 + 40) = 344;
                            v438 = (int *)sub_180376770(a1, v177, v433);
                            if ( *v177 >= 0 )
                            {
                              v8 = (*(_QWORD *)v177)-- == 1;
                              if ( v8 )
                              {
                                v439 = sub_18000CC40(v177);
                                (*(void (__fastcall **)(int *))(v439 + 48))(v177);
                              }
                            }
                            if ( *v433 >= 0 )
                            {
                              v8 = (*(_QWORD *)v433)-- == 1;
                              if ( v8 )
                              {
                                v440 = sub_18000CC40(v433);
                                (*(void (__fastcall **)(int *))(v440 + 48))(v433);
                              }
                            }
                            if ( v438 == nullptr )
                              goto LABEL_2475;
                            if ( *v438 >= 0 )
                            {
                              v8 = (*(_QWORD *)v438)-- == 1;
                              if ( v8 )
                              {
                                v441 = sub_18000CC40(v438);
                                (*(void (__fastcall **)(int *))(v441 + 48))(v438);
                              }
                            }
// Original IDA line 4290 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                            v442 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                            if ( v442 == nullptr )
                            {
                              sub_18000CE30(a1, &v1438);
                              v151 = 351;
                              goto LABEL_2516;
                            }
                            v151 = 351;
                            v443 = *(_QWORD *)(v1445 + 16);
                            *(_DWORD *)(v10 + 40) = 351;
// Original IDA line 4300 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                            v444 = (int *)sub_18037E6C0(a1, v442, qword_1803BE0D8, v443);
                            if ( *v442 >= 0 )
                            {
                              v8 = (*(_QWORD *)v442)-- == 1;
                              if ( v8 )
                              {
                                v445 = sub_18000CC40(v442);
                                (*(void (__fastcall **)(int *))(v445 + 48))(v442);
                              }
                            }
                            if ( v444 == nullptr )
                              goto LABEL_2475;
                            if ( *v444 >= 0 )
                            {
                              v8 = (*(_QWORD *)v444)-- == 1;
                              if ( v8 )
                              {
                                v446 = sub_18000CC40(v444);
                                (*(void (__fastcall **)(int *))(v446 + 48))(v444);
                              }
                            }
// Original IDA line 4321 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                            v447 = qword_1803BE2C8;
                            v23 = (int *)PyDict_NewPresized(2);
                            PyDict_SetItem(v23, v447, Py_TrueStruct);
// Original IDA line 4324 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                            v448 = qword_1803BE2F8;
// Original IDA line 4325 | qword_1803BE458 -> novels.services.agent_session#169 bytes [4309616, 4309638]: "awaitingConfirmation"
                            v449 = qword_1803BE458;
                            v70 = (int *)PyDict_NewPresized(4);
                            PyDict_SetItem(v70, v449, Py_TrueStruct);
// Original IDA line 4328 | qword_1803BE460 -> novels.services.agent_session#170 bytes [4309638, 4309650]: "checkpoint"
                            PyDict_SetItem(v70, qword_1803BE460, v432);
// Original IDA line 4329 | qword_1803BE468 -> novels.services.agent_session#171 bytes [4309650, 4309664]: "shortFiction"
                            PyDict_SetItem(v70, qword_1803BE468, v432);
                            v151 = 358;
                            goto LABEL_822;
                          }
                        }
                        sub_18000CE30(a1, &v1438);
                        v10 = v1440;
                        v151 = 344;
                        goto LABEL_2516;
                      }
                    }
                    sub_18000CE30(v376, &v1438);
                    v380 = v1438;
                    sub_18000CE70(&v1438);
                    v1438 = v380;
                    v10 = v1440;
                    v151 = 340;
                    goto LABEL_2516;
                  }
// Original IDA line 4348 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                  v467 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
                  v468 = (int *)v467;
                  if ( v467 == 0 )
                  {
                    sub_18000CE30(a1, &v1438);
                    v151 = 361;
                    v10 = v275;
                    goto LABEL_2516;
                  }
// Original IDA line 4357 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                  v469 = (int *)sub_18037F0C0(a1, v467, qword_1803BE0B8);
                  if ( *v468 >= 0 )
                  {
                    v8 = (*(_QWORD *)v468)-- == 1;
                    if ( v8 )
                    {
                      v470 = sub_18000CC40(v468);
                      (*(void (__fastcall **)(int *))(v470 + 48))(v468);
                    }
                  }
                  if ( v469 == nullptr )
                  {
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 361;
                    goto LABEL_2516;
                  }
// Original IDA line 4374 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                  v1453 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                  v471 = v1453;
                  if ( v1453 == nullptr )
                  {
                    sub_18000CE30(a1, &v1438);
                    if ( *v469 >= 0 )
                    {
                      v8 = (*(_QWORD *)v469)-- == 1;
                      if ( v8 )
                      {
                        v472 = sub_18000CC40(v469);
                        (*(void (__fastcall **)(int *))(v472 + 48))(v469);
                      }
                    }
                    v10 = v1440;
                    v151 = 361;
                    goto LABEL_2516;
                  }
// Original IDA line 4392 | qword_1803BE478 -> novels.services.agent_session#173 bytes [4309683, 4309694]: "短篇《"
                  v473 = (_DWORD *)qword_1803BE478;
// Original IDA line 4393 | qword_1803BE470 -> novels.services.agent_session#172 bytes [4309664, 4309683]: "short_fiction_run"
                  v1442 = (_QWORD *)qword_1803BE470;
                  v474 = sub_180374A40(a1, 3);
                  v177 = (int *)v474;
                  if ( *v473 != -1 )
                    ++*v473;
                  sub_18000CC80(v474, 0, v473);
// Original IDA line 4399 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                  v475 = (int *)sub_18037F0C0(a1, v345, qword_1803BDFD8);
                  v151 = 363;
                  if ( v475 != nullptr )
                  {
                    *(_DWORD *)(v1440 + 40) = 363;
// Original IDA line 4404 | qword_1803BE480 -> novels.services.agent_session#174 bytes [4309694, 4309704]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309696, 4309703], "value": "title"}, {"tag": "s", "fileRange": [4309703, 4309704], "value": ""}]}
                    v476 = (int *)sub_180378210(a1, v475, qword_1803BE480);
                    if ( *v475 >= 0 )
                    {
                      v8 = (*(_QWORD *)v475)-- == 1;
                      if ( v8 )
                      {
                        v477 = sub_18000CC40(v475);
                        (*(void (__fastcall **)(int *))(v477 + 48))(v475);
                      }
                    }
                    if ( v476 != nullptr )
                    {
// Original IDA line 4416 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                      v478 = sub_180375520(a1, v476, qword_1803BDFE8);
                      if ( *v476 >= 0 )
                      {
                        v8 = (*(_QWORD *)v476)-- == 1;
                        if ( v8 )
                        {
                          v479 = sub_18000CC40(v476);
                          (*(void (__fastcall **)(int *))(v479 + 48))(v476);
                        }
                      }
                      if ( v478 != 0 )
                      {
                        sub_18000CC80(v177, 1, v478);
// Original IDA line 4429 | qword_1803BE488 -> novels.services.agent_session#175 bytes [4309704, 4309718]: "》已生成"
                        v481 = qword_1803BE488;
// Original IDA line 4430 | qword_1803BE488 -> novels.services.agent_session#175 bytes [4309704, 4309718]: "》已生成"
                        if ( *(_DWORD *)qword_1803BE488 != -1 )
// Original IDA line 4431 | qword_1803BE488 -> novels.services.agent_session#175 bytes [4309704, 4309718]: "》已生成"
                          ++*(_DWORD *)qword_1803BE488;
                        sub_18000CC80(v177, 2, v481);
// Original IDA line 4433 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                        v482 = PyUnicode_Join(qword_1803BDFE8, v177);
                        v29 = *v177 < 0;
                        v483 = v482;
                        v1448 = (int *)v482;
                        if ( !v29 )
                        {
                          v8 = (*(_QWORD *)v177)-- == 1;
                          if ( v8 )
                          {
                            v484 = sub_18000CC40(v177);
                            (*(void (__fastcall **)(int *))(v484 + 48))(v177);
                          }
                        }
                        if ( v483 == 0 )
                        {
                          sub_18000CE30(a1, &v1438);
                          if ( *v469 >= 0 )
                          {
                            v8 = (*(_QWORD *)v469)-- == 1;
                            if ( v8 )
                            {
                              v485 = sub_18000CC40(v469);
                              v486 = v469;
LABEL_894:
                              (*(void (__fastcall **)(int *))(v485 + 48))(v486);
                              goto LABEL_895;
                            }
                          }
                          goto LABEL_895;
                        }
                        v488 = (int *)sub_180001930(a1, v1449);
                        if ( v488 == nullptr )
                        {
                          sub_18000CE30(a1, &v1438);
                          v489 = v1438;
                          sub_18000CE70(&v1438);
                          v1438 = v489;
                          v10 = v1440;
                          v151 = 365;
                          goto LABEL_2516;
                        }
                        v490 = (int *)sub_180372160(a1, 0);
                        for ( j = (_DWORD *)sub_18000D430(v488); j != nullptr; j = (_DWORD *)sub_18000D430(v488) )
                        {
                          v492 = (int *)v1450;
                          v493 = v1458;
                          v494 = v1450;
                          v1450 = j;
                          v495 = v1458;
                          if ( v494 != nullptr && *v492 >= 0 )
                          {
                            v8 = (*(_QWORD *)v492)-- == 1;
                            if ( v8 )
                            {
                              v496 = sub_18000CC40(v492);
                              (*(void (__fastcall **)(_QWORD *))(v496 + 48))(v494);
                              v493 = v495;
                            }
                          }
                          v497 = *j;
                          v498 = v493;
                          v1458 = j;
                          v499 = v497 + 1;
                          if ( v499 != 0 )
                            *j = v499;
                          if ( v493 != nullptr && *v495 >= 0 )
                          {
                            v8 = (*(_QWORD *)v495)-- == 1;
                            if ( v8 )
                            {
                              v500 = sub_18000CC40(v493);
                              (*(void (__fastcall **)(int *))(v500 + 48))(v498);
                            }
                          }
// Original IDA line 4507 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                          v501 = qword_1803BE378;
                          v502 = PyDict_NewPresized(2);
                          PyDict_SetItem(v502, v501, j);
// Original IDA line 4510 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"; qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                          PyDict_SetItem(v502, qword_1803BE380, qword_1803BE2E0);
                          if ( (unsigned __int8)sub_180372560(v490, v502) == 0
                            || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                          {
                            goto LABEL_974;
                          }
                        }
                        if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
                        {
LABEL_974:
                          sub_18000CE30(a1, &v1438);
                          v536 = v1438;
                          sub_18000CE70(&v1438);
                          if ( *v488 >= 0 )
                          {
                            v8 = (*(_QWORD *)v488)-- == 1;
                            if ( v8 )
                            {
                              v537 = sub_18000CC40(v488);
                              (*(void (__fastcall **)(int *))(v537 + 48))(v488);
                            }
                          }
                          if ( *v490 >= 0 )
                          {
                            v8 = (*(_QWORD *)v490)-- == 1;
                            if ( v8 )
                            {
                              v538 = sub_18000CC40(v490);
                              (*(void (__fastcall **)(int *))(v538 + 48))(v490);
                            }
                          }
                          v539 = v1458;
                          if ( v1458 != nullptr && *v1458 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1458)-- == 1;
                            v1450 = *(_QWORD **)v539;
                            if ( v8 )
                            {
                              v540 = sub_18000CC40(v539);
                              (*(void (__fastcall **)(int *))(v540 + 48))(v539);
                            }
                          }
                          v1438 = v536;
                          sub_18000CE70(&v1438);
                          if ( v539 != nullptr && *v539 >= 0 )
                          {
                            v8 = (*(_QWORD *)v539)-- == 1;
                            v1450 = *(_QWORD **)v539;
                            if ( v8 )
                            {
                              v541 = sub_18000CC40(v539);
                              (*(void (__fastcall **)(int *))(v541 + 48))(v539);
                            }
                          }
                          v1438 = v536;
                          v10 = v1440;
                          v151 = 365;
                          goto LABEL_2516;
                        }
                        if ( *v490 != -1 )
                          ++*v490;
                        if ( *v488 >= 0 )
                        {
                          v8 = (*(_QWORD *)v488)-- == 1;
                          if ( v8 )
                          {
                            v503 = sub_18000CC40(v488);
                            (*(void (__fastcall **)(int *))(v503 + 48))(v488);
                          }
                        }
                        if ( *v490 >= 0 )
                        {
                          v8 = (*(_QWORD *)v490)-- == 1;
                          if ( v8 )
                          {
                            v504 = sub_18000CC40(v490);
                            (*(void (__fastcall **)(int *))(v504 + 48))(v490);
                          }
                        }
                        v505 = v1458;
                        if ( v1458 != nullptr && *v1458 >= 0 )
                        {
                          v8 = (*(_QWORD *)v1458)-- == 1;
                          v506 = *(_QWORD **)v505;
                          if ( v8 )
                          {
                            v507 = sub_18000CC40(v505);
                            (*(void (__fastcall **)(int *))(v507 + 48))(v505);
                            LODWORD(v506) = *v505;
                          }
                          if ( (int)v506 >= 0 )
                          {
                            v8 = (*(_QWORD *)v505)-- == 1;
                            v1450 = *(_QWORD **)v505;
                            if ( v8 )
                            {
                              v508 = sub_18000CC40(v505);
                              (*(void (__fastcall **)(int *))(v508 + 48))(v505);
                            }
                          }
                        }
                        v509 = v1453;
                        v510 = v1448;
                        v151 = 361;
                        v511 = v1453;
                        *(_DWORD *)(v1440 + 40) = 361;
                        v1512[0] = v1442;
                        v1511[0] = v1454;
                        v1512[1] = v510;
                        v1511[1] = v490;
// Original IDA line 4620 | qword_1803BE490 -> novels.services.agent_session#176 bytes [4309718, 4309737]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309720, 4309729], "value": "details"}, {"tag": "a", "fileRange": [4309729, 4309737], "value": "stages"}]}
                        v512 = (int *)sub_18037D650(a1, v511, v1512, v1511, qword_1803BE490);
                        if ( *v509 >= 0 )
                        {
                          v8 = (*(_QWORD *)v509)-- == 1;
                          if ( v8 )
                          {
                            v513 = sub_18000CC40(v509);
                            (*(void (__fastcall **)(int *))(v513 + 48))(v509);
                          }
                        }
                        if ( *v510 >= 0 )
                        {
                          v8 = (*(_QWORD *)v510)-- == 1;
                          if ( v8 )
                          {
                            v514 = sub_18000CC40(v510);
                            (*(void (__fastcall **)(int *))(v514 + 48))(v510);
                          }
                        }
                        if ( *v490 >= 0 )
                        {
                          v8 = (*(_QWORD *)v490)-- == 1;
                          if ( v8 )
                          {
                            v515 = sub_18000CC40(v490);
                            (*(void (__fastcall **)(int *))(v515 + 48))(v490);
                          }
                        }
                        if ( v512 == nullptr )
                        {
                          sub_18000CE30(a1, &v1438);
                          goto LABEL_941;
                        }
                        v10 = v1440;
                        *(_DWORD *)(v1440 + 40) = 361;
                        v517 = (int *)sub_180376770(a1, v469, v512);
                        if ( *v469 >= 0 )
                        {
                          v8 = (*(_QWORD *)v469)-- == 1;
                          if ( v8 )
                          {
                            v518 = sub_18000CC40(v469);
                            (*(void (__fastcall **)(int *))(v518 + 48))(v469);
                          }
                        }
                        if ( *v512 >= 0 )
                        {
                          v8 = (*(_QWORD *)v512)-- == 1;
                          if ( v8 )
                          {
                            v519 = sub_18000CC40(v512);
                            (*(void (__fastcall **)(int *))(v519 + 48))(v512);
                          }
                        }
                        if ( v517 == nullptr )
                          goto LABEL_2475;
                        if ( *v517 >= 0 )
                        {
                          v8 = (*(_QWORD *)v517)-- == 1;
                          if ( v8 )
                          {
                            v520 = sub_18000CC40(v517);
                            (*(void (__fastcall **)(int *))(v520 + 48))(v517);
                          }
                        }
// Original IDA line 4685 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                        v521 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                        v522 = (int *)v521;
                        if ( v521 == 0 )
                        {
                          sub_18000CE30(a1, &v1438);
                          v151 = 370;
                          goto LABEL_2516;
                        }
                        v523 = v1445;
                        v151 = 370;
                        v524 = *(_QWORD *)(v1445 + 16);
                        *(_DWORD *)(v10 + 40) = 370;
// Original IDA line 4697 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                        v525 = (int *)sub_18037E6C0(a1, v521, qword_1803BE0D8, v524);
                        if ( *v522 >= 0 )
                        {
                          v8 = (*(_QWORD *)v522)-- == 1;
                          if ( v8 )
                          {
                            v526 = sub_18000CC40(v522);
                            (*(void (__fastcall **)(int *))(v526 + 48))(v522);
                          }
                        }
                        if ( v525 == nullptr )
                          goto LABEL_2475;
                        if ( *v525 >= 0 )
                        {
                          v8 = (*(_QWORD *)v525)-- == 1;
                          if ( v8 )
                          {
                            v527 = sub_18000CC40(v525);
                            (*(void (__fastcall **)(int *))(v527 + 48))(v525);
                          }
                        }
// Original IDA line 4718 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                        v528 = qword_1803BE2C8;
                        v23 = (int *)PyDict_NewPresized(2);
                        PyDict_SetItem(v23, v528, Py_TrueStruct);
// Original IDA line 4721 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                        v248 = qword_1803BE2F8;
// Original IDA line 4722 | qword_1803BE468 -> novels.services.agent_session#171 bytes [4309650, 4309664]: "shortFiction"
                        v529 = qword_1803BE468;
                        v252 = (int *)PyDict_NewPresized(2);
                        PyDict_SetItem(v252, v529, v1454);
                        v151 = 375;
// Original IDA line 4726 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
                        v530 = qword_1803BE1B8;
                        v531 = *(_QWORD *)(v523 + 16);
                        *(_DWORD *)(v1440 + 40) = 375;
// Original IDA line 4729 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
                        v532 = sub_18037E560(a1, v531, qword_1803BE278);
                        v533 = (int *)v532;
                        if ( v532 != 0 )
                        {
                          PyDict_SetItem(v252, v530, v532);
                          if ( *v533 >= 0 )
                          {
                            v8 = (*(_QWORD *)v533)-- == 1;
                            if ( v8 )
                            {
                              v534 = sub_18000CC40(v533);
                              (*(void (__fastcall **)(int *))(v534 + 48))(v533);
                            }
                          }
                          goto LABEL_971;
                        }
LABEL_965:
                        sub_18000CE30(a1, &v1438);
                        if ( *v252 >= 0 )
                        {
                          v8 = (*(_QWORD *)v252)-- == 1;
                          if ( v8 )
                          {
                            v454 = sub_18000CC40(v252);
                            v455 = v252;
                            goto LABEL_827;
                          }
                        }
LABEL_828:
                        if ( *v23 < 0 )
                          goto LABEL_272;
                        v8 = (*(_QWORD *)v23)-- == 1;
                        if ( !v8 )
                          goto LABEL_272;
                        goto LABEL_415;
                      }
                    }
                  }
                  sub_18000CE30(a1, &v1438);
                  if ( *v469 >= 0 )
                  {
                    v8 = (*(_QWORD *)v469)-- == 1;
                    if ( v8 )
                    {
                      v480 = sub_18000CC40(v469);
                      (*(void (__fastcall **)(int *))(v480 + 48))(v469);
                    }
                  }
                  if ( *v471 >= 0 )
                  {
                    v8 = (*(_QWORD *)v471)-- == 1;
                    if ( v8 )
                    {
                      v183 = sub_18000CC40(v471);
                      v184 = v471;
                      goto LABEL_315;
                    }
                  }
LABEL_316:
                  if ( *v177 < 0 )
                    goto LABEL_272;
                  v8 = (*(_QWORD *)v177)-- == 1;
                  if ( !v8 )
                    goto LABEL_272;
                  goto LABEL_318;
                }
              }
            }
LABEL_495:
            sub_18000CE30(a1, &v1438);
            if ( *v284 >= 0 )
            {
              v8 = (*(_QWORD *)v284)-- == 1;
              if ( v8 )
              {
                v287 = sub_18000CC40(v284);
                (*(void (__fastcall **)(int *))(v287 + 48))(v284);
              }
            }
            v10 = v275;
            goto LABEL_2516;
          }
          if ( *v23 >= 0 )
          {
            v8 = (*(_QWORD *)v23)-- == 1;
            if ( v8 )
            {
              v322 = sub_18000CC40(v23);
              (*(void (__fastcall **)(int *))(v322 + 48))(v23);
            }
          }
          v323 = sub_180001510(v308);
          if ( v323 != -1 )
          {
// Original IDA line 4823 | qword_1803BE400 -> novels.services.agent_session#158 bytes [4309402, 4309410]: "短篇"
            v23 = (int *)qword_1803BE400;
            if ( v323 == 1 )
              v23 = (int *)v308;
            if ( *v23 != -1 )
              ++*v23;
            goto LABEL_597;
          }
LABEL_565:
          v314 = a1;
LABEL_566:
          sub_18000CE30(v314, &v1438);
          goto LABEL_567;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 319;
      v10 = v275;
      goto LABEL_2516;
    }
// Original IDA line 4842 | qword_1803BE498 -> novels.services.agent_session#177 bytes [4309737, 4309749]: "play_start"
    v542 = sub_1803836D0(v45, qword_1803BE498);
    if ( v542 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 379;
      goto LABEL_2516;
    }
    if ( v542 == 1 )
    {
// Original IDA line 4851 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
      v543 = sub_18037F0C0(a1, *(_QWORD *)(v148 + 16), qword_1803BDF18);
      v544 = (int *)v543;
      if ( v543 != 0 )
      {
// Original IDA line 4855 | qword_1803BE4A0 -> novels.services.agent_session#178 bytes [4309749, 4309761]: "start_play"
        v545 = (int *)sub_18037F0C0(a1, v543, qword_1803BE4A0);
        if ( *v544 >= 0 )
        {
          v8 = (*(_QWORD *)v544)-- == 1;
          if ( v8 )
          {
            v546 = sub_18000CC40(v544);
            (*(void (__fastcall **)(int *))(v546 + 48))(v544);
          }
        }
        if ( v545 != nullptr )
        {
          v151 = 380;
          *(_DWORD *)(v10 + 40) = 380;
          v1496 = v5;
          v1495 = v1456;
// Original IDA line 4871 | qword_1803BE340 -> novels.services.agent_session#134 bytes [4309052, 4309061]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309054, 4309061], "value": "model"}]}
          v547 = sub_18037D030(a1, v545, &v1496, &v1495, qword_1803BE340);
          v29 = *v545 < 0;
          v548 = (int *)v547;
          v1449 = (int *)v547;
          if ( !v29 )
          {
            v8 = (*(_QWORD *)v545)-- == 1;
            if ( v8 )
            {
              v549 = sub_18000CC40(v545);
              (*(void (__fastcall **)(int *))(v549 + 48))(v545);
            }
          }
          if ( v548 == nullptr )
            goto LABEL_2475;
          v550 = *(_QWORD *)(v149 + 16);
          v1457 = v548;
// Original IDA line 4888 | qword_1803BDFA0 -> novels.services.agent_session#18 bytes [4307199, 4307205]: "kind"; qword_1803BDFA8 -> novels.services.agent_session#19 bytes [4307205, 4307211]: "play"
          if ( (unsigned __int8)sub_18037F3A0(a1, v550, qword_1803BDFA0, qword_1803BDFA8) == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 381;
            goto LABEL_2516;
          }
// Original IDA line 4894 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v551 = sub_18037F0C0(a1, v548, qword_1803BDFD8);
          v552 = (int *)v551;
          if ( v551 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 382;
            goto LABEL_2516;
          }
          v151 = 382;
          *(_DWORD *)(v10 + 40) = 382;
// Original IDA line 4904 | qword_1803BE4A8 -> novels.services.agent_session#179 bytes [4309761, 4309772]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309763, 4309772], "value": "worldId"}]}
          v177 = (int *)sub_1803770D0(a1, v551, qword_1803BE4A8);
          if ( *v552 >= 0 )
          {
            v8 = (*(_QWORD *)v552)-- == 1;
            if ( v8 )
            {
              v553 = sub_18000CC40(v552);
              (*(void (__fastcall **)(int *))(v553 + 48))(v552);
            }
          }
          if ( v177 == nullptr )
            goto LABEL_2475;
          v554 = sub_180001510(v177);
          if ( v554 == -1 )
          {
            v435 = a1;
            goto LABEL_798;
          }
          if ( v554 != 1 )
          {
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v555 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v555 + 48))(v177);
              }
            }
// Original IDA line 4933 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
            v177 = (int *)qword_1803BDFE8;
// Original IDA line 4934 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
            if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 4935 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              ++*(_DWORD *)qword_1803BDFE8;
          }
          v556 = (int *)sub_1803749A0(v177);
          if ( *v177 >= 0 )
          {
            v8 = (*(_QWORD *)v177)-- == 1;
            if ( v8 )
            {
              v557 = sub_18000CC40(v177);
              (*(void (__fastcall **)(int *))(v557 + 48))(v177);
            }
          }
          if ( v556 == nullptr )
            goto LABEL_2475;
          v558 = v1445;
// Original IDA line 4950 | qword_1803BDFB0 -> novels.services.agent_session#20 bytes [4307211, 4307220]: "worldId"
          v559 = sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFB0, v556);
          if ( *v556 >= 0 )
          {
            v8 = (*(_QWORD *)v556)-- == 1;
            if ( v8 )
            {
              v560 = sub_18000CC40(v556);
              (*(void (__fastcall **)(int *))(v560 + 48))(v556);
            }
          }
          if ( v559 == 0 )
            goto LABEL_2475;
// Original IDA line 4962 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v561 = sub_18037F0C0(a1, v548, qword_1803BDFD8);
          v562 = (int *)v561;
          if ( v561 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 383;
            goto LABEL_2516;
          }
          v151 = 383;
          *(_DWORD *)(v10 + 40) = 383;
// Original IDA line 4972 | qword_1803BE3E8 -> novels.services.agent_session#155 bytes [4309373, 4309382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309375, 4309382], "value": "title"}]}
          v177 = (int *)sub_1803770D0(a1, v561, qword_1803BE3E8);
          if ( *v562 >= 0 )
          {
            v8 = (*(_QWORD *)v562)-- == 1;
            if ( v8 )
            {
              v563 = sub_18000CC40(v562);
              (*(void (__fastcall **)(int *))(v563 + 48))(v562);
            }
          }
          if ( v177 == nullptr )
            goto LABEL_2475;
          v564 = sub_180001510(v177);
          if ( v564 == -1 )
          {
            v435 = a1;
            goto LABEL_798;
          }
          if ( v564 != 1 )
          {
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v565 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v565 + 48))(v177);
              }
            }
// Original IDA line 5001 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
            v177 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v558 + 16), qword_1803BE350);
            if ( v177 == nullptr )
              goto LABEL_2475;
          }
          v566 = (int *)sub_1803749A0(v177);
          if ( *v177 >= 0 )
          {
            v8 = (*(_QWORD *)v177)-- == 1;
            if ( v8 )
            {
              v567 = sub_18000CC40(v177);
              (*(void (__fastcall **)(int *))(v567 + 48))(v177);
            }
          }
          if ( v566 == nullptr )
            goto LABEL_2475;
// Original IDA line 5017 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
          v568 = sub_18037F3A0(a1, *(_QWORD *)(v558 + 16), qword_1803BE350, v566);
          if ( *v566 >= 0 )
          {
            v8 = (*(_QWORD *)v566)-- == 1;
            if ( v8 )
            {
              v569 = sub_18000CC40(v566);
              (*(void (__fastcall **)(int *))(v569 + 48))(v566);
            }
          }
          if ( v568 == 0 )
            goto LABEL_2475;
// Original IDA line 5029 | qword_1803BE0D0 -> novels.services.agent_session#56 bytes [4307684, 4307693]: "isDraft"
          if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v558 + 16), qword_1803BE0D0, Py_FalseStruct) == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 384;
            goto LABEL_2516;
          }
// Original IDA line 5035 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v570 = sub_18037F0C0(a1, v548, qword_1803BDFD8);
          v571 = (int *)v570;
          if ( v570 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 385;
            goto LABEL_2516;
          }
          v151 = 385;
          *(_DWORD *)(v10 + 40) = 385;
// Original IDA line 5045 | qword_1803BE4B0 -> novels.services.agent_session#180 bytes [4309772, 4309788]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309774, 4309788], "value": "openingScene"}]}
          v177 = (int *)sub_1803770D0(a1, v570, qword_1803BE4B0);
          if ( *v571 >= 0 )
          {
            v8 = (*(_QWORD *)v571)-- == 1;
            if ( v8 )
            {
              v572 = sub_18000CC40(v571);
              (*(void (__fastcall **)(int *))(v572 + 48))(v571);
            }
          }
          if ( v177 == nullptr )
            goto LABEL_2475;
          v573 = sub_180001510(v177);
          if ( v573 == -1 )
          {
            v435 = a1;
            goto LABEL_798;
          }
          if ( v573 != 1 )
          {
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v574 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v574 + 48))(v177);
              }
            }
// Original IDA line 5074 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
            v177 = (int *)qword_1803BDFE8;
// Original IDA line 5075 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
            if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 5076 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              ++*(_DWORD *)qword_1803BDFE8;
          }
          v575 = (int *)sub_1803749A0(v177);
          if ( *v177 >= 0 )
          {
            v8 = (*(_QWORD *)v177)-- == 1;
            if ( v8 )
            {
              v576 = sub_18000CC40(v177);
              (*(void (__fastcall **)(int *))(v576 + 48))(v177);
            }
          }
          if ( v575 == nullptr )
            goto LABEL_2475;
          v1483 = v575;
// Original IDA line 5091 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v577 = sub_18037F0C0(a1, v548, qword_1803BDFD8);
          v578 = (int *)v577;
          if ( v577 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 386;
            goto LABEL_2516;
          }
          v151 = 386;
          *(_DWORD *)(v10 + 40) = 386;
// Original IDA line 5101 | qword_1803BE4B8 -> novels.services.agent_session#181 bytes [4309788, 4309808]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309790, 4309808], "value": "suggestedActions"}]}
          v177 = (int *)sub_1803770D0(a1, v577, qword_1803BE4B8);
          if ( *v578 >= 0 )
          {
            v8 = (*(_QWORD *)v578)-- == 1;
            if ( v8 )
            {
              v579 = sub_18000CC40(v578);
              (*(void (__fastcall **)(int *))(v579 + 48))(v578);
            }
          }
          if ( v177 == nullptr )
            goto LABEL_2475;
          v580 = sub_180001510(v177);
          if ( v580 == -1 )
          {
            v435 = a1;
            goto LABEL_798;
          }
          if ( v580 == 1 )
          {
            v581 = (__int64)v177;
          }
          else
          {
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v582 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v582 + 48))(v177);
              }
            }
            v581 = sub_180372160(a1, 0);
          }
          v1467 = (int *)v581;
          if ( *v575 != -1 )
            ++*v575;
          v1462 = v575;
          v583 = sub_180001510(v581);
          if ( v583 == -1 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 388;
            goto LABEL_2516;
          }
          if ( v583 != 0 )
          {
// Original IDA line 5149 | qword_1803BE4C8 -> novels.services.agent_session#183 bytes [4309827, 4309832]: "、"
            v584 = qword_1803BE4C8;
            v1487 = v575;
// Original IDA line 5151 | qword_1803BE4C0 -> novels.services.agent_session#182 bytes [4309808, 4309827]: "\n\n可选行动："
            v585 = qword_1803BE4C0;
            v586 = sub_180001930(a1, v581);
            v587 = (int *)v586;
            if ( v586 == 0 )
              goto LABEL_1093;
            v1497 = sub_18036D750(v586);
            v588 = (int *)sub_180209550(a1, &v1497);
            if ( *v587 >= 0 )
            {
              v8 = (*(_QWORD *)v587)-- == 1;
              if ( v8 )
              {
                v589 = sub_18000CC40(v587);
                (*(void (__fastcall **)(int *))(v589 + 48))(v587);
              }
            }
            v590 = (int *)sub_180374990(a1, v584, v588);
            if ( *v588 >= 0 )
            {
              v8 = (*(_QWORD *)v588)-- == 1;
              if ( v8 )
              {
                v591 = sub_18000CC40(v588);
                (*(void (__fastcall **)(int *))(v591 + 48))(v588);
              }
            }
            if ( v590 == nullptr )
              goto LABEL_1093;
            v592 = (int *)sub_180380E40(v585, v590);
            if ( *v590 >= 0 )
            {
              v8 = (*(_QWORD *)v590)-- == 1;
              if ( v8 )
              {
                v593 = sub_18000CC40(v590);
                (*(void (__fastcall **)(int *))(v593 + 48))(v590);
              }
            }
            v594 = sub_180382E90(&v1487, v592);
            if ( *v592 >= 0 )
            {
              v8 = (*(_QWORD *)v592)-- == 1;
              if ( v8 )
              {
                v595 = sub_18000CC40(v592);
                (*(void (__fastcall **)(int *))(v595 + 48))(v592);
              }
            }
            if ( v594 == 0 )
            {
LABEL_1093:
              sub_18000CE30(a1, &v1438);
              v151 = 389;
              goto LABEL_2516;
            }
            v1462 = v1487;
          }
// Original IDA line 5208 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
          if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 390;
            goto LABEL_2516;
          }
// Original IDA line 5214 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
          v596 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
          v597 = (int *)v596;
          if ( v596 != 0 )
          {
// Original IDA line 5218 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
            v177 = (int *)sub_18037F0C0(a1, v596, qword_1803BE0B8);
            if ( *v597 >= 0 )
            {
              v8 = (*(_QWORD *)v597)-- == 1;
              if ( v8 )
              {
                v598 = sub_18000CC40(v597);
                (*(void (__fastcall **)(int *))(v598 + 48))(v597);
              }
            }
            if ( v177 != nullptr )
            {
// Original IDA line 5230 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
              v23 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
              if ( v23 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v177 >= 0 )
                {
                  v8 = (*(_QWORD *)v177)-- == 1;
                  if ( v8 )
                  {
                    v599 = sub_18000CC40(v177);
                    (*(void (__fastcall **)(int *))(v599 + 48))(v177);
                  }
                }
                v151 = 391;
                goto LABEL_2516;
              }
// Original IDA line 5246 | qword_1803BE4D0 -> novels.services.agent_session#184 bytes [4309832, 4309849]: "互动世界《"
              v600 = (_DWORD *)qword_1803BE4D0;
// Original IDA line 5247 | qword_1803BE498 -> novels.services.agent_session#177 bytes [4309737, 4309749]: "play_start"
              v1442 = (_QWORD *)qword_1803BE498;
              v601 = sub_180374A40(a1, 3);
              v602 = (int *)v601;
              if ( *v600 != -1 )
                ++*v600;
              sub_18000CC80(v601, 0, v600);
// Original IDA line 5253 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
              v603 = sub_18037F0C0(a1, v548, qword_1803BDFD8);
              v604 = (int *)v603;
              v151 = 393;
              if ( v603 == 0 )
              {
LABEL_1134:
                sub_18000CE30(a1, &v1438);
                if ( *v177 >= 0 )
                {
                  v8 = (*(_QWORD *)v177)-- == 1;
                  if ( v8 )
                  {
                    v609 = sub_18000CC40(v177);
                    (*(void (__fastcall **)(int *))(v609 + 48))(v177);
                  }
                }
                if ( *v23 >= 0 )
                {
                  v8 = (*(_QWORD *)v23)-- == 1;
                  if ( v8 )
                  {
                    v610 = sub_18000CC40(v23);
                    (*(void (__fastcall **)(int *))(v610 + 48))(v23);
                  }
                }
                if ( *v602 >= 0 )
                {
                  v8 = (*(_QWORD *)v602)-- == 1;
                  if ( v8 )
                  {
                    v436 = sub_18000CC40(v602);
                    v437 = v602;
LABEL_2515:
                    (*(void (__fastcall **)(int *))(v436 + 48))(v437);
                  }
                }
LABEL_2516:
                v1471 = v1438;
                sub_18000CE70(&v1438);
                v1507 = sub_18000CCC0(a1);
                v1314 = sub_18000CE80(&v1471);
                if ( v1314 != 0 )
                  v1315 = sub_18000CC90(v1314, v10, v151);
                else
                  v1315 = sub_18038E2A0(v10, v151);
                sub_180001300(&v1471, v1315);
                sub_18000D050(a1, &v1471);
                v1316 = sub_18000CC40(*(_QWORD *)(a1 + 248));
                if ( (unsigned int)sub_18000CE90(a1, v1316, PyExc_Exception) == 0 )
                {
                  v14 = 288;
                  if ( (unsigned __int8)sub_180373290(a1, &v1438) != 0 )
                    v14 = 0;
                  v1369 = sub_18000CE80(&v1438);
                  if ( v1369 != 0 && *(_QWORD *)(v1369 + 24) == v10 )
                    *(_DWORD *)(v10 + 40) = *(_DWORD *)(v1369 + 36);
                  v1367 = v1438;
                  goto LABEL_2627;
                }
                v1317 = *(int **)(a1 + 248);
                v1447 = v1317;
                if ( *v1317 != -1 )
                  ++*v1317;
                v14 = 564;
// Original IDA line 5317 | qword_1803BE6B8 -> novels.services.agent_session#245 bytes [4310848, 4310865]: "utils.llm_retry"
                v1318 = qword_1803BE6B8;
                v1319 = qword_1803CA5B8;
// Original IDA line 5319 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
                v1437 = qword_1803BE028;
// Original IDA line 5320 | qword_1803BE6C0 -> novels.services.agent_session#246 bytes [4310865, 4310911]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310867, 4310887], "value": "friendly_llm_error"}, {"tag": "a", "fileRange": [4310887, 4310911], "value": "is_transient_llm_error"}]}
                v1434 = qword_1803BE6C0;
                *(_DWORD *)(v10 + 40) = 564;
                v1320 = sub_1803717E0(a1, v1318, v1319, Py_NoneStruct, v1434, v1437);
                v1321 = (int *)v1320;
                if ( v1320 != 0 )
                {
                  if ( (unsigned int)sub_18000CC50(v1320, PyModule_Type) != 0 )
// Original IDA line 5327 | qword_1803BE6C8 -> novels.services.agent_session#247 bytes [4310911, 4310931]: "friendly_llm_error"
                    v1322 = sub_180371D80(a1, (_DWORD)v1321, qword_1803CA5B8, qword_1803BE6C8, qword_1803AB250);
                  else
// Original IDA line 5329 | qword_1803BE6C8 -> novels.services.agent_session#247 bytes [4310911, 4310931]: "friendly_llm_error"
                    v1322 = sub_180371890(a1, v1321, qword_1803BE6C8);
                  v1323 = v1322;
                  if ( v1322 == 0
                    || ((v1484 = (int *)v1322, (unsigned int)sub_18000CC50(v1321, PyModule_Type) == 0)
// Original IDA line 5333 | qword_1803BE6D0 -> novels.services.agent_session#248 bytes [4310931, 4310955]: "is_transient_llm_error"
                      ? (v1324 = sub_180371890(a1, v1321, qword_1803BE6D0))
// Original IDA line 5334 | qword_1803BE6D0 -> novels.services.agent_session#248 bytes [4310931, 4310955]: "is_transient_llm_error"
                      : (v1324 = sub_180371D80(a1, (_DWORD)v1321, qword_1803CA5B8, qword_1803BE6D0, qword_1803AB250)),
                        v1325 = v1324,
                        v1324 == 0) )
                  {
                    sub_18000CE30(a1, &v1438);
                    v1326 = v1438;
                    sub_18000CE70(&v1438);
                    if ( *v1321 >= 0 )
                    {
                      v8 = (*(_QWORD *)v1321)-- == 1;
                      if ( v8 )
                      {
                        v1327 = sub_18000CC40(v1321);
                        (*(void (__fastcall **)(int *))(v1327 + 48))(v1321);
                      }
                    }
                    v1438 = v1326;
                    goto LABEL_2616;
                  }
                  v1485 = (int *)v1324;
                  if ( *v1321 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1321)-- == 1;
                    if ( v8 )
                    {
                      v1328 = sub_18000CC40(v1321);
                      (*(void (__fastcall **)(int *))(v1328 + 48))(v1321);
                    }
                  }
                  v14 = 565;
                  *(_DWORD *)(v10 + 40) = 565;
                  v1329 = sub_180376770(a1, v1325, v1317);
                  v1330 = (int *)v1329;
                  if ( v1329 != 0 )
                  {
                    v1331 = sub_180001510(v1329);
                    if ( v1331 == -1 )
                    {
                      sub_18000CE30(a1, &v1438);
                      if ( *v1330 >= 0 )
                      {
                        v8 = (*(_QWORD *)v1330)-- == 1;
                        if ( v8 )
                        {
                          v1332 = sub_18000CC40(v1330);
                          (*(void (__fastcall **)(int *))(v1332 + 48))(v1330);
                        }
                      }
                      goto LABEL_2616;
                    }
                    if ( *v1330 >= 0 )
                    {
                      v8 = (*(_QWORD *)v1330)-- == 1;
                      if ( v8 )
                      {
                        v1333 = sub_18000CC40(v1330);
                        (*(void (__fastcall **)(int *))(v1333 + 48))(v1330);
                      }
                    }
                    if ( v1331 != 0 )
                    {
                      *(_DWORD *)(v10 + 40) = 565;
                      v1334 = sub_180376770(a1, v1323, v1317);
                    }
                    else
                    {
                      v1334 = sub_1803749A0(v1317);
                    }
                    v1335 = v1334;
                    if ( v1334 != 0 )
                    {
                      v1336 = v1463;
                      v1486 = (int *)v1334;
                      v1337 = sub_180001510(v1463);
                      if ( v1337 == -1 )
                      {
                        sub_18000CE30(a1, &v1438);
                        v14 = 566;
                        goto LABEL_2615;
                      }
// Original IDA line 5414 | qword_1803BE6D8 -> novels.services.agent_session#249 bytes [4310955, 4310961]: "task"
                      v1338 = qword_1803BE6D8;
                      if ( v1337 == 1 )
                        v1338 = v1336;
                      v1339 = v1466;
                      v14 = 566;
                      *(_DWORD *)(v10 + 40) = 566;
                      v1527[0] = v1338;
                      v1527[1] = v1447;
                      v1340 = (int *)sub_180377750(a1, v1339, v1527);
                      v1341 = v1340;
                      if ( v1340 != nullptr )
                      {
                        if ( *v1340 >= 0 )
                        {
                          v8 = (*(_QWORD *)v1340)-- == 1;
                          if ( v8 )
                          {
                            v1342 = sub_18000CC40(v1340);
                            (*(void (__fastcall **)(int *))(v1342 + 48))(v1341);
                          }
                        }
// Original IDA line 5435 | qword_1803BE0F8 -> novels.services.agent_session#61 bytes [4307721, 4307728]: "_emit"
                        v1343 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE0F8);
                        if ( v1343 != nullptr )
                        {
                          v1345 = sub_180001510(v1336);
                          if ( v1345 == -1 )
                          {
                            sub_18000CE30(a1, &v1438);
                            if ( *v1343 >= 0 )
                            {
                              v8 = (*(_QWORD *)v1343)-- == 1;
                              if ( v8 )
                              {
                                v1346 = sub_18000CC40(v1343);
                                (*(void (__fastcall **)(int *))(v1346 + 48))(v1343);
                              }
                            }
                            v1344 = 570;
                          }
                          else
                          {
// Original IDA line 5455 | qword_1803BE6D8 -> novels.services.agent_session#249 bytes [4310955, 4310961]: "task"
                            v1347 = qword_1803BE6D8;
                            if ( v1345 == 1 )
                              v1347 = v1336;
// Original IDA line 5458 | qword_1803BE6E0 -> novels.services.agent_session#250 bytes [4310961, 4310978]: "执行失败："
                            v1348 = (_DWORD *)qword_1803BE6E0;
                            v1442 = (_QWORD *)Py_FalseStruct;
                            v1349 = sub_180374A40(a1, 2);
                            v1350 = (int *)v1349;
                            if ( *v1348 != -1 )
                              ++*v1348;
                            sub_18000CC80(v1349, 0, v1348);
// Original IDA line 5465 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                            v1351 = sub_180375520(a1, v1335, qword_1803BDFE8);
                            if ( v1351 != 0 )
                            {
                              sub_18000CC80(v1350, 1, v1351);
// Original IDA line 5469 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                              v1354 = (int *)PyUnicode_Join(qword_1803BDFE8, v1350);
                              if ( *v1350 >= 0 )
                              {
                                v8 = (*(_QWORD *)v1350)-- == 1;
                                if ( v8 )
                                {
                                  v1355 = sub_18000CC40(v1350);
                                  (*(void (__fastcall **)(int *))(v1355 + 48))(v1350);
                                }
                              }
                              if ( v1354 != nullptr )
                              {
                                v1344 = 568;
                                *(_DWORD *)(v1440 + 40) = 568;
                                v1536[1] = v1442;
                                v1536[0] = v1347;
                                v1536[2] = v1354;
                                v1536[3] = v1335;
                                v1536[4] = v1335;
// Original IDA line 5488 | qword_1803BE6F0 -> novels.services.agent_session#252 bytes [4310990, 4311032]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310992, 4310998], "value": "tool"}, {"tag": "a", "fileRange": [4310998, 4311007], "value": "success"}, {"tag": "a", "fileRange": [4311007, 4311016], "value": "summary"}, {"tag": "a", "... (truncated literal); qword_1803BE6E8 -> novels.services.agent_session#251 bytes [4310978, 4310990]: {"type": "tuple", "items": [{"tag": "u", "fileRange": [4310980, 4310990], "value": "tool:end"}]}
                                v1357 = (int *)sub_18037D280(a1, v1343, qword_1803BE6E8, v1536, qword_1803BE6F0);
                                if ( *v1343 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1343)-- == 1;
                                  if ( v8 )
                                  {
                                    v1358 = sub_18000CC40(v1343);
                                    (*(void (__fastcall **)(int *))(v1358 + 48))(v1343);
                                  }
                                }
                                if ( *v1354 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1354)-- == 1;
                                  if ( v8 )
                                  {
                                    v1359 = sub_18000CC40(v1354);
                                    (*(void (__fastcall **)(int *))(v1359 + 48))(v1354);
                                  }
                                }
                                if ( v1357 != nullptr )
                                {
                                  if ( *v1357 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1357)-- == 1;
                                    if ( v8 )
                                    {
                                      v1365 = sub_18000CC40(v1357);
                                      (*(void (__fastcall **)(int *))(v1365 + 48))(v1357);
                                    }
                                  }
                                  v10 = v1440;
                                  goto LABEL_2610;
                                }
                                sub_18000CE30(a1, &v1438);
                              }
                              else
                              {
                                sub_18000CE30(a1, &v1438);
                                if ( *v1343 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1343)-- == 1;
                                  if ( v8 )
                                  {
                                    v1356 = sub_18000CC40(v1343);
                                    (*(void (__fastcall **)(int *))(v1356 + 48))(v1343);
                                  }
                                }
                                v1344 = 572;
                              }
                            }
                            else
                            {
                              sub_18000CE30(a1, &v1438);
                              v1344 = 572;
                              if ( *v1343 >= 0 )
                              {
                                v8 = (*(_QWORD *)v1343)-- == 1;
                                if ( v8 )
                                {
                                  v1352 = sub_18000CC40(v1343);
                                  (*(void (__fastcall **)(int *))(v1352 + 48))(v1343);
                                }
                              }
                              if ( *v1350 >= 0 )
                              {
                                v8 = (*(_QWORD *)v1350)-- == 1;
                                if ( v8 )
                                {
                                  v1353 = sub_18000CC40(v1350);
                                  (*(void (__fastcall **)(int *))(v1353 + 48))(v1350);
                                }
                              }
                            }
                            v10 = v1440;
                          }
                        }
                        else
                        {
                          sub_18000CE30(a1, &v1438);
                          v1344 = 568;
                        }
                        v1472 = v1438;
                        sub_18000CE70(&v1438);
                        v1489 = sub_18000CCC0(a1);
                        v1360 = sub_18000CE80(&v1472);
                        if ( v1360 != 0 )
                          v1361 = sub_18000CC90(v1360, v10, v1344);
                        else
                          v1361 = sub_18038E2A0(v10, v1344);
                        sub_180001300(&v1472, v1361);
                        sub_18000D050(a1, &v1472);
                        v1362 = sub_18000CC40(*(_QWORD *)(a1 + 248));
                        if ( (unsigned int)sub_18000CE90(a1, v1362, PyExc_Exception) == 0 )
                        {
                          v14 = 567;
                          if ( (unsigned __int8)sub_180373290(a1, &v1438) != 0 )
                            v14 = 0;
                          v1363 = sub_18000CE80(&v1438);
                          if ( v1363 != 0 && *(_QWORD *)(v1363 + 24) == v10 )
                            *(_DWORD *)(v10 + 40) = *(_DWORD *)(v1363 + 36);
                          v1364 = v1438;
                          sub_18000CE70(&v1438);
                          sub_18000CCE0(a1, &v1489);
                          v1438 = v1364;
                          goto LABEL_2615;
                        }
                        sub_18000CCE0(a1, &v1489);
LABEL_2610:
                        v14 = 578;
                        if ( (unsigned __int8)sub_180373290(a1, &v1438) != 0 )
                          v14 = 0;
                        v1366 = sub_18000CE80(&v1438);
                        if ( v1366 != 0 && *(_QWORD *)(v1366 + 24) == v10 )
                          *(_DWORD *)(v10 + 40) = *(_DWORD *)(v1366 + 36);
                        goto LABEL_2615;
                      }
                    }
                    sub_18000CE30(a1, &v1438);
LABEL_2615:
                    v1317 = v1447;
LABEL_2616:
                    v1367 = v1438;
                    sub_18000CE70(&v1438);
                    if ( v1317 != nullptr && *v1317 >= 0 )
                    {
                      v8 = (*(_QWORD *)v1317)-- == 1;
                      if ( v8 )
                      {
                        v1368 = sub_18000CC40(v1317);
                        (*(void (__fastcall **)(int *))(v1368 + 48))(v1317);
                      }
                    }
                    v1438 = v1367;
LABEL_2627:
                    sub_18000CE70(&v1438);
                    sub_18000CCE0(a1, &v1507);
                    v1438 = v1367;
LABEL_2628:
                    v5 = v1443;
                    goto LABEL_2629;
                  }
                }
                sub_18000CE30(a1, &v1438);
                goto LABEL_2616;
              }
              *(_DWORD *)(v10 + 40) = 393;
// Original IDA line 5634 | qword_1803BE480 -> novels.services.agent_session#174 bytes [4309694, 4309704]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309696, 4309703], "value": "title"}, {"tag": "s", "fileRange": [4309703, 4309704], "value": ""}]}
              v605 = (int *)sub_180378210(a1, v603, qword_1803BE480);
              if ( *v604 >= 0 )
              {
                v8 = (*(_QWORD *)v604)-- == 1;
                if ( v8 )
                {
                  v606 = sub_18000CC40(v604);
                  (*(void (__fastcall **)(int *))(v606 + 48))(v604);
                }
              }
              if ( v605 == nullptr )
                goto LABEL_1133;
// Original IDA line 5646 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              v607 = sub_180375520(a1, v605, qword_1803BDFE8);
              if ( *v605 >= 0 )
              {
                v8 = (*(_QWORD *)v605)-- == 1;
                if ( v8 )
                {
                  v608 = sub_18000CC40(v605);
                  (*(void (__fastcall **)(int *))(v608 + 48))(v605);
                }
              }
              if ( v607 == 0 )
              {
LABEL_1133:
                v10 = v1440;
                goto LABEL_1134;
              }
              sub_18000CC80(v602, 1, v607);
// Original IDA line 5663 | qword_1803BE4D8 -> novels.services.agent_session#185 bytes [4309849, 4309863]: "》已启动"
              v611 = qword_1803BE4D8;
// Original IDA line 5664 | qword_1803BE4D8 -> novels.services.agent_session#185 bytes [4309849, 4309863]: "》已启动"
              if ( *(_DWORD *)qword_1803BE4D8 != -1 )
// Original IDA line 5665 | qword_1803BE4D8 -> novels.services.agent_session#185 bytes [4309849, 4309863]: "》已启动"
                ++*(_DWORD *)qword_1803BE4D8;
              sub_18000CC80(v602, 2, v611);
// Original IDA line 5667 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              v612 = (int *)PyUnicode_Join(qword_1803BDFE8, v602);
              if ( *v602 >= 0 )
              {
                v8 = (*(_QWORD *)v602)-- == 1;
                if ( v8 )
                {
                  v613 = sub_18000CC40(v602);
                  (*(void (__fastcall **)(int *))(v613 + 48))(v602);
                }
              }
              v316 = a1;
              if ( v612 == nullptr )
                goto LABEL_572;
              v614 = v1442;
              v10 = v1440;
              v151 = 391;
              *(_DWORD *)(v1440 + 40) = 391;
              v1513[0] = v614;
              v1498 = v1449;
              v1513[1] = v612;
// Original IDA line 5687 | qword_1803BE4E0 -> novels.services.agent_session#186 bytes [4309863, 4309874]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309865, 4309874], "value": "details"}]}
              v615 = (int *)sub_18037D650(a1, v23, v1513, &v1498, qword_1803BE4E0);
              if ( *v23 >= 0 )
              {
                v8 = (*(_QWORD *)v23)-- == 1;
                if ( v8 )
                {
                  v616 = sub_18000CC40(v23);
                  (*(void (__fastcall **)(int *))(v616 + 48))(v23);
                }
              }
              if ( *v612 >= 0 )
              {
                v8 = (*(_QWORD *)v612)-- == 1;
                if ( v8 )
                {
                  v617 = sub_18000CC40(v612);
                  (*(void (__fastcall **)(int *))(v617 + 48))(v612);
                }
              }
              v435 = a1;
              if ( v615 == nullptr )
                goto LABEL_798;
              *(_DWORD *)(v10 + 40) = 391;
              v618 = (int *)sub_180376770(a1, v177, v615);
              if ( *v177 >= 0 )
              {
                v8 = (*(_QWORD *)v177)-- == 1;
                if ( v8 )
                {
                  v619 = sub_18000CC40(v177);
                  (*(void (__fastcall **)(int *))(v619 + 48))(v177);
                }
              }
              if ( *v615 >= 0 )
              {
                v8 = (*(_QWORD *)v615)-- == 1;
                if ( v8 )
                {
                  v620 = sub_18000CC40(v615);
                  (*(void (__fastcall **)(int *))(v620 + 48))(v615);
                }
              }
              if ( v618 == nullptr )
                goto LABEL_2475;
              if ( *v618 >= 0 )
              {
                v8 = (*(_QWORD *)v618)-- == 1;
                if ( v8 )
                {
                  v621 = sub_18000CC40(v618);
                  (*(void (__fastcall **)(int *))(v621 + 48))(v618);
                }
              }
// Original IDA line 5740 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
              v622 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
              v623 = (int *)v622;
              if ( v622 != 0 )
              {
// Original IDA line 5744 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                v70 = (int *)sub_18037F0C0(a1, v622, qword_1803BE0B8);
                if ( *v623 >= 0 )
                {
                  v8 = (*(_QWORD *)v623)-- == 1;
                  if ( v8 )
                  {
                    v624 = sub_18000CC40(v623);
                    (*(void (__fastcall **)(int *))(v624 + 48))(v623);
                  }
                }
                if ( v70 != nullptr )
                {
// Original IDA line 5756 | qword_1803BE058 -> novels.services.agent_session#41 bytes [4307538, 4307544]: "role"
                  v625 = qword_1803BE058;
// Original IDA line 5757 | qword_1803BE240 -> novels.services.agent_session#102 bytes [4308406, 4308417]: "assistant"
                  v626 = qword_1803BE240;
                  v23 = (int *)PyDict_NewPresized(3);
                  PyDict_SetItem(v23, v625, v626);
// Original IDA line 5760 | qword_1803BE068 -> novels.services.agent_session#43 bytes [4307552, 4307561]: "content"
                  PyDict_SetItem(v23, qword_1803BE068, v1462);
// Original IDA line 5761 | qword_1803BE0C0 -> novels.services.agent_session#54 bytes [4307660, 4307671]: "createdAt"
                  v627 = qword_1803BE0C0;
                  v628 = sub_1801F4170(a1);
                  if ( v628 == 0 )
                  {
// Original IDA line 5765 | qword_1803BE0C8 -> novels.services.agent_session#55 bytes [4307671, 4307684]: "utc_now_iso"
                    sub_180373340(a1, &v1438, qword_1803BE0C8);
                    v151 = 400;
LABEL_1176:
                    if ( *v70 >= 0 )
                    {
                      v8 = (*(_QWORD *)v70)-- == 1;
                      if ( v8 )
                      {
                        v631 = sub_18000CC40(v70);
                        (*(void (__fastcall **)(int *))(v631 + 48))(v70);
                      }
                    }
LABEL_2512:
                    if ( *v23 < 0 )
                      goto LABEL_2516;
                    v8 = (*(_QWORD *)v23)-- == 1;
                    if ( !v8 )
                      goto LABEL_2516;
                    v436 = sub_18000CC40(v23);
                    v437 = v23;
                    goto LABEL_2515;
                  }
                  v151 = 400;
                  *(_DWORD *)(v10 + 40) = 400;
                  v629 = sub_180376210(a1, v628);
                  v630 = (int *)v629;
                  if ( v629 != 0 )
                  {
                    PyDict_SetItem(v23, v627, v629);
                    if ( *v630 >= 0 )
                    {
                      v8 = (*(_QWORD *)v630)-- == 1;
                      if ( v8 )
                      {
                        v632 = sub_18000CC40(v630);
                        (*(void (__fastcall **)(int *))(v632 + 48))(v630);
                      }
                    }
                    v151 = 396;
                    *(_DWORD *)(v10 + 40) = 396;
                    v633 = (int *)sub_180376770(a1, v70, v23);
                    if ( *v70 >= 0 )
                    {
                      v8 = (*(_QWORD *)v70)-- == 1;
                      if ( v8 )
                      {
                        v634 = sub_18000CC40(v70);
                        (*(void (__fastcall **)(int *))(v634 + 48))(v70);
                      }
                    }
                    if ( *v23 >= 0 )
                    {
                      v8 = (*(_QWORD *)v23)-- == 1;
                      if ( v8 )
                      {
                        v635 = sub_18000CC40(v23);
                        (*(void (__fastcall **)(int *))(v635 + 48))(v23);
                      }
                    }
                    if ( v633 == nullptr )
                      goto LABEL_2475;
                    if ( *v633 >= 0 )
                    {
                      v8 = (*(_QWORD *)v633)-- == 1;
                      if ( v8 )
                      {
                        v636 = sub_18000CC40(v633);
                        (*(void (__fastcall **)(int *))(v636 + 48))(v633);
                      }
                    }
// Original IDA line 5835 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                    v637 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                    v638 = (int *)v637;
                    if ( v637 == 0 )
                    {
                      sub_18000CE30(a1, &v1438);
                      v151 = 403;
                      goto LABEL_2516;
                    }
                    v639 = v1445;
                    v151 = 403;
                    v640 = *(_QWORD *)(v1445 + 16);
                    *(_DWORD *)(v10 + 40) = 403;
// Original IDA line 5847 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                    v641 = (int *)sub_18037E6C0(a1, v637, qword_1803BE0D8, v640);
                    if ( *v638 >= 0 )
                    {
                      v8 = (*(_QWORD *)v638)-- == 1;
                      if ( v8 )
                      {
                        v642 = sub_18000CC40(v638);
                        (*(void (__fastcall **)(int *))(v642 + 48))(v638);
                      }
                    }
                    if ( v641 == nullptr )
                      goto LABEL_2475;
                    if ( *v641 >= 0 )
                    {
                      v8 = (*(_QWORD *)v641)-- == 1;
                      if ( v8 )
                      {
                        v643 = sub_18000CC40(v641);
                        (*(void (__fastcall **)(int *))(v643 + 48))(v641);
                      }
                    }
// Original IDA line 5868 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                    v644 = qword_1803BE2C8;
                    v23 = (int *)PyDict_NewPresized(2);
                    PyDict_SetItem(v23, v644, Py_TrueStruct);
// Original IDA line 5871 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                    v248 = qword_1803BE2F8;
// Original IDA line 5872 | qword_1803BDFA8 -> novels.services.agent_session#19 bytes [4307205, 4307211]: "play"
                    v645 = qword_1803BDFA8;
                    v252 = (int *)PyDict_NewPresized(3);
                    PyDict_SetItem(v252, v645, v1449);
                    v151 = 408;
// Original IDA line 5876 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
                    v646 = qword_1803BE1B8;
                    v647 = *(_QWORD *)(v639 + 16);
                    *(_DWORD *)(v1440 + 40) = 408;
// Original IDA line 5879 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
                    v648 = sub_18037E560(a1, v647, qword_1803BE278);
                    v649 = (int *)v648;
                    if ( v648 != 0 )
                    {
                      PyDict_SetItem(v252, v646, v648);
                      if ( *v649 >= 0 )
                      {
                        v8 = (*(_QWORD *)v649)-- == 1;
                        if ( v8 )
                        {
                          v650 = sub_18000CC40(v649);
                          (*(void (__fastcall **)(int *))(v650 + 48))(v649);
                        }
                      }
// Original IDA line 5893 | qword_1803BE4E8 -> novels.services.agent_session#187 bytes [4309874, 4309888]: "switchToPlay"
                      PyDict_SetItem(v252, qword_1803BE4E8, Py_TrueStruct);
                      goto LABEL_971;
                    }
                    goto LABEL_965;
                  }
LABEL_1175:
                  sub_18000CE30(a1, &v1438);
                  goto LABEL_1176;
                }
              }
              sub_18000CE30(a1, &v1438);
              v151 = 396;
              goto LABEL_2516;
            }
          }
          sub_18000CE30(a1, &v1438);
          v151 = 391;
          goto LABEL_2516;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 380;
      goto LABEL_2516;
    }
// Original IDA line 5917 | qword_1803BE4F0 -> novels.services.agent_session#188 bytes [4309888, 4309904]: "generate_cover"
    v651 = sub_1803836D0(v45, qword_1803BE4F0);
    if ( v651 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 413;
      goto LABEL_2516;
    }
    if ( v651 == 1 )
    {
// Original IDA line 5926 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
      v652 = sub_18037F0C0(a1, *(_QWORD *)(v148 + 16), qword_1803BDF18);
      v653 = (int *)v652;
      if ( v652 != 0 )
      {
// Original IDA line 5930 | qword_1803BE4F0 -> novels.services.agent_session#188 bytes [4309888, 4309904]: "generate_cover"
        v654 = (int *)sub_18037F0C0(a1, v652, qword_1803BE4F0);
        if ( *v653 >= 0 )
        {
          v8 = (*(_QWORD *)v653)-- == 1;
          if ( v8 )
          {
            v655 = sub_18000CC40(v653);
            (*(void (__fastcall **)(int *))(v655 + 48))(v653);
          }
        }
        if ( v654 != nullptr )
        {
// Original IDA line 5942 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
          v656 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
          if ( v656 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v654 >= 0 )
            {
              v8 = (*(_QWORD *)v654)-- == 1;
              if ( v8 )
              {
                v657 = sub_18000CC40(v654);
                (*(void (__fastcall **)(int *))(v657 + 48))(v654);
              }
            }
            v151 = 416;
            goto LABEL_2516;
          }
// Original IDA line 5958 | qword_1803BDF50 -> novels.services.agent_session#8 bytes [4307033, 4307056]: "cover_image_generator"
          v658 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF50);
          if ( v658 == nullptr )
          {
            sub_18000CE30(a1, &v1438);
            if ( *v654 >= 0 )
            {
              v8 = (*(_QWORD *)v654)-- == 1;
              if ( v8 )
              {
                v659 = sub_18000CC40(v654);
                (*(void (__fastcall **)(int *))(v659 + 48))(v654);
              }
            }
            if ( *v656 >= 0 )
            {
              v8 = (*(_QWORD *)v656)-- == 1;
              if ( v8 )
              {
                v660 = sub_18000CC40(v656);
                (*(void (__fastcall **)(int *))(v660 + 48))(v656);
              }
            }
            v151 = 418;
            goto LABEL_2516;
          }
          v151 = 414;
          *(_DWORD *)(v10 + 40) = 414;
          v1499 = v5;
          v1531[1] = v1456;
          v1531[0] = v656;
          v1531[2] = v658;
// Original IDA line 5989 | qword_1803BE4F8 -> novels.services.agent_session#189 bytes [4309904, 4309939]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309906, 4309915], "value": "book_id"}, {"tag": "a", "fileRange": [4309915, 4309922], "value": "model"}, {"tag": "a", "fileRange": [4309922, 4309939], "value": "image_generator"}]}
          v661 = sub_18037D030(a1, v654, &v1499, v1531, qword_1803BE4F8);
          v29 = *v654 < 0;
          v662 = (int *)v661;
          v1449 = (int *)v661;
          if ( !v29 )
          {
            v8 = (*(_QWORD *)v654)-- == 1;
            if ( v8 )
            {
              v663 = sub_18000CC40(v654);
              (*(void (__fastcall **)(int *))(v663 + 48))(v654);
            }
          }
          if ( *v656 >= 0 )
          {
            v8 = (*(_QWORD *)v656)-- == 1;
            if ( v8 )
            {
              v664 = sub_18000CC40(v656);
              (*(void (__fastcall **)(int *))(v664 + 48))(v656);
            }
          }
          if ( *v658 >= 0 )
          {
            v8 = (*(_QWORD *)v658)-- == 1;
            if ( v8 )
            {
              v665 = sub_18000CC40(v658);
              (*(void (__fastcall **)(int *))(v665 + 48))(v658);
            }
          }
          if ( v662 == nullptr )
            goto LABEL_2475;
          v1457 = v662;
// Original IDA line 6023 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
          if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 420;
            goto LABEL_2516;
          }
// Original IDA line 6029 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
          v666 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
          v667 = (int *)v666;
          if ( v666 != 0 )
          {
// Original IDA line 6033 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
            v177 = (int *)sub_18037F0C0(a1, v666, qword_1803BE0B8);
            if ( *v667 >= 0 )
            {
              v8 = (*(_QWORD *)v667)-- == 1;
              if ( v8 )
              {
                v668 = sub_18000CC40(v667);
                (*(void (__fastcall **)(int *))(v668 + 48))(v667);
              }
            }
            if ( v177 != nullptr )
            {
// Original IDA line 6045 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
              v23 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
              if ( v23 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v177 >= 0 )
                {
                  v8 = (*(_QWORD *)v177)-- == 1;
                  if ( v8 )
                  {
                    v669 = sub_18000CC40(v177);
                    (*(void (__fastcall **)(int *))(v669 + 48))(v177);
                  }
                }
                v151 = 421;
                goto LABEL_2516;
              }
// Original IDA line 6061 | qword_1803BE500 -> novels.services.agent_session#190 bytes [4309939, 4309944]: "《"
              v670 = (_DWORD *)qword_1803BE500;
// Original IDA line 6062 | qword_1803BE4F0 -> novels.services.agent_session#188 bytes [4309888, 4309904]: "generate_cover"
              v1442 = (_QWORD *)qword_1803BE4F0;
              v671 = sub_180374A40(a1, 3);
              v602 = (int *)v671;
              if ( *v670 != -1 )
                ++*v670;
              sub_18000CC80(v671, 0, v670);
// Original IDA line 6068 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
              v672 = sub_18037F0C0(a1, v662, qword_1803BDFD8);
              v673 = (int *)v672;
              v151 = 423;
              if ( v672 == 0 )
                goto LABEL_1134;
              *(_DWORD *)(v10 + 40) = 423;
// Original IDA line 6074 | qword_1803BE480 -> novels.services.agent_session#174 bytes [4309694, 4309704]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309696, 4309703], "value": "title"}, {"tag": "s", "fileRange": [4309703, 4309704], "value": ""}]}
              v674 = (int *)sub_180378210(a1, v672, qword_1803BE480);
              if ( *v673 >= 0 )
              {
                v8 = (*(_QWORD *)v673)-- == 1;
                if ( v8 )
                {
                  v675 = sub_18000CC40(v673);
                  (*(void (__fastcall **)(int *))(v675 + 48))(v673);
                }
              }
              if ( v674 == nullptr )
                goto LABEL_1133;
// Original IDA line 6086 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              v676 = sub_180375520(a1, v674, qword_1803BDFE8);
              if ( *v674 >= 0 )
              {
                v8 = (*(_QWORD *)v674)-- == 1;
                if ( v8 )
                {
                  v677 = sub_18000CC40(v674);
                  (*(void (__fastcall **)(int *))(v677 + 48))(v674);
                }
              }
              if ( v676 == 0 )
                goto LABEL_1133;
              sub_18000CC80(v602, 1, v676);
// Original IDA line 6099 | qword_1803BE508 -> novels.services.agent_session#191 bytes [4309944, 4309964]: "》封面已处理"
              v678 = qword_1803BE508;
// Original IDA line 6100 | qword_1803BE508 -> novels.services.agent_session#191 bytes [4309944, 4309964]: "》封面已处理"
              if ( *(_DWORD *)qword_1803BE508 != -1 )
// Original IDA line 6101 | qword_1803BE508 -> novels.services.agent_session#191 bytes [4309944, 4309964]: "》封面已处理"
                ++*(_DWORD *)qword_1803BE508;
              sub_18000CC80(v602, 2, v678);
// Original IDA line 6103 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
              v679 = (int *)PyUnicode_Join(qword_1803BDFE8, v602);
              if ( *v602 >= 0 )
              {
                v8 = (*(_QWORD *)v602)-- == 1;
                if ( v8 )
                {
                  v680 = sub_18000CC40(v602);
                  (*(void (__fastcall **)(int *))(v680 + 48))(v602);
                }
              }
              v316 = a1;
              if ( v679 == nullptr )
                goto LABEL_572;
              v681 = v1442;
              v10 = v1440;
              v151 = 421;
              *(_DWORD *)(v1440 + 40) = 421;
              v1514[0] = v681;
              v1500 = v1449;
              v1514[1] = v679;
// Original IDA line 6123 | qword_1803BE4E0 -> novels.services.agent_session#186 bytes [4309863, 4309874]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309865, 4309874], "value": "details"}]}
              v682 = (int *)sub_18037D650(a1, v23, v1514, &v1500, qword_1803BE4E0);
              if ( *v23 >= 0 )
              {
                v8 = (*(_QWORD *)v23)-- == 1;
                if ( v8 )
                {
                  v683 = sub_18000CC40(v23);
                  (*(void (__fastcall **)(int *))(v683 + 48))(v23);
                }
              }
              if ( *v679 >= 0 )
              {
                v8 = (*(_QWORD *)v679)-- == 1;
                if ( v8 )
                {
                  v684 = sub_18000CC40(v679);
                  (*(void (__fastcall **)(int *))(v684 + 48))(v679);
                }
              }
              v435 = a1;
              if ( v682 == nullptr )
                goto LABEL_798;
              *(_DWORD *)(v10 + 40) = 421;
              v685 = (int *)sub_180376770(a1, v177, v682);
              if ( *v177 >= 0 )
              {
                v8 = (*(_QWORD *)v177)-- == 1;
                if ( v8 )
                {
                  v686 = sub_18000CC40(v177);
                  (*(void (__fastcall **)(int *))(v686 + 48))(v177);
                }
              }
              if ( *v682 >= 0 )
              {
                v8 = (*(_QWORD *)v682)-- == 1;
                if ( v8 )
                {
                  v687 = sub_18000CC40(v682);
                  (*(void (__fastcall **)(int *))(v687 + 48))(v682);
                }
              }
              if ( v685 == nullptr )
                goto LABEL_2475;
              if ( *v685 >= 0 )
              {
                v8 = (*(_QWORD *)v685)-- == 1;
                if ( v8 )
                {
                  v688 = sub_18000CC40(v685);
                  (*(void (__fastcall **)(int *))(v688 + 48))(v685);
                }
              }
// Original IDA line 6176 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
              v689 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
              v690 = (int *)v689;
              if ( v689 == 0 )
              {
                sub_18000CE30(a1, &v1438);
                v151 = 426;
                goto LABEL_2516;
              }
              v691 = v1445;
              v151 = 426;
              v692 = *(_QWORD *)(v1445 + 16);
              *(_DWORD *)(v10 + 40) = 426;
// Original IDA line 6188 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
              v693 = (int *)sub_18037E6C0(a1, v689, qword_1803BE0D8, v692);
              if ( *v690 >= 0 )
              {
                v8 = (*(_QWORD *)v690)-- == 1;
                if ( v8 )
                {
                  v694 = sub_18000CC40(v690);
                  (*(void (__fastcall **)(int *))(v694 + 48))(v690);
                }
              }
              if ( v693 == nullptr )
                goto LABEL_2475;
              if ( *v693 >= 0 )
              {
                v8 = (*(_QWORD *)v693)-- == 1;
                if ( v8 )
                {
                  v695 = sub_18000CC40(v693);
                  (*(void (__fastcall **)(int *))(v695 + 48))(v693);
                }
              }
// Original IDA line 6209 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
              v696 = qword_1803BE2C8;
              v23 = (int *)PyDict_NewPresized(2);
              PyDict_SetItem(v23, v696, Py_TrueStruct);
// Original IDA line 6212 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
              v448 = qword_1803BE2F8;
// Original IDA line 6213 | qword_1803BE510 -> novels.services.agent_session#192 bytes [4309964, 4309971]: "cover"
              v697 = qword_1803BE510;
// Original IDA line 6214 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
              v1442 = (_QWORD *)qword_1803BE2F8;
              v70 = (int *)PyDict_NewPresized(3);
              PyDict_SetItem(v70, v697, v1449);
              v698 = *(_QWORD *)(v691 + 16);
// Original IDA line 6218 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
              v1448 = (int *)qword_1803BE348;
// Original IDA line 6219 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
              v699 = sub_18037F0C0(a1, v698, qword_1803BDFC0);
              v700 = (int *)v699;
              if ( v699 == 0 )
              {
LABEL_1313:
                sub_18000CE30(a1, &v1438);
                v151 = 431;
                goto LABEL_824;
              }
              v701 = sub_180001510(v699);
              if ( v701 == -1 )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v700 >= 0 )
                {
                  v8 = (*(_QWORD *)v700)-- == 1;
                  if ( v8 )
                  {
                    v702 = sub_18000CC40(v700);
                    (*(void (__fastcall **)(int *))(v702 + 48))(v700);
                  }
                }
LABEL_1303:
                v151 = 431;
                goto LABEL_824;
              }
              if ( *v700 >= 0 )
              {
                v8 = (*(_QWORD *)v700)-- == 1;
                if ( v8 )
                {
                  v703 = sub_18000CC40(v700);
                  (*(void (__fastcall **)(int *))(v703 + 48))(v700);
                }
              }
              if ( v701 != 0 )
              {
// Original IDA line 6256 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                v704 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                v705 = (int *)v704;
                if ( v704 == 0 )
                  goto LABEL_1313;
// Original IDA line 6260 | qword_1803BDFC8 -> novels.services.agent_session#23 bytes [4307250, 4307268]: "get_book_payload"
                v706 = (int *)sub_18037F0C0(a1, v704, qword_1803BDFC8);
                if ( *v705 >= 0 )
                {
                  v8 = (*(_QWORD *)v705)-- == 1;
                  if ( v8 )
                  {
                    v707 = sub_18000CC40(v705);
                    (*(void (__fastcall **)(int *))(v707 + 48))(v705);
                  }
                }
                if ( v706 == nullptr )
                  goto LABEL_1313;
// Original IDA line 6272 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                v708 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v691 + 16), qword_1803BDFC0);
                if ( v708 == nullptr )
                {
                  sub_18000CE30(a1, &v1438);
                  if ( *v706 >= 0 )
                  {
                    v8 = (*(_QWORD *)v706)-- == 1;
                    if ( v8 )
                    {
                      v709 = sub_18000CC40(v706);
                      (*(void (__fastcall **)(int *))(v709 + 48))(v706);
                      v151 = 431;
                      goto LABEL_824;
                    }
                  }
                  goto LABEL_1303;
                }
                v151 = 431;
                *(_DWORD *)(v1440 + 40) = 431;
                v710 = (int *)sub_180376770(a1, v706, v708);
                if ( *v706 >= 0 )
                {
                  v8 = (*(_QWORD *)v706)-- == 1;
                  if ( v8 )
                  {
                    v711 = sub_18000CC40(v706);
                    (*(void (__fastcall **)(int *))(v711 + 48))(v706);
                  }
                }
                if ( *v708 >= 0 )
                {
                  v8 = (*(_QWORD *)v708)-- == 1;
                  if ( v8 )
                  {
                    v712 = sub_18000CC40(v708);
                    (*(void (__fastcall **)(int *))(v712 + 48))(v708);
                  }
                }
                if ( v710 == nullptr )
                  goto LABEL_823;
                v448 = (__int64)v1442;
              }
              else
              {
                v710 = (int *)Py_NoneStruct;
                if ( Py_NoneStruct != -1 )
                  ++Py_NoneStruct;
              }
              PyDict_SetItem(v70, v1448, v710);
              if ( *v710 >= 0 )
              {
                v8 = (*(_QWORD *)v710)-- == 1;
                if ( v8 )
                {
                  v713 = sub_18000CC40(v710);
                  (*(void (__fastcall **)(int *))(v713 + 48))(v710);
                }
              }
              v151 = 432;
LABEL_822:
// Original IDA line 6332 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
              v450 = qword_1803BE1B8;
              v451 = *(_QWORD *)(v1445 + 16);
              *(_DWORD *)(v1440 + 40) = v151;
// Original IDA line 6335 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
              v452 = sub_18037E560(a1, v451, qword_1803BE278);
              v453 = (int *)v452;
              if ( v452 == 0 )
              {
LABEL_823:
                sub_18000CE30(a1, &v1438);
LABEL_824:
                if ( *v70 >= 0 )
                {
                  v8 = (*(_QWORD *)v70)-- == 1;
                  if ( v8 )
                  {
                    v454 = sub_18000CC40(v70);
                    v455 = v70;
LABEL_827:
                    (*(void (__fastcall **)(int *))(v454 + 48))(v455);
                    goto LABEL_828;
                  }
                }
                goto LABEL_828;
              }
              PyDict_SetItem(v70, v450, v452);
              if ( *v453 >= 0 )
              {
                v8 = (*(_QWORD *)v453)-- == 1;
                if ( v8 )
                {
                  v456 = sub_18000CC40(v453);
                  (*(void (__fastcall **)(int *))(v456 + 48))(v453);
                }
              }
              v93 = (_QWORD *)v448;
LABEL_146:
              PyDict_SetItem(v23, v93, v70);
              if ( *v70 >= 0 )
              {
                v8 = (*(_QWORD *)v70)-- == 1;
                if ( v8 )
                {
                  v94 = sub_18000CC40(v70);
                  (*(void (__fastcall **)(int *))(v94 + 48))(v70);
                }
              }
LABEL_149:
              v5 = v1451;
              goto LABEL_150;
            }
          }
          sub_18000CE30(a1, &v1438);
          v151 = 421;
          goto LABEL_2516;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 414;
      goto LABEL_2516;
    }
// Original IDA line 6392 | qword_1803BE518 -> novels.services.agent_session#193 bytes [4309971, 4309983]: "write_next"
    v714 = sub_1803836D0(v45, qword_1803BE518);
    if ( v714 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 436;
      goto LABEL_2516;
    }
    if ( v714 == 1 )
    {
// Original IDA line 6401 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
      v715 = sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
      v716 = (int *)v715;
      if ( v715 != 0 )
      {
        v717 = sub_180001510(v715);
        if ( *v716 >= 0 )
        {
          v8 = (*(_QWORD *)v716)-- == 1;
          if ( v8 )
          {
            v718 = sub_18000CC40(v716);
            (*(void (__fastcall **)(int *))(v718 + 48))(v716);
          }
        }
        if ( v717 != -1 )
        {
          if ( v717 == 0 )
            goto LABEL_2269;
          v719 = v1441;
// Original IDA line 6420 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
          v720 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
          v721 = (int *)v720;
          if ( v720 != 0 )
          {
// Original IDA line 6424 | qword_1803BE528 -> novels.services.agent_session#195 bytes [4310033, 4310053]: "write_next_chapter"
            v722 = (int *)sub_18037F0C0(a1, v720, qword_1803BE528);
            if ( *v721 >= 0 )
            {
              v8 = (*(_QWORD *)v721)-- == 1;
              if ( v8 )
              {
                v723 = sub_18000CC40(v721);
                (*(void (__fastcall **)(int *))(v723 + 48))(v721);
              }
            }
            if ( v722 != nullptr )
            {
// Original IDA line 6436 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
              v724 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
              if ( v724 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v722 >= 0 )
                {
                  v8 = (*(_QWORD *)v722)-- == 1;
                  if ( v8 )
                  {
                    v725 = sub_18000CC40(v722);
                    (*(void (__fastcall **)(int *))(v725 + 48))(v722);
                  }
                }
                v151 = 439;
                goto LABEL_2516;
              }
              v726 = v1456;
              v151 = 439;
              *(_DWORD *)(v10 + 40) = 439;
              v1501 = v726;
              v1502 = v724;
// Original IDA line 6457 | qword_1803BE340 -> novels.services.agent_session#134 bytes [4309052, 4309061]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309054, 4309061], "value": "model"}]}
              v727 = sub_18037D030(a1, v722, &v1502, &v1501, qword_1803BE340);
              v29 = *v722 < 0;
              v728 = (int *)v727;
              v1454 = (_QWORD *)v727;
              if ( !v29 )
              {
                v8 = (*(_QWORD *)v722)-- == 1;
                if ( v8 )
                {
                  v729 = sub_18000CC40(v722);
                  (*(void (__fastcall **)(int *))(v729 + 48))(v722);
                }
              }
              if ( *v724 >= 0 )
              {
                v8 = (*(_QWORD *)v724)-- == 1;
                if ( v8 )
                {
                  v730 = sub_18000CC40(v724);
                  (*(void (__fastcall **)(int *))(v730 + 48))(v724);
                }
              }
              if ( v728 == nullptr )
                goto LABEL_2475;
              v731 = *(_QWORD *)(v719 + 16);
              v1457 = v728;
// Original IDA line 6483 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
              v732 = sub_18037F0C0(a1, v731, qword_1803BDF10);
              v733 = (int *)v732;
              if ( v732 != 0 )
              {
// Original IDA line 6487 | qword_1803BE530 -> novels.services.agent_session#196 bytes [4310053, 4310064]: "load_book"
                v734 = (int *)sub_18037F0C0(a1, v732, qword_1803BE530);
                if ( *v733 >= 0 )
                {
                  v8 = (*(_QWORD *)v733)-- == 1;
                  if ( v8 )
                  {
                    v735 = sub_18000CC40(v733);
                    (*(void (__fastcall **)(int *))(v735 + 48))(v733);
                  }
                }
                if ( v734 != nullptr )
                {
                  v736 = v1445;
// Original IDA line 6500 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                  v737 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                  v738 = (int *)v737;
                  if ( v737 == 0 )
                  {
                    sub_18000CE30(a1, &v1438);
                    if ( *v734 >= 0 )
                    {
                      v8 = (*(_QWORD *)v734)-- == 1;
                      if ( v8 )
                      {
                        v739 = sub_18000CC40(v734);
                        (*(void (__fastcall **)(int *))(v739 + 48))(v734);
                      }
                    }
                    v151 = 440;
                    goto LABEL_2516;
                  }
                  v151 = 440;
                  *(_DWORD *)(v10 + 40) = 440;
                  v740 = sub_180376770(a1, v734, v737);
                  if ( *v734 >= 0 )
                  {
                    v8 = (*(_QWORD *)v734)-- == 1;
                    if ( v8 )
                    {
                      v741 = sub_18000CC40(v734);
                      (*(void (__fastcall **)(int *))(v741 + 48))(v734);
                    }
                  }
                  if ( *v738 >= 0 )
                  {
                    v8 = (*(_QWORD *)v738)-- == 1;
                    if ( v8 )
                    {
                      v742 = sub_18000CC40(v738);
                      (*(void (__fastcall **)(int *))(v742 + 48))(v738);
                    }
                  }
                  v310 = a1;
                  if ( v740 == 0 )
                    goto LABEL_2060;
                  v743 = *(_QWORD *)(v736 + 16);
                  v1461 = (int *)v740;
// Original IDA line 6543 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                  if ( (unsigned __int8)sub_18037F3A0(a1, v743, qword_1803BE258, Py_NoneStruct) == 0 )
                  {
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 441;
                    goto LABEL_2516;
                  }
                  v744 = sub_1801F38A0(a1);
                  if ( v744 == 0 )
                  {
// Original IDA line 6553 | qword_1803BE538 -> novels.services.agent_session#197 bytes [4310064, 4310093]: "build_chapter_write_summary"
                    sub_180373340(a1, &v1438, qword_1803BE538);
                    v10 = v1440;
                    v151 = 442;
                    goto LABEL_2516;
                  }
// Original IDA line 6558 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                  v745 = (int *)sub_18037F0C0(a1, v740, qword_1803BE350);
                  if ( v745 == nullptr )
                  {
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 442;
                    goto LABEL_2516;
                  }
                  v151 = 442;
                  *(_DWORD *)(v1440 + 40) = 442;
                  v1515[0] = v745;
                  v1515[1] = v728;
                  v746 = sub_180377750(a1, v744, v1515);
                  v29 = *v745 < 0;
                  v1450 = (_QWORD *)v746;
                  if ( !v29 )
                  {
                    v8 = (*(_QWORD *)v745)-- == 1;
                    if ( v8 )
                    {
                      v747 = sub_18000CC40(v745);
                      (*(void (__fastcall **)(int *))(v747 + 48))(v745);
                      v746 = (__int64)v1450;
                    }
                  }
                  v310 = a1;
                  if ( v746 == 0 )
                    goto LABEL_2060;
                  v748 = *(_QWORD *)(v736 + 16);
                  v1460 = (int *)v746;
// Original IDA line 6588 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                  v749 = sub_18037F0C0(a1, v748, qword_1803BE070);
                  v750 = (int *)v749;
                  if ( v749 != 0 )
                  {
// Original IDA line 6592 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                    v23 = (int *)sub_18037F0C0(a1, v749, qword_1803BE0B8);
                    if ( *v750 >= 0 )
                    {
                      v8 = (*(_QWORD *)v750)-- == 1;
                      if ( v8 )
                      {
                        v751 = sub_18000CC40(v750);
                        (*(void (__fastcall **)(int *))(v751 + 48))(v750);
                      }
                    }
                    if ( v23 != nullptr )
                    {
// Original IDA line 6604 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                      v752 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                      if ( v752 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v753 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v753 + 48))(v23);
                          }
                        }
                        v10 = v1440;
                        v151 = 443;
                        goto LABEL_2516;
                      }
                      v754 = *(_QWORD *)(v736 + 16);
// Original IDA line 6622 | qword_1803BE540 -> novels.services.agent_session#198 bytes [4310093, 4310101]: "writer"
                      v1453 = (int *)qword_1803BE540;
// Original IDA line 6623 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                      v755 = (int *)sub_18037F0C0(a1, v754, qword_1803BDFC0);
                      if ( v755 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v756 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v756 + 48))(v23);
                          }
                        }
                        if ( *v752 >= 0 )
                        {
                          v8 = (*(_QWORD *)v752)-- == 1;
                          if ( v8 )
                          {
                            v757 = sub_18000CC40(v752);
                            (*(void (__fastcall **)(int *))(v757 + 48))(v752);
                          }
                        }
                        v10 = v1440;
                        v151 = 447;
                        goto LABEL_2516;
                      }
// Original IDA line 6649 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                      v1455 = (int *)sub_18037F0C0(a1, v740, qword_1803BE350);
                      v758 = v1455;
                      if ( v1455 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v759 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v759 + 48))(v23);
                          }
                        }
                        if ( *v752 >= 0 )
                        {
                          v8 = (*(_QWORD *)v752)-- == 1;
                          if ( v8 )
                          {
                            v760 = sub_18000CC40(v752);
                            (*(void (__fastcall **)(int *))(v760 + 48))(v752);
                          }
                        }
                        if ( *v755 >= 0 )
                        {
                          v8 = (*(_QWORD *)v755)-- == 1;
                          if ( v8 )
                          {
                            v761 = sub_18000CC40(v755);
                            (*(void (__fastcall **)(int *))(v761 + 48))(v755);
                          }
                        }
                        v10 = v1440;
                        v151 = 448;
                        goto LABEL_2516;
                      }
                      v1442 = (_QWORD *)sub_1801F3AC0(a1);
                      if ( v1442 == nullptr )
                      {
// Original IDA line 6688 | qword_1803BE548 -> novels.services.agent_session#199 bytes [4310101, 4310123]: "chapter_needs_review"
                        sub_180373340(a1, &v1438, qword_1803BE548);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v762 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v762 + 48))(v23);
                          }
                        }
                        if ( *v752 >= 0 )
                        {
                          v8 = (*(_QWORD *)v752)-- == 1;
                          if ( v8 )
                          {
                            v763 = sub_18000CC40(v752);
                            (*(void (__fastcall **)(int *))(v763 + 48))(v752);
                          }
                        }
                        if ( *v755 >= 0 )
                        {
                          v8 = (*(_QWORD *)v755)-- == 1;
                          if ( v8 )
                          {
                            v764 = sub_18000CC40(v755);
                            (*(void (__fastcall **)(int *))(v764 + 48))(v755);
                          }
                        }
                        if ( *v758 >= 0 )
                        {
                          v8 = (*(_QWORD *)v758)-- == 1;
                          if ( v8 )
                          {
                            v765 = sub_18000CC40(v758);
                            (*(void (__fastcall **)(int *))(v765 + 48))(v758);
                          }
                        }
                        goto LABEL_1432;
                      }
// Original IDA line 6727 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                      v766 = (int *)sub_18037F0C0(a1, v1454, qword_1803BDFD8);
                      if ( v766 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v23 >= 0 )
                        {
                          v8 = (*(_QWORD *)v23)-- == 1;
                          if ( v8 )
                          {
                            v767 = sub_18000CC40(v23);
                            (*(void (__fastcall **)(int *))(v767 + 48))(v23);
                          }
                        }
                        if ( *v752 >= 0 )
                        {
                          v8 = (*(_QWORD *)v752)-- == 1;
                          if ( v8 )
                          {
                            v768 = sub_18000CC40(v752);
                            (*(void (__fastcall **)(int *))(v768 + 48))(v752);
                          }
                        }
                        if ( *v755 >= 0 )
                        {
                          v8 = (*(_QWORD *)v755)-- == 1;
                          if ( v8 )
                          {
                            v769 = sub_18000CC40(v755);
                            (*(void (__fastcall **)(int *))(v769 + 48))(v755);
                          }
                        }
                        if ( *v758 >= 0 )
                        {
                          v8 = (*(_QWORD *)v758)-- == 1;
                          if ( v8 )
                          {
                            v770 = sub_18000CC40(v758);
                            (*(void (__fastcall **)(int *))(v770 + 48))(v758);
                            v10 = v1440;
                            v151 = 449;
                            goto LABEL_2516;
                          }
                        }
LABEL_1432:
                        v10 = v1440;
                        v151 = 449;
                        goto LABEL_2516;
                      }
                      v151 = 449;
                      *(_DWORD *)(v1440 + 40) = 449;
// Original IDA line 6777 | qword_1803BE550 -> novels.services.agent_session#200 bytes [4310123, 4310133]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310125, 4310133], "value": "status"}]}
                      v177 = (int *)sub_1803770D0(a1, v766, qword_1803BE550);
                      if ( *v766 >= 0 )
                      {
                        v8 = (*(_QWORD *)v766)-- == 1;
                        if ( v8 )
                        {
                          v771 = sub_18000CC40(v766);
                          (*(void (__fastcall **)(int *))(v771 + 48))(v766);
                        }
                      }
                      if ( v177 == nullptr )
                        goto LABEL_1450;
                      v776 = sub_180001510(v177);
                      if ( v776 != -1 )
                      {
                        if ( v776 != 1 )
                        {
                          if ( *v177 >= 0 )
                          {
                            v8 = (*(_QWORD *)v177)-- == 1;
                            if ( v8 )
                            {
                              v782 = sub_18000CC40(v177);
                              (*(void (__fastcall **)(int *))(v782 + 48))(v177);
                            }
                          }
// Original IDA line 6803 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          v177 = (int *)qword_1803BDFE8;
// Original IDA line 6804 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 6805 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                            ++*(_DWORD *)qword_1803BDFE8;
                        }
                        v783 = (int *)sub_1803749A0(v177);
                        if ( *v177 >= 0 )
                        {
                          v8 = (*(_QWORD *)v177)-- == 1;
                          if ( v8 )
                          {
                            v784 = sub_18000CC40(v177);
                            (*(void (__fastcall **)(int *))(v784 + 48))(v177);
                          }
                        }
                        v772 = a1;
                        if ( v783 == nullptr )
                          goto LABEL_1451;
                        v785 = v1442;
                        *(_DWORD *)(v1440 + 40) = 449;
                        v177 = (int *)sub_180376770(a1, v785, v783);
                        if ( *v783 >= 0 )
                        {
                          v8 = (*(_QWORD *)v783)-- == 1;
                          if ( v8 )
                          {
                            v786 = sub_18000CC40(v783);
                            (*(void (__fastcall **)(int *))(v786 + 48))(v783);
                          }
                        }
                        if ( v177 == nullptr )
                        {
LABEL_1450:
                          v772 = a1;
LABEL_1451:
                          sub_18000CE30(v772, &v1438);
                          if ( *v23 >= 0 )
                          {
                            v8 = (*(_QWORD *)v23)-- == 1;
                            if ( v8 )
                            {
                              v773 = sub_18000CC40(v23);
                              (*(void (__fastcall **)(int *))(v773 + 48))(v23);
                            }
                          }
                          if ( *v752 >= 0 )
                          {
                            v8 = (*(_QWORD *)v752)-- == 1;
                            if ( v8 )
                            {
                              v774 = sub_18000CC40(v752);
                              (*(void (__fastcall **)(int *))(v774 + 48))(v752);
                            }
                          }
                          if ( *v755 >= 0 )
                          {
                            v8 = (*(_QWORD *)v755)-- == 1;
                            if ( v8 )
                            {
                              v775 = sub_18000CC40(v755);
                              (*(void (__fastcall **)(int *))(v775 + 48))(v755);
                            }
                          }
                          v23 = v1455;
LABEL_575:
                          if ( *v23 >= 0 )
                          {
                            v8 = (*(_QWORD *)v23)-- == 1;
                            if ( v8 )
                              goto LABEL_415;
                          }
                          goto LABEL_277;
                        }
                        v787 = sub_180001510(v177);
                        if ( v787 != -1 )
                        {
                          if ( *v177 >= 0 )
                          {
                            v8 = (*(_QWORD *)v177)-- == 1;
                            if ( v8 )
                            {
                              v788 = sub_18000CC40(v177);
                              (*(void (__fastcall **)(int *))(v788 + 48))(v177);
                            }
                          }
// Original IDA line 6887 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"
                          v789 = qword_1803BE2E0;
                          if ( v787 != 0 )
// Original IDA line 6889 | qword_1803BE558 -> novels.services.agent_session#201 bytes [4310133, 4310140]: "error"
                            v789 = qword_1803BE558;
                          v1439 = (int *)v789;
// Original IDA line 6891 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                          v790 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                          v791 = (int *)v790;
                          v792 = a1;
                          if ( v790 != 0 )
                          {
// Original IDA line 6896 | qword_1803BE560 -> novels.services.agent_session#202 bytes [4310140, 4310154]: "STAGES_WRITE"
                            v794 = (int *)sub_18037F0C0(a1, v790, qword_1803BE560);
                            if ( *v791 >= 0 )
                            {
                              v8 = (*(_QWORD *)v791)-- == 1;
                              if ( v8 )
                              {
                                v795 = sub_18000CC40(v791);
                                (*(void (__fastcall **)(int *))(v795 + 48))(v791);
                              }
                            }
                            v792 = a1;
                            if ( v794 != nullptr )
                            {
                              v796 = sub_180001930(a1, v794);
                              v29 = *v794 < 0;
                              v797 = (int *)v796;
                              v1449 = (int *)v796;
                              if ( !v29 )
                              {
                                v8 = (*(_QWORD *)v794)-- == 1;
                                if ( v8 )
                                {
                                  v798 = sub_18000CC40(v794);
                                  (*(void (__fastcall **)(int *))(v798 + 48))(v794);
                                }
                              }
                              v792 = a1;
                              if ( v797 != nullptr )
                              {
                                v1446 = (_QWORD *)sub_180372160(a1, 0);
                                v799 = v1446;
                                for ( k = (int *)sub_18000D430(v797); k != nullptr; k = (int *)sub_18000D430(v1449) )
                                {
                                  v801 = v1448;
                                  v802 = v1444;
                                  v803 = v1448;
                                  v1448 = k;
                                  v804 = v1444;
                                  if ( v803 != nullptr && *v801 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v801)-- == 1;
                                    if ( v8 )
                                    {
                                      v805 = sub_18000CC40(v801);
                                      (*(void (__fastcall **)(int *))(v805 + 48))(v803);
                                      v802 = v804;
                                    }
                                  }
                                  v806 = *k;
                                  v807 = v802;
                                  v1444 = k;
                                  v808 = v806 + 1;
                                  if ( v808 != 0 )
                                    *k = v808;
                                  if ( v802 != nullptr && *v804 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v804)-- == 1;
                                    if ( v8 )
                                    {
                                      v809 = sub_18000CC40(v802);
                                      (*(void (__fastcall **)(int *))(v809 + 48))(v807);
                                    }
                                  }
// Original IDA line 6959 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                                  v810 = qword_1803BE378;
                                  v811 = PyDict_NewPresized(2);
                                  PyDict_SetItem(v811, v810, k);
// Original IDA line 6962 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"; qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                                  PyDict_SetItem(v811, qword_1803BE380, qword_1803BE2E0);
                                  v812 = v811;
                                  v799 = v1446;
                                  if ( (unsigned __int8)sub_180372560(v1446, v812) == 0
                                    || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                  {
                                    goto LABEL_1596;
                                  }
                                  v797 = v1449;
                                }
                                if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
                                {
LABEL_1596:
                                  sub_18000CE30(a1, &v1438);
                                  v855 = v1438;
                                  sub_18000CE70(&v1438);
                                  v856 = v1449;
                                  if ( *v1449 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1449)-- == 1;
                                    if ( v8 )
                                    {
                                      v857 = sub_18000CC40(v856);
                                      (*(void (__fastcall **)(int *))(v857 + 48))(v856);
                                    }
                                  }
                                  v858 = v1446;
                                  if ( *(int *)v1446 >= 0 )
                                  {
                                    v8 = (*v1446)-- == 1;
                                    if ( v8 )
                                    {
                                      v859 = sub_18000CC40(v858);
                                      (*(void (__fastcall **)(_QWORD *))(v859 + 48))(v858);
                                    }
                                  }
                                  v860 = (int **)v1444;
                                  if ( v1444 != nullptr && *v1444 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1444)-- == 1;
                                    v1448 = *v860;
                                    if ( v8 )
                                    {
                                      v861 = sub_18000CC40(v860);
                                      (*(void (__fastcall **)(int **))(v861 + 48))(v860);
                                    }
                                  }
                                  v1438 = v855;
                                  sub_18000CE70(&v1438);
                                  if ( v860 != nullptr && *(int *)v860 >= 0 )
                                  {
                                    v8 = *v860 == (int *)1;
                                    *v860 = (int *)((char *)*v860 - 1);
                                    v1448 = *v860;
                                    if ( v8 )
                                    {
                                      v862 = sub_18000CC40(v860);
                                      (*(void (__fastcall **)(int **))(v862 + 48))(v860);
                                    }
                                  }
                                  v1438 = v855;
                                  v10 = v1440;
                                  v151 = 450;
                                  goto LABEL_2516;
                                }
                                if ( *v799 != -1 )
                                  ++*v799;
                                if ( *v797 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v797)-- == 1;
                                  if ( v8 )
                                  {
                                    v813 = sub_18000CC40(v797);
                                    (*(void (__fastcall **)(int *))(v813 + 48))(v797);
                                  }
                                }
                                if ( (int)*v799 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v799)-- == 1;
                                  if ( v8 )
                                  {
                                    v814 = sub_18000CC40(v799);
                                    (*(void (__fastcall **)(_DWORD *))(v814 + 48))(v799);
                                  }
                                }
                                v815 = v1444;
                                if ( v1444 != nullptr && *v1444 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1444)-- == 1;
                                  v816 = *(int **)v815;
                                  if ( v8 )
                                  {
                                    v817 = sub_18000CC40(v815);
                                    (*(void (__fastcall **)(int *))(v817 + 48))(v815);
                                    LODWORD(v816) = *v815;
                                  }
                                  if ( (int)v816 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v815)-- == 1;
                                    v1448 = *(int **)v815;
                                    if ( v8 )
                                    {
                                      v818 = sub_18000CC40(v815);
                                      (*(void (__fastcall **)(int *))(v818 + 48))(v815);
                                    }
                                  }
                                }
                                v819 = v1455;
                                v151 = 443;
                                *(_DWORD *)(v1440 + 40) = 443;
                                v1516[0] = v1453;
                                v1516[1] = v1450;
                                v1534[0] = v1454;
                                v1534[3] = v1439;
                                v1534[4] = v1446;
                                v1534[1] = v755;
                                v1534[2] = v819;
// Original IDA line 7079 | qword_1803BE568 -> novels.services.agent_session#203 bytes [4310154, 4310200]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310156, 4310165], "value": "details"}, {"tag": "a", "fileRange": [4310165, 4310173], "value": "bookId"}, {"tag": "a", "fileRange": [4310173, 4310184], "value": "bookTitle"}, {"tag": "a... (truncated literal)
                                v820 = (int *)sub_18037D650(a1, v752, v1516, v1534, qword_1803BE568);
                                if ( *v752 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v752)-- == 1;
                                  if ( v8 )
                                  {
                                    v821 = sub_18000CC40(v752);
                                    (*(void (__fastcall **)(int *))(v821 + 48))(v752);
                                  }
                                }
                                if ( *v755 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v755)-- == 1;
                                  if ( v8 )
                                  {
                                    v822 = sub_18000CC40(v755);
                                    (*(void (__fastcall **)(int *))(v822 + 48))(v755);
                                  }
                                }
                                if ( *v819 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v819)-- == 1;
                                  if ( v8 )
                                  {
                                    v823 = sub_18000CC40(v819);
                                    (*(void (__fastcall **)(int *))(v823 + 48))(v819);
                                  }
                                }
                                v824 = v1446;
                                if ( *(int *)v1446 >= 0 )
                                {
                                  v8 = (*v1446)-- == 1;
                                  if ( v8 )
                                  {
                                    v825 = sub_18000CC40(v824);
                                    (*(void (__fastcall **)(_QWORD *))(v825 + 48))(v824);
                                  }
                                }
                                if ( v820 == nullptr )
                                {
LABEL_2201:
                                  sub_18000CE30(a1, &v1438);
                                  goto LABEL_575;
                                }
                                *(_DWORD *)(v1440 + 40) = 443;
                                v826 = (int *)sub_180376770(a1, v23, v820);
                                if ( *v23 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v23)-- == 1;
                                  if ( v8 )
                                  {
                                    v827 = sub_18000CC40(v23);
                                    (*(void (__fastcall **)(int *))(v827 + 48))(v23);
                                  }
                                }
                                if ( *v820 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v820)-- == 1;
                                  if ( v8 )
                                  {
                                    v828 = sub_18000CC40(v820);
                                    (*(void (__fastcall **)(int *))(v828 + 48))(v820);
                                  }
                                }
                                if ( v826 == nullptr )
                                  goto LABEL_2465;
                                if ( *v826 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v826)-- == 1;
                                  if ( v8 )
                                  {
                                    v829 = sub_18000CC40(v826);
                                    (*(void (__fastcall **)(int *))(v829 + 48))(v826);
                                  }
                                }
// Original IDA line 7154 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                                v830 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                                v831 = (int *)v830;
                                if ( v830 == 0 )
                                {
                                  sub_18000CE30(a1, &v1438);
                                  v10 = v1440;
                                  v151 = 455;
                                  goto LABEL_2516;
                                }
                                v832 = v1445;
                                v151 = 455;
                                v10 = v1440;
                                v833 = *(_QWORD *)(v1445 + 16);
                                *(_DWORD *)(v1440 + 40) = 455;
// Original IDA line 7168 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                                v834 = (int *)sub_18037E6C0(a1, v830, qword_1803BE0D8, v833);
                                if ( *v831 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v831)-- == 1;
                                  if ( v8 )
                                  {
                                    v835 = sub_18000CC40(v831);
                                    (*(void (__fastcall **)(int *))(v835 + 48))(v831);
                                  }
                                }
                                if ( v834 == nullptr )
                                  goto LABEL_2475;
                                if ( *v834 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v834)-- == 1;
                                  if ( v8 )
                                  {
                                    v836 = sub_18000CC40(v834);
                                    (*(void (__fastcall **)(int *))(v836 + 48))(v834);
                                  }
                                }
// Original IDA line 7189 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                                v837 = qword_1803BE2C8;
                                v23 = (int *)PyDict_NewPresized(2);
                                PyDict_SetItem(v23, v837, Py_TrueStruct);
// Original IDA line 7192 | qword_1803BE570 -> novels.services.agent_session#204 bytes [4310200, 4310209]: "chapter"
                                v838 = qword_1803BE570;
// Original IDA line 7193 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                                v1448 = (int *)qword_1803BE2F8;
                                v70 = (int *)PyDict_NewPresized(3);
                                PyDict_SetItem(v70, v838, v1454);
// Original IDA line 7196 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
                                v1442 = (_QWORD *)qword_1803BE348;
// Original IDA line 7197 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                                v839 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                                v840 = (int *)v839;
                                if ( v839 != 0 )
                                {
// Original IDA line 7201 | qword_1803BDFC8 -> novels.services.agent_session#23 bytes [4307250, 4307268]: "get_book_payload"
                                  v841 = (int *)sub_18037F0C0(a1, v839, qword_1803BDFC8);
                                  if ( *v840 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v840)-- == 1;
                                    if ( v8 )
                                    {
                                      v842 = sub_18000CC40(v840);
                                      (*(void (__fastcall **)(int *))(v842 + 48))(v840);
                                    }
                                  }
                                  if ( v841 != nullptr )
                                  {
// Original IDA line 7213 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                                    v843 = sub_18037F0C0(a1, *(_QWORD *)(v832 + 16), qword_1803BDFC0);
                                    v844 = (int *)v843;
                                    if ( v843 == 0 )
                                    {
                                      sub_18000CE30(a1, &v1438);
                                      if ( *v841 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v841)-- == 1;
                                        if ( v8 )
                                        {
                                          v845 = sub_18000CC40(v841);
                                          (*(void (__fastcall **)(int *))(v845 + 48))(v841);
                                        }
                                      }
                                      v151 = 460;
                                      goto LABEL_1176;
                                    }
                                    v151 = 460;
                                    *(_DWORD *)(v10 + 40) = 460;
                                    v846 = (int *)sub_180376770(a1, v841, v843);
                                    if ( *v841 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v841)-- == 1;
                                      if ( v8 )
                                      {
                                        v847 = sub_18000CC40(v841);
                                        (*(void (__fastcall **)(int *))(v847 + 48))(v841);
                                      }
                                    }
                                    if ( *v844 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v844)-- == 1;
                                      if ( v8 )
                                      {
                                        v848 = sub_18000CC40(v844);
                                        (*(void (__fastcall **)(int *))(v848 + 48))(v844);
                                      }
                                    }
                                    if ( v846 == nullptr )
                                    {
LABEL_1586:
                                      sub_18000CE30(a1, &v1438);
                                      v10 = v1440;
                                      goto LABEL_1176;
                                    }
                                    PyDict_SetItem(v70, v1442, v846);
                                    if ( *v846 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v846)-- == 1;
                                      if ( v8 )
                                      {
                                        v849 = sub_18000CC40(v846);
                                        (*(void (__fastcall **)(int *))(v849 + 48))(v846);
                                      }
                                    }
                                    v151 = 461;
LABEL_1591:
// Original IDA line 7270 | qword_1803BE1B8 -> novels.services.agent_session#85 bytes [4308196, 4308205]: "session"
                                    v850 = qword_1803BE1B8;
                                    v10 = v1440;
                                    v851 = *(_QWORD *)(v1445 + 16);
                                    *(_DWORD *)(v1440 + 40) = v151;
// Original IDA line 7274 | qword_1803BE278 -> novels.services.agent_session#109 bytes [4308547, 4308556]: "to_dict"
                                    v852 = sub_18037E560(a1, v851, qword_1803BE278);
                                    v853 = (int *)v852;
                                    if ( v852 != 0 )
                                    {
                                      PyDict_SetItem(v70, v850, v852);
                                      if ( *v853 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v853)-- == 1;
                                        if ( v8 )
                                        {
                                          v854 = sub_18000CC40(v853);
                                          (*(void (__fastcall **)(int *))(v854 + 48))(v853);
                                        }
                                      }
                                      v93 = v1448;
                                      goto LABEL_146;
                                    }
                                    goto LABEL_1175;
                                  }
                                }
                                sub_18000CE30(a1, &v1438);
                                v151 = 460;
                                goto LABEL_1176;
                              }
                            }
                          }
                          sub_18000CE30(v792, &v1438);
                          v793 = v1438;
                          sub_18000CE70(&v1438);
                          v1438 = v793;
                          v10 = v1440;
                          v151 = 450;
                          goto LABEL_2516;
                        }
                      }
                      sub_18000CE30(a1, &v1438);
                      if ( *v23 >= 0 )
                      {
                        v8 = (*(_QWORD *)v23)-- == 1;
                        if ( v8 )
                        {
                          v777 = sub_18000CC40(v23);
                          (*(void (__fastcall **)(int *))(v777 + 48))(v23);
                        }
                      }
                      if ( *v752 >= 0 )
                      {
                        v8 = (*(_QWORD *)v752)-- == 1;
                        if ( v8 )
                        {
                          v778 = sub_18000CC40(v752);
                          (*(void (__fastcall **)(int *))(v778 + 48))(v752);
                        }
                      }
                      if ( *v755 >= 0 )
                      {
                        v8 = (*(_QWORD *)v755)-- == 1;
                        if ( v8 )
                        {
                          v779 = sub_18000CC40(v755);
                          (*(void (__fastcall **)(int *))(v779 + 48))(v755);
                        }
                      }
                      v780 = v1455;
                      if ( *v1455 >= 0 )
                      {
                        v8 = (*(_QWORD *)v1455)-- == 1;
                        if ( v8 )
                        {
                          v781 = sub_18000CC40(v780);
                          (*(void (__fastcall **)(int *))(v781 + 48))(v780);
                        }
                      }
LABEL_567:
                      if ( *v177 >= 0 )
                      {
                        v8 = (*(_QWORD *)v177)-- == 1;
                        if ( v8 )
                        {
LABEL_318:
                          v185 = sub_18000CC40(v177);
                          (*(void (__fastcall **)(int *))(v185 + 48))(v177);
                          v10 = v1440;
                          goto LABEL_2516;
                        }
                      }
                      goto LABEL_277;
                    }
                  }
                  sub_18000CE30(a1, &v1438);
                  v10 = v1440;
                  v151 = 443;
                  goto LABEL_2516;
                }
              }
              sub_18000CE30(a1, &v1438);
              v151 = 440;
              goto LABEL_2516;
            }
          }
          sub_18000CE30(a1, &v1438);
          v151 = 439;
          goto LABEL_2516;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 437;
      goto LABEL_2516;
    }
// Original IDA line 7383 | qword_1803BE578 -> novels.services.agent_session#205 bytes [4310209, 4310224]: "audit_chapter"
    v863 = sub_1803836D0(v45, qword_1803BE578);
    if ( v863 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 465;
      goto LABEL_2516;
    }
    if ( v863 == 1 )
    {
// Original IDA line 7392 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
      v864 = sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
      v865 = (int *)v864;
      if ( v864 != 0 )
      {
        v866 = sub_180001510(v864);
        if ( *v865 >= 0 )
        {
          v8 = (*(_QWORD *)v865)-- == 1;
          if ( v8 )
          {
            v867 = sub_18000CC40(v865);
            (*(void (__fastcall **)(int *))(v867 + 48))(v865);
          }
        }
        if ( v866 != -1 )
        {
          if ( v866 != 0 )
          {
// Original IDA line 7410 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
            v868 = sub_18037F0C0(a1, v5, qword_1803BDFD8);
            v869 = (int *)v868;
            if ( v868 == 0 )
            {
              sub_18000CE30(a1, &v1438);
              v151 = 468;
              goto LABEL_2516;
            }
            v151 = 468;
            *(_DWORD *)(v10 + 40) = 468;
// Original IDA line 7420 | qword_1803BE580 -> novels.services.agent_session#206 bytes [4310224, 4310241]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310226, 4310241], "value": "chapterNumber"}]}
            v870 = sub_1803770D0(a1, v868, qword_1803BE580);
            if ( *v869 >= 0 )
            {
              v8 = (*(_QWORD *)v869)-- == 1;
              if ( v8 )
              {
                v871 = sub_18000CC40(v869);
                (*(void (__fastcall **)(int *))(v871 + 48))(v869);
              }
            }
            if ( v870 == 0 )
              goto LABEL_2475;
            v1468 = (int *)v870;
// Original IDA line 7433 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
            v872 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
            v873 = (int *)v872;
            if ( v872 != 0 )
            {
// Original IDA line 7437 | qword_1803BE578 -> novels.services.agent_session#205 bytes [4310209, 4310224]: "audit_chapter"
              v874 = (int *)sub_18037F0C0(a1, v872, qword_1803BE578);
              if ( *v873 >= 0 )
              {
                v8 = (*(_QWORD *)v873)-- == 1;
                if ( v8 )
                {
                  v875 = sub_18000CC40(v873);
                  (*(void (__fastcall **)(int *))(v875 + 48))(v873);
                }
              }
              if ( v874 != nullptr )
              {
// Original IDA line 7449 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                v876 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
                if ( v876 == nullptr )
                {
                  sub_18000CE30(a1, &v1438);
                  if ( *v874 >= 0 )
                  {
                    v8 = (*(_QWORD *)v874)-- == 1;
                    if ( v8 )
                    {
                      v877 = sub_18000CC40(v874);
                      (*(void (__fastcall **)(int *))(v877 + 48))(v874);
                    }
                  }
                  v151 = 470;
                  goto LABEL_2516;
                }
                v878 = sub_180001510(v870);
                if ( v878 == -1 )
                  goto LABEL_1639;
                if ( v878 != 0 )
                {
                  v881 = (int *)PyNumber_Long(v870);
                  if ( v881 == nullptr )
                  {
LABEL_1639:
                    sub_18000CE30(a1, &v1438);
                    if ( *v874 >= 0 )
                    {
                      v8 = (*(_QWORD *)v874)-- == 1;
                      if ( v8 )
                      {
                        v879 = sub_18000CC40(v874);
                        (*(void (__fastcall **)(int *))(v879 + 48))(v874);
                      }
                    }
                    if ( *v876 >= 0 )
                    {
                      v8 = (*(_QWORD *)v876)-- == 1;
                      if ( v8 )
                      {
                        v880 = sub_18000CC40(v876);
                        (*(void (__fastcall **)(int *))(v880 + 48))(v876);
                      }
                    }
                    v151 = 471;
                    goto LABEL_2516;
                  }
                }
                else
                {
                  v881 = (int *)Py_NoneStruct;
                  if ( Py_NoneStruct != -1 )
                    ++Py_NoneStruct;
                }
                v882 = v1456;
                v151 = 469;
                *(_DWORD *)(v10 + 40) = 469;
                v1503 = v882;
                v1517[0] = v876;
                v1517[1] = v881;
// Original IDA line 7509 | qword_1803BE340 -> novels.services.agent_session#134 bytes [4309052, 4309061]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309054, 4309061], "value": "model"}]}
                v883 = sub_18037D650(a1, v874, v1517, &v1503, qword_1803BE340);
                v29 = *v874 < 0;
                v884 = (int *)v883;
                v1454 = (_QWORD *)v883;
                if ( !v29 )
                {
                  v8 = (*(_QWORD *)v874)-- == 1;
                  if ( v8 )
                  {
                    v885 = sub_18000CC40(v874);
                    (*(void (__fastcall **)(int *))(v885 + 48))(v874);
                  }
                }
                if ( *v876 >= 0 )
                {
                  v8 = (*(_QWORD *)v876)-- == 1;
                  if ( v8 )
                  {
                    v886 = sub_18000CC40(v876);
                    (*(void (__fastcall **)(int *))(v886 + 48))(v876);
                  }
                }
                if ( *v881 >= 0 )
                {
                  v8 = (*(_QWORD *)v881)-- == 1;
                  if ( v8 )
                  {
                    v887 = sub_18000CC40(v881);
                    (*(void (__fastcall **)(int *))(v887 + 48))(v881);
                  }
                }
                if ( v884 == nullptr )
                  goto LABEL_2475;
                v1457 = v884;
// Original IDA line 7543 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                v888 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                v889 = (int *)v888;
                if ( v888 != 0 )
                {
// Original IDA line 7547 | qword_1803BE530 -> novels.services.agent_session#196 bytes [4310053, 4310064]: "load_book"
                  v890 = (int *)sub_18037F0C0(a1, v888, qword_1803BE530);
                  if ( *v889 >= 0 )
                  {
                    v8 = (*(_QWORD *)v889)-- == 1;
                    if ( v8 )
                    {
                      v891 = sub_18000CC40(v889);
                      (*(void (__fastcall **)(int *))(v891 + 48))(v889);
                    }
                  }
                  if ( v890 != nullptr )
                  {
// Original IDA line 7559 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                    v892 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                    v893 = (int *)v892;
                    if ( v892 == 0 )
                    {
                      sub_18000CE30(a1, &v1438);
                      if ( *v890 >= 0 )
                      {
                        v8 = (*(_QWORD *)v890)-- == 1;
                        if ( v8 )
                        {
                          v894 = sub_18000CC40(v890);
                          (*(void (__fastcall **)(int *))(v894 + 48))(v890);
                        }
                      }
                      v151 = 474;
                      goto LABEL_2516;
                    }
                    v151 = 474;
                    *(_DWORD *)(v10 + 40) = 474;
                    v895 = sub_180376770(a1, v890, v892);
                    if ( *v890 >= 0 )
                    {
                      v8 = (*(_QWORD *)v890)-- == 1;
                      if ( v8 )
                      {
                        v896 = sub_18000CC40(v890);
                        (*(void (__fastcall **)(int *))(v896 + 48))(v890);
                      }
                    }
                    if ( *v893 >= 0 )
                    {
                      v8 = (*(_QWORD *)v893)-- == 1;
                      if ( v8 )
                      {
                        v897 = sub_18000CC40(v893);
                        (*(void (__fastcall **)(int *))(v897 + 48))(v893);
                      }
                    }
                    v310 = a1;
                    if ( v895 == 0 )
                      goto LABEL_2060;
                    v1461 = (int *)v895;
// Original IDA line 7601 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                    v898 = sub_18037F0C0(a1, v884, qword_1803BDFD8);
                    v899 = (int *)v898;
                    if ( v898 == 0 )
                    {
                      sub_18000CE30(a1, &v1438);
                      v10 = v1440;
                      v151 = 475;
                      goto LABEL_2516;
                    }
                    v900 = v1440;
                    v151 = 475;
                    *(_DWORD *)(v1440 + 40) = 475;
// Original IDA line 7613 | qword_1803BE588 -> novels.services.agent_session#207 bytes [4310241, 4310250]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310243, 4310250], "value": "audit"}]}
                    v901 = (int *)sub_1803770D0(a1, v898, qword_1803BE588);
                    if ( *v899 >= 0 )
                    {
                      v8 = (*(_QWORD *)v899)-- == 1;
                      if ( v8 )
                      {
                        v902 = sub_18000CC40(v899);
                        (*(void (__fastcall **)(int *))(v902 + 48))(v899);
                      }
                    }
                    if ( v901 == nullptr )
                    {
                      sub_18000CE30(a1, &v1438);
                      goto LABEL_1688;
                    }
                    v903 = sub_180001510(v901);
                    if ( v903 != -1 )
                    {
                      if ( v903 == 1 )
                      {
                        v905 = (__int64)v901;
                      }
                      else
                      {
                        if ( *v901 >= 0 )
                        {
                          v8 = (*(_QWORD *)v901)-- == 1;
                          if ( v8 )
                          {
                            v906 = sub_18000CC40(v901);
                            (*(void (__fastcall **)(int *))(v906 + 48))(v901);
                          }
                        }
                        v905 = sub_180370EC0(a1);
                      }
                      v1469 = (int *)v905;
// Original IDA line 7649 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                      v177 = (int *)sub_18037F0C0(a1, v884, qword_1803BDFD8);
                      if ( v177 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        v151 = 476;
                        v10 = v900;
                        goto LABEL_2516;
                      }
// Original IDA line 7657 | qword_1803BE590 -> novels.services.agent_session#208 bytes [4310250, 4310258]: "passed"
                      v907 = qword_1803BE590;
// Original IDA line 7658 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                      v908 = sub_18037F0C0(a1, v905, qword_1803BDFD8);
                      v909 = (int *)v908;
                      if ( v908 == 0 )
                      {
                        sub_18000CE30(a1, &v1438);
                        if ( *v177 >= 0 )
                        {
                          v8 = (*(_QWORD *)v177)-- == 1;
                          if ( v8 )
                          {
                            v910 = sub_18000CC40(v177);
                            (*(void (__fastcall **)(int *))(v910 + 48))(v177);
                          }
                        }
                        v151 = 476;
                        v10 = v900;
                        goto LABEL_2516;
                      }
                      v151 = 476;
                      *(_DWORD *)(v900 + 40) = 476;
// Original IDA line 7678 | qword_1803BE598 -> novels.services.agent_session#209 bytes [4310258, 4310268]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310260, 4310268], "value": "passed"}]}
                      v911 = (int *)sub_1803770D0(a1, v908, qword_1803BE598);
                      if ( *v909 >= 0 )
                      {
                        v8 = (*(_QWORD *)v909)-- == 1;
                        if ( v8 )
                        {
                          v912 = sub_18000CC40(v909);
                          (*(void (__fastcall **)(int *))(v912 + 48))(v909);
                        }
                      }
                      v314 = a1;
                      if ( v911 == nullptr )
                        goto LABEL_566;
                      *(_DWORD *)(v1440 + 40) = 476;
                      v1518[0] = v907;
                      v1518[1] = v911;
                      v913 = (int *)sub_180377750(a1, v177, v1518);
                      if ( *v177 >= 0 )
                      {
                        v8 = (*(_QWORD *)v177)-- == 1;
                        if ( v8 )
                        {
                          v914 = sub_18000CC40(v177);
                          (*(void (__fastcall **)(int *))(v914 + 48))(v177);
                        }
                      }
                      if ( *v911 >= 0 )
                      {
                        v8 = (*(_QWORD *)v911)-- == 1;
                        if ( v8 )
                        {
                          v915 = sub_18000CC40(v911);
                          (*(void (__fastcall **)(int *))(v915 + 48))(v911);
                        }
                      }
                      if ( v913 == nullptr )
                        goto LABEL_2465;
                      v916 = sub_180001510(v913);
                      if ( *v913 >= 0 )
                      {
                        v8 = (*(_QWORD *)v913)-- == 1;
                        if ( v8 )
                        {
                          v917 = sub_18000CC40(v913);
                          (*(void (__fastcall **)(int *))(v917 + 48))(v913);
                        }
                      }
                      if ( v916 == -1 )
                        goto LABEL_2465;
                      v918 = (_DWORD *)Py_FalseStruct;
                      if ( v916 != 0 )
                        v918 = (_DWORD *)Py_TrueStruct;
                      v1464 = v918;
                      if ( *v918 != -1 )
                        ++*v918;
                      v919 = v1454;
// Original IDA line 7734 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                      v920 = (int *)sub_18037F0C0(a1, v1454, qword_1803BDFD8);
                      if ( v920 == nullptr )
                      {
                        sub_18000CE30(a1, &v1438);
                        v10 = v1440;
                        v151 = 477;
                        goto LABEL_2516;
                      }
                      v151 = 477;
                      *(_DWORD *)(v1440 + 40) = 477;
// Original IDA line 7744 | qword_1803BE3E0 -> novels.services.agent_session#154 bytes [4309362, 4309373]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309364, 4309373], "value": "summary"}]}
                      v177 = (int *)sub_1803770D0(a1, v920, qword_1803BE3E0);
                      if ( *v920 >= 0 )
                      {
                        v8 = (*(_QWORD *)v920)-- == 1;
                        if ( v8 )
                        {
                          v921 = sub_18000CC40(v920);
                          (*(void (__fastcall **)(int *))(v921 + 48))(v920);
                        }
                      }
                      if ( v177 == nullptr )
                        goto LABEL_2465;
                      v922 = sub_180001510(v177);
                      if ( v922 == -1 )
                        goto LABEL_565;
                      if ( v922 != 1 )
                      {
                        if ( *v177 >= 0 )
                        {
                          v8 = (*(_QWORD *)v177)-- == 1;
                          if ( v8 )
                          {
                            v923 = sub_18000CC40(v177);
                            (*(void (__fastcall **)(int *))(v923 + 48))(v177);
                          }
                        }
                        v924 = sub_1801F3790(a1);
                        if ( v924 == 0 )
                        {
// Original IDA line 7773 | qword_1803BE5A0 -> novels.services.agent_session#210 bytes [4310268, 4310289]: "build_audit_summary"
                          sub_180373340(a1, &v1438, qword_1803BE5A0);
                          v10 = v1440;
                          goto LABEL_2516;
                        }
// Original IDA line 7777 | qword_1803BE5A8 -> novels.services.agent_session#211 bytes [4310289, 4310304]: "chapterNumber"
                        v925 = sub_180001B80(a1, v919, qword_1803BE5A8);
                        v926 = (int *)v925;
                        if ( v925 == 0 )
                          goto LABEL_2465;
                        v927 = (int *)PyNumber_Long(v925);
                        if ( *v926 >= 0 )
                        {
                          v8 = (*(_QWORD *)v926)-- == 1;
                          if ( v8 )
                          {
                            v928 = sub_18000CC40(v926);
                            (*(void (__fastcall **)(int *))(v928 + 48))(v926);
                          }
                        }
                        v310 = a1;
                        if ( v927 == nullptr )
                          goto LABEL_2060;
                        *(_DWORD *)(v1440 + 40) = 477;
                        v1504 = v1464;
                        v1505 = v927;
// Original IDA line 7797 | qword_1803BE598 -> novels.services.agent_session#209 bytes [4310258, 4310268]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310260, 4310268], "value": "passed"}]}
                        v177 = (int *)sub_18037D030(a1, v924, &v1505, &v1504, qword_1803BE598);
                        if ( *v927 >= 0 )
                        {
                          v8 = (*(_QWORD *)v927)-- == 1;
                          if ( v8 )
                          {
                            v929 = sub_18000CC40(v927);
                            (*(void (__fastcall **)(int *))(v929 + 48))(v927);
                          }
                        }
                        if ( v177 == nullptr )
                          goto LABEL_2465;
                        v918 = v1464;
                      }
                      v930 = sub_1803749A0(v177);
                      v29 = *v177 < 0;
                      v1448 = (int *)v930;
                      if ( !v29 )
                      {
                        v8 = (*(_QWORD *)v177)-- == 1;
                        if ( v8 )
                        {
                          v931 = sub_18000CC40(v177);
                          (*(void (__fastcall **)(int *))(v931 + 48))(v177);
                          v930 = (__int64)v1448;
                        }
                      }
                      v310 = a1;
                      if ( v930 != 0 )
                      {
                        v1460 = (int *)v930;
// Original IDA line 7828 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                        if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
                        {
                          sub_18000CE30(a1, &v1438);
                          v10 = v1440;
                          v151 = 478;
                          goto LABEL_2516;
                        }
// Original IDA line 7835 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                        v932 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
                        v933 = (int *)v932;
                        if ( v932 != 0 )
                        {
// Original IDA line 7839 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                          v23 = (int *)sub_18037F0C0(a1, v932, qword_1803BE0B8);
                          if ( *v933 >= 0 )
                          {
                            v8 = (*(_QWORD *)v933)-- == 1;
                            if ( v8 )
                            {
                              v934 = sub_18000CC40(v933);
                              (*(void (__fastcall **)(int *))(v934 + 48))(v933);
                            }
                          }
                          if ( v23 != nullptr )
                          {
// Original IDA line 7851 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                            v1446 = (_QWORD *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                            v935 = (int *)v1446;
                            if ( v1446 == nullptr )
                            {
                              sub_18000CE30(a1, &v1438);
                              if ( *v23 >= 0 )
                              {
                                v8 = (*(_QWORD *)v23)-- == 1;
                                if ( v8 )
                                {
                                  v936 = sub_18000CC40(v23);
                                  (*(void (__fastcall **)(int *))(v936 + 48))(v23);
                                }
                              }
                              v10 = v1440;
                              v151 = 479;
                              goto LABEL_2516;
                            }
// Original IDA line 7869 | qword_1803BE5B0 -> novels.services.agent_session#212 bytes [4310304, 4310313]: "auditor"
                            v1439 = (int *)qword_1803BE5B0;
// Original IDA line 7870 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                            v1449 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                            v937 = v1449;
                            if ( v1449 == nullptr )
                            {
                              sub_18000CE30(a1, &v1438);
                              if ( *v23 >= 0 )
                              {
                                v8 = (*(_QWORD *)v23)-- == 1;
                                if ( v8 )
                                {
                                  v938 = sub_18000CC40(v23);
                                  (*(void (__fastcall **)(int *))(v938 + 48))(v23);
                                }
                              }
                              if ( *v935 >= 0 )
                              {
                                v8 = (*(_QWORD *)v935)-- == 1;
                                if ( v8 )
                                {
                                  v939 = sub_18000CC40(v935);
                                  (*(void (__fastcall **)(int *))(v939 + 48))(v935);
                                }
                              }
                              v10 = v1440;
                              v151 = 482;
                              goto LABEL_2516;
                            }
// Original IDA line 7897 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                            v1450 = (_QWORD *)sub_18037F0C0(a1, v895, qword_1803BE350);
                            if ( v1450 == nullptr )
                            {
                              sub_18000CE30(a1, &v1438);
                              if ( *v23 >= 0 )
                              {
                                v8 = (*(_QWORD *)v23)-- == 1;
                                if ( v8 )
                                {
                                  v940 = sub_18000CC40(v23);
                                  (*(void (__fastcall **)(int *))(v940 + 48))(v23);
                                }
                              }
                              if ( *v935 >= 0 )
                              {
                                v8 = (*(_QWORD *)v935)-- == 1;
                                if ( v8 )
                                {
                                  v941 = sub_18000CC40(v935);
                                  (*(void (__fastcall **)(int *))(v941 + 48))(v935);
                                }
                              }
                              if ( *v937 >= 0 )
                              {
                                v8 = (*(_QWORD *)v937)-- == 1;
                                if ( v8 )
                                {
                                  v942 = sub_18000CC40(v937);
                                  (*(void (__fastcall **)(int *))(v942 + 48))(v937);
                                }
                              }
                              v10 = v1440;
                              v151 = 483;
                              goto LABEL_2516;
                            }
                            v943 = sub_180001510(v918);
// Original IDA line 7933 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"
                            v944 = qword_1803BE2E0;
                            if ( v943 == 0 )
// Original IDA line 7935 | qword_1803BE558 -> novels.services.agent_session#201 bytes [4310133, 4310140]: "error"
                              v944 = qword_1803BE558;
                            v1453 = (int *)v944;
// Original IDA line 7937 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                            v945 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                            v946 = (int *)v945;
                            v947 = a1;
                            if ( v945 != 0 )
                            {
// Original IDA line 7942 | qword_1803BE5B8 -> novels.services.agent_session#213 bytes [4310313, 4310327]: "STAGES_AUDIT"
                              v949 = (int *)sub_18037F0C0(a1, v945, qword_1803BE5B8);
                              if ( *v946 >= 0 )
                              {
                                v8 = (*(_QWORD *)v946)-- == 1;
                                if ( v8 )
                                {
                                  v950 = sub_18000CC40(v946);
                                  (*(void (__fastcall **)(int *))(v950 + 48))(v946);
                                }
                              }
                              v947 = a1;
                              if ( v949 != nullptr )
                              {
                                v951 = (int *)sub_180001930(a1, v949);
                                if ( *v949 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v949)-- == 1;
                                  if ( v8 )
                                  {
                                    v952 = sub_18000CC40(v949);
                                    (*(void (__fastcall **)(int *))(v952 + 48))(v949);
                                  }
                                }
                                v947 = a1;
                                if ( v951 != nullptr )
                                {
                                  v953 = (_DWORD *)sub_180372160(a1, 0);
                                  v954 = (_DWORD *)sub_18000D430(v951);
                                  if ( v954 != nullptr )
                                  {
                                    while ( 1 )
                                    {
                                      v955 = (int *)v1442;
                                      v956 = v1459;
                                      v957 = v1442;
                                      v1442 = v954;
                                      v958 = v1459;
                                      if ( v957 != nullptr && *v955 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v955)-- == 1;
                                        if ( v8 )
                                        {
                                          v959 = sub_18000CC40(v955);
                                          (*(void (__fastcall **)(_QWORD *))(v959 + 48))(v957);
                                          v956 = v958;
                                        }
                                      }
                                      v960 = *v954;
                                      v961 = v956;
                                      v1459 = v954;
                                      v962 = v960 + 1;
                                      if ( v962 != 0 )
                                        *v954 = v962;
                                      if ( v956 != nullptr && *v958 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v958)-- == 1;
                                        if ( v8 )
                                        {
                                          v963 = sub_18000CC40(v956);
                                          (*(void (__fastcall **)(int *))(v963 + 48))(v961);
                                        }
                                      }
// Original IDA line 8004 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                                      v964 = qword_1803BE378;
                                      v965 = PyDict_NewPresized(2);
                                      PyDict_SetItem(v965, v964, v954);
// Original IDA line 8007 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"; qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                                      PyDict_SetItem(v965, qword_1803BE380, qword_1803BE2E0);
                                      if ( (unsigned __int8)sub_180372560(v953, v965) == 0
                                        || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                      {
                                        break;
                                      }
                                      v954 = (_DWORD *)sub_18000D430(v951);
                                      if ( v954 == nullptr )
                                        goto LABEL_1813;
                                    }
                                  }
                                  else
                                  {
LABEL_1813:
                                    if ( (unsigned __int8)sub_1800012A0(a1) != 0 )
                                    {
                                      if ( *v953 != -1 )
                                        ++*v953;
                                      if ( *v951 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v951)-- == 1;
                                        if ( v8 )
                                        {
                                          v966 = sub_18000CC40(v951);
                                          (*(void (__fastcall **)(int *))(v966 + 48))(v951);
                                        }
                                      }
                                      if ( (int)*v953 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v953)-- == 1;
                                        if ( v8 )
                                        {
                                          v967 = sub_18000CC40(v953);
                                          (*(void (__fastcall **)(_DWORD *))(v967 + 48))(v953);
                                        }
                                      }
                                      v968 = v1459;
                                      if ( v1459 != nullptr && *v1459 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1459)-- == 1;
                                        v969 = *(_QWORD **)v968;
                                        if ( v8 )
                                        {
                                          v970 = sub_18000CC40(v968);
                                          (*(void (__fastcall **)(int *))(v970 + 48))(v968);
                                          LODWORD(v969) = *v968;
                                        }
                                        if ( (int)v969 >= 0 )
                                        {
                                          v8 = (*(_QWORD *)v968)-- == 1;
                                          v1442 = *(_QWORD **)v968;
                                          if ( v8 )
                                          {
                                            v971 = sub_18000CC40(v968);
                                            (*(void (__fastcall **)(int *))(v971 + 48))(v968);
                                          }
                                        }
                                      }
                                      v972 = (int *)v1446;
                                      v973 = v1449;
                                      v151 = 479;
                                      v974 = v1446;
                                      *(_DWORD *)(v1440 + 40) = 479;
                                      v1519[0] = v1439;
                                      v1519[1] = v1448;
                                      v1535[1] = v1450;
                                      v1535[2] = v1453;
                                      v1535[3] = v1454;
                                      v1535[0] = v973;
                                      v1535[4] = v953;
// Original IDA line 8077 | qword_1803BE5C0 -> novels.services.agent_session#214 bytes [4310327, 4310373]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310329, 4310337], "value": "bookId"}, {"tag": "a", "fileRange": [4310337, 4310348], "value": "bookTitle"}, {"tag": "a", "fileRange": [4310348, 4310356], "value": "status"}, {"tag": "a"... (truncated literal)
                                      v975 = (int *)sub_18037D650(a1, v974, v1519, v1535, qword_1803BE5C0);
                                      if ( *v972 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v972)-- == 1;
                                        if ( v8 )
                                        {
                                          v976 = sub_18000CC40(v972);
                                          (*(void (__fastcall **)(int *))(v976 + 48))(v972);
                                        }
                                      }
                                      if ( *v973 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v973)-- == 1;
                                        if ( v8 )
                                        {
                                          v977 = sub_18000CC40(v973);
                                          (*(void (__fastcall **)(int *))(v977 + 48))(v973);
                                        }
                                      }
                                      v978 = v1450;
                                      if ( *(int *)v1450 >= 0 )
                                      {
                                        v8 = (*v1450)-- == 1;
                                        if ( v8 )
                                        {
                                          v979 = sub_18000CC40(v978);
                                          (*(void (__fastcall **)(_QWORD *))(v979 + 48))(v978);
                                        }
                                      }
                                      if ( (int)*v953 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v953)-- == 1;
                                        if ( v8 )
                                        {
                                          v980 = sub_18000CC40(v953);
                                          (*(void (__fastcall **)(_DWORD *))(v980 + 48))(v953);
                                        }
                                      }
                                      if ( v975 == nullptr )
                                        goto LABEL_2201;
                                      v10 = v1440;
                                      *(_DWORD *)(v1440 + 40) = 479;
                                      v981 = (int *)sub_180376770(a1, v23, v975);
                                      if ( *v23 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v23)-- == 1;
                                        if ( v8 )
                                        {
                                          v982 = sub_18000CC40(v23);
                                          (*(void (__fastcall **)(int *))(v982 + 48))(v23);
                                        }
                                      }
                                      if ( *v975 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v975)-- == 1;
                                        if ( v8 )
                                        {
                                          v983 = sub_18000CC40(v975);
                                          (*(void (__fastcall **)(int *))(v983 + 48))(v975);
                                        }
                                      }
                                      if ( v981 == nullptr )
                                        goto LABEL_2475;
                                      if ( *v981 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v981)-- == 1;
                                        if ( v8 )
                                        {
                                          v984 = sub_18000CC40(v981);
                                          (*(void (__fastcall **)(int *))(v984 + 48))(v981);
                                        }
                                      }
// Original IDA line 8149 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                                      v985 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                                      v986 = (int *)v985;
                                      if ( v985 == 0 )
                                      {
                                        sub_18000CE30(a1, &v1438);
                                        v151 = 491;
                                        goto LABEL_2516;
                                      }
                                      v987 = v1445;
                                      v151 = 491;
                                      v988 = *(_QWORD *)(v1445 + 16);
                                      *(_DWORD *)(v10 + 40) = 491;
// Original IDA line 8161 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                                      v989 = (int *)sub_18037E6C0(a1, v985, qword_1803BE0D8, v988);
                                      if ( *v986 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v986)-- == 1;
                                        if ( v8 )
                                        {
                                          v990 = sub_18000CC40(v986);
                                          (*(void (__fastcall **)(int *))(v990 + 48))(v986);
                                        }
                                      }
                                      if ( v989 == nullptr )
                                        goto LABEL_2475;
                                      if ( *v989 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v989)-- == 1;
                                        if ( v8 )
                                        {
                                          v991 = sub_18000CC40(v989);
                                          (*(void (__fastcall **)(int *))(v991 + 48))(v989);
                                        }
                                      }
// Original IDA line 8182 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                                      v992 = qword_1803BE2C8;
                                      v23 = (int *)PyDict_NewPresized(2);
                                      PyDict_SetItem(v23, v992, Py_TrueStruct);
// Original IDA line 8185 | qword_1803BE5C8 -> novels.services.agent_session#215 bytes [4310373, 4310380]: "audit"
                                      v993 = qword_1803BE5C8;
// Original IDA line 8186 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                                      v1448 = (int *)qword_1803BE2F8;
                                      v70 = (int *)PyDict_NewPresized(3);
                                      PyDict_SetItem(v70, v993, v1454);
// Original IDA line 8189 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
                                      v1442 = (_QWORD *)qword_1803BE348;
// Original IDA line 8190 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                                      v994 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                                      v995 = (int *)v994;
                                      if ( v994 != 0 )
                                      {
// Original IDA line 8194 | qword_1803BDFC8 -> novels.services.agent_session#23 bytes [4307250, 4307268]: "get_book_payload"
                                        v996 = (int *)sub_18037F0C0(a1, v994, qword_1803BDFC8);
                                        if ( *v995 >= 0 )
                                        {
                                          v8 = (*(_QWORD *)v995)-- == 1;
                                          if ( v8 )
                                          {
                                            v997 = sub_18000CC40(v995);
                                            (*(void (__fastcall **)(int *))(v997 + 48))(v995);
                                          }
                                        }
                                        if ( v996 != nullptr )
                                        {
// Original IDA line 8206 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                                          v998 = sub_18037F0C0(a1, *(_QWORD *)(v987 + 16), qword_1803BDFC0);
                                          v999 = (int *)v998;
                                          if ( v998 == 0 )
                                          {
                                            sub_18000CE30(a1, &v1438);
                                            if ( *v996 >= 0 )
                                            {
                                              v8 = (*(_QWORD *)v996)-- == 1;
                                              if ( v8 )
                                              {
                                                v1000 = sub_18000CC40(v996);
                                                (*(void (__fastcall **)(int *))(v1000 + 48))(v996);
                                              }
                                            }
                                            v151 = 496;
                                            goto LABEL_1176;
                                          }
                                          v151 = 496;
                                          *(_DWORD *)(v10 + 40) = 496;
                                          v1001 = (int *)sub_180376770(a1, v996, v998);
                                          if ( *v996 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v996)-- == 1;
                                            if ( v8 )
                                            {
                                              v1002 = sub_18000CC40(v996);
                                              (*(void (__fastcall **)(int *))(v1002 + 48))(v996);
                                            }
                                          }
                                          if ( *v999 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v999)-- == 1;
                                            if ( v8 )
                                            {
                                              v1003 = sub_18000CC40(v999);
                                              (*(void (__fastcall **)(int *))(v1003 + 48))(v999);
                                            }
                                          }
                                          if ( v1001 == nullptr )
                                            goto LABEL_1586;
                                          PyDict_SetItem(v70, v1442, v1001);
                                          if ( *v1001 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1001)-- == 1;
                                            if ( v8 )
                                            {
                                              v1004 = sub_18000CC40(v1001);
                                              (*(void (__fastcall **)(int *))(v1004 + 48))(v1001);
                                            }
                                          }
                                          v151 = 497;
                                          goto LABEL_1591;
                                        }
                                      }
                                      sub_18000CE30(a1, &v1438);
                                      v151 = 496;
                                      goto LABEL_1176;
                                    }
                                  }
                                  sub_18000CE30(a1, &v1438);
                                  v1005 = v1438;
                                  sub_18000CE70(&v1438);
                                  if ( *v951 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v951)-- == 1;
                                    if ( v8 )
                                    {
                                      v1006 = sub_18000CC40(v951);
                                      (*(void (__fastcall **)(int *))(v1006 + 48))(v951);
                                    }
                                  }
                                  if ( (int)*v953 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v953)-- == 1;
                                    if ( v8 )
                                    {
                                      v1007 = sub_18000CC40(v953);
                                      (*(void (__fastcall **)(_DWORD *))(v1007 + 48))(v953);
                                    }
                                  }
                                  v1008 = v1459;
                                  if ( v1459 != nullptr && *v1459 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1459)-- == 1;
                                    v1442 = *(_QWORD **)v1008;
                                    if ( v8 )
                                    {
                                      v1009 = sub_18000CC40(v1008);
                                      (*(void (__fastcall **)(int *))(v1009 + 48))(v1008);
                                    }
                                  }
                                  v1438 = v1005;
                                  sub_18000CE70(&v1438);
                                  if ( v1008 != nullptr && *v1008 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1008)-- == 1;
                                    v1442 = *(_QWORD **)v1008;
                                    if ( v8 )
                                    {
                                      v1010 = sub_18000CC40(v1008);
                                      (*(void (__fastcall **)(int *))(v1010 + 48))(v1008);
                                    }
                                  }
                                  v1438 = v1005;
                                  v10 = v1440;
                                  v151 = 486;
                                  goto LABEL_2516;
                                }
                              }
                            }
                            sub_18000CE30(v947, &v1438);
                            v948 = v1438;
                            sub_18000CE70(&v1438);
                            v1438 = v948;
                            v10 = v1440;
                            v151 = 486;
                            goto LABEL_2516;
                          }
                        }
                        sub_18000CE30(a1, &v1438);
                        v10 = v1440;
                        v151 = 479;
                        goto LABEL_2516;
                      }
LABEL_2060:
                      sub_18000CE30(v310, &v1438);
                      v10 = v1440;
                      goto LABEL_2516;
                    }
                    sub_18000CE30(a1, &v1438);
                    if ( *v901 < 0 || (v8 = *(_QWORD *)v901 == 1, --*(_QWORD *)v901, !v8) )
                    {
LABEL_1688:
                      v10 = v900;
                      goto LABEL_2516;
                    }
LABEL_1692:
                    v904 = sub_18000CC40(v901);
                    (*(void (__fastcall **)(int *))(v904 + 48))(v901);
                    v10 = v900;
                    goto LABEL_2516;
                  }
                }
                sub_18000CE30(a1, &v1438);
                v151 = 474;
                goto LABEL_2516;
              }
            }
            sub_18000CE30(a1, &v1438);
            v151 = 469;
            goto LABEL_2516;
          }
LABEL_2269:
// Original IDA line 8359 | qword_1803BE520 -> novels.services.agent_session#194 bytes [4309983, 4310033]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4309985, 4309994], "value": "success"}, {"tag": "a", "fileRange": [4309994, 4310003], "value": "message"}], "values": [{"tag": "F", "fileRange": [4310003, 4310004], "value": false}, {"tag... (truncated literal)
          v23 = (int *)sub_18036FB80(a1, qword_1803BE520);
          goto LABEL_150;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 466;
      goto LABEL_2516;
    }
// Original IDA line 8367 | qword_1803BE5D0 -> novels.services.agent_session#216 bytes [4310380, 4310396]: "revise_chapter"
    v1011 = sub_1803836D0(v45, qword_1803BE5D0);
    if ( v1011 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 501;
      goto LABEL_2516;
    }
    if ( v1011 == 1 )
    {
// Original IDA line 8376 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
      v1012 = sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
      v1013 = (int *)v1012;
      if ( v1012 != 0 )
      {
        v1014 = sub_180001510(v1012);
        if ( *v1013 >= 0 )
        {
          v8 = (*(_QWORD *)v1013)-- == 1;
          if ( v8 )
          {
            v1015 = sub_18000CC40(v1013);
            (*(void (__fastcall **)(int *))(v1015 + 48))(v1013);
          }
        }
        if ( v1014 != -1 )
        {
          if ( v1014 == 0 )
            goto LABEL_2269;
// Original IDA line 8394 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v1016 = sub_18037F0C0(a1, v5, qword_1803BDFD8);
          v1017 = (int *)v1016;
          if ( v1016 == 0 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 504;
            goto LABEL_2516;
          }
          v151 = 504;
          *(_DWORD *)(v10 + 40) = 504;
// Original IDA line 8404 | qword_1803BE580 -> novels.services.agent_session#206 bytes [4310224, 4310241]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310226, 4310241], "value": "chapterNumber"}]}
          v1018 = sub_1803770D0(a1, v1016, qword_1803BE580);
          if ( *v1017 >= 0 )
          {
            v8 = (*(_QWORD *)v1017)-- == 1;
            if ( v8 )
            {
              v1019 = sub_18000CC40(v1017);
              (*(void (__fastcall **)(int *))(v1019 + 48))(v1017);
            }
          }
          if ( v1018 == 0 )
            goto LABEL_2475;
          v1468 = (int *)v1018;
          v1020 = sub_180001510(v1018);
          if ( v1020 == -1 )
          {
            sub_18000CE30(a1, &v1438);
            v151 = 505;
            goto LABEL_2516;
          }
          if ( v1020 == 0 )
          {
// Original IDA line 8426 | qword_1803BE5D8 -> novels.services.agent_session#217 bytes [4310396, 4310457]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4310398, 4310407], "value": "success"}, {"tag": "a", "fileRange": [4310407, 4310416], "value": "message"}], "values": [{"tag": "F", "fileRange": [4310416, 4310417], "value": false}, {"tag... (truncated literal)
            v23 = (int *)sub_18036FB80(a1, qword_1803BE5D8);
            goto LABEL_150;
          }
// Original IDA line 8429 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
          v1021 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
          v1022 = (int *)v1021;
          if ( v1021 != 0 )
          {
// Original IDA line 8433 | qword_1803BE5D0 -> novels.services.agent_session#216 bytes [4310380, 4310396]: "revise_chapter"
            v1023 = (int *)sub_18037F0C0(a1, v1021, qword_1803BE5D0);
            if ( *v1022 >= 0 )
            {
              v8 = (*(_QWORD *)v1022)-- == 1;
              if ( v8 )
              {
                v1024 = sub_18000CC40(v1022);
                (*(void (__fastcall **)(int *))(v1024 + 48))(v1022);
              }
            }
            if ( v1023 != nullptr )
            {
// Original IDA line 8445 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
              v1025 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
              if ( v1025 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1026 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1026 + 48))(v1023);
                  }
                }
                v151 = 508;
                goto LABEL_2516;
              }
              v1027 = (int *)PyNumber_Long(v1018);
              if ( v1027 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1028 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1028 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1029 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1029 + 48))(v1025);
                  }
                }
                v151 = 509;
                goto LABEL_2516;
              }
// Original IDA line 8486 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
              v1030 = (int *)sub_18037F0C0(a1, v5, qword_1803BDFD8);
              if ( v1030 == nullptr )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1031 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1031 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1032 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1032 + 48))(v1025);
                  }
                }
                if ( *v1027 >= 0 )
                {
                  v8 = (*(_QWORD *)v1027)-- == 1;
                  if ( v8 )
                  {
                    v1033 = sub_18000CC40(v1027);
                    (*(void (__fastcall **)(int *))(v1033 + 48))(v1027);
                  }
                }
                v10 = v1440;
                v151 = 510;
                goto LABEL_2516;
              }
              v151 = 510;
              *(_DWORD *)(v1440 + 40) = 510;
// Original IDA line 8523 | qword_1803BE5E0 -> novels.services.agent_session#218 bytes [4310457, 4310465]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310459, 4310465], "value": "mode"}]}
              v469 = (int *)sub_1803770D0(a1, v1030, qword_1803BE5E0);
              if ( *v1030 >= 0 )
              {
                v8 = (*(_QWORD *)v1030)-- == 1;
                if ( v8 )
                {
                  v1034 = sub_18000CC40(v1030);
                  (*(void (__fastcall **)(int *))(v1034 + 48))(v1030);
                }
              }
              if ( v469 == nullptr )
              {
                v1035 = a1;
                goto LABEL_1954;
              }
              v1039 = sub_180001510(v469);
              if ( v1039 == -1 )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1040 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1040 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1041 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1041 + 48))(v1025);
                  }
                }
                if ( *v1027 >= 0 )
                {
                  v8 = (*(_QWORD *)v1027)-- == 1;
                  if ( v8 )
                  {
                    v1042 = sub_18000CC40(v1027);
                    (*(void (__fastcall **)(int *))(v1042 + 48))(v1027);
                  }
                }
                goto LABEL_941;
              }
              if ( v1039 != 1 )
              {
                if ( *v469 >= 0 )
                {
                  v8 = (*(_QWORD *)v469)-- == 1;
                  if ( v8 )
                  {
                    v1043 = sub_18000CC40(v469);
                    (*(void (__fastcall **)(int *))(v1043 + 48))(v469);
                  }
                }
// Original IDA line 8582 | qword_1803BE5E8 -> novels.services.agent_session#219 bytes [4310465, 4310471]: "auto"
                v469 = (int *)qword_1803BE5E8;
// Original IDA line 8583 | qword_1803BE5E8 -> novels.services.agent_session#219 bytes [4310465, 4310471]: "auto"
                if ( *(_DWORD *)qword_1803BE5E8 != -1 )
// Original IDA line 8584 | qword_1803BE5E8 -> novels.services.agent_session#219 bytes [4310465, 4310471]: "auto"
                  ++*(_DWORD *)qword_1803BE5E8;
              }
              v471 = (int *)sub_1803749A0(v469);
              if ( *v469 >= 0 )
              {
                v8 = (*(_QWORD *)v469)-- == 1;
                if ( v8 )
                {
                  v1044 = sub_18000CC40(v469);
                  (*(void (__fastcall **)(int *))(v1044 + 48))(v469);
                }
              }
              v1035 = a1;
              if ( v471 == nullptr )
              {
LABEL_1954:
                sub_18000CE30(v1035, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1036 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1036 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1037 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1037 + 48))(v1025);
                  }
                }
                if ( *v1027 >= 0 )
                {
                  v8 = (*(_QWORD *)v1027)-- == 1;
                  if ( v8 )
                  {
                    v1038 = sub_18000CC40(v1027);
                    (*(void (__fastcall **)(int *))(v1038 + 48))(v1027);
                    v10 = v1440;
                    goto LABEL_2516;
                  }
                }
                goto LABEL_277;
              }
// Original IDA line 8632 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
              v1045 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
              v1442 = (_QWORD *)v1045;
              if ( v1045 == 0 )
              {
                sub_18000CE30(a1, &v1438);
                if ( *v1023 >= 0 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1046 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1046 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1047 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1047 + 48))(v1025);
                  }
                }
                if ( *v1027 >= 0 )
                {
                  v8 = (*(_QWORD *)v1027)-- == 1;
                  if ( v8 )
                  {
                    v1048 = sub_18000CC40(v1027);
                    (*(void (__fastcall **)(int *))(v1048 + 48))(v1027);
                  }
                }
                if ( *v471 >= 0 )
                {
                  v8 = (*(_QWORD *)v471)-- == 1;
                  if ( v8 )
                  {
                    v1049 = sub_18000CC40(v471);
                    (*(void (__fastcall **)(int *))(v1049 + 48))(v471);
                  }
                }
                v10 = v1440;
                v151 = 511;
                goto LABEL_2516;
              }
              v151 = 511;
              *(_DWORD *)(v1440 + 40) = 511;
// Original IDA line 8679 | qword_1803BE3D8 -> novels.services.agent_session#153 bytes [4309347, 4309362]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309349, 4309362], "value": "instruction"}]}
              v1050 = sub_1803770D0(a1, v1045, qword_1803BE3D8);
              v1051 = v1442;
              v469 = (int *)v1050;
              if ( *(int *)v1442 >= 0 )
              {
                v8 = (*v1442)-- == 1;
                if ( v8 )
                {
                  v1052 = sub_18000CC40(v1051);
                  (*(void (__fastcall **)(_QWORD *))(v1052 + 48))(v1442);
                }
              }
              if ( v469 == nullptr )
                goto LABEL_2001;
              v1056 = sub_180001510(v469);
              if ( v1056 == -1 )
                goto LABEL_2012;
              if ( v1056 != 1 )
              {
                if ( *v469 >= 0 )
                {
                  v8 = (*(_QWORD *)v469)-- == 1;
                  if ( v8 )
                  {
                    v1061 = sub_18000CC40(v469);
                    (*(void (__fastcall **)(int *))(v1061 + 48))(v469);
                  }
                }
// Original IDA line 8707 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                v1062 = sub_18037F0C0(a1, v1443, qword_1803BDFD8);
                v1442 = (_QWORD *)v1062;
                if ( v1062 == 0 )
                  goto LABEL_2001;
                *(_DWORD *)(v1440 + 40) = 511;
// Original IDA line 8712 | qword_1803BE3E0 -> novels.services.agent_session#154 bytes [4309362, 4309373]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309364, 4309373], "value": "summary"}]}
                v1063 = sub_1803770D0(a1, v1062, qword_1803BE3E0);
                v1064 = v1442;
                v469 = (int *)v1063;
                if ( *(int *)v1442 >= 0 )
                {
                  v8 = (*v1442)-- == 1;
                  if ( v8 )
                  {
                    v1065 = sub_18000CC40(v1064);
                    (*(void (__fastcall **)(_QWORD *))(v1065 + 48))(v1442);
                  }
                }
                if ( v469 == nullptr )
                {
LABEL_2001:
                  v1053 = a1;
                  goto LABEL_2002;
                }
                v1066 = sub_180001510(v469);
                if ( v1066 == -1 )
                {
LABEL_2012:
                  sub_18000CE30(a1, &v1438);
                  if ( *v1023 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1023)-- == 1;
                    if ( v8 )
                    {
                      v1057 = sub_18000CC40(v1023);
                      (*(void (__fastcall **)(int *))(v1057 + 48))(v1023);
                    }
                  }
                  if ( *v1025 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1025)-- == 1;
                    if ( v8 )
                    {
                      v1058 = sub_18000CC40(v1025);
                      (*(void (__fastcall **)(int *))(v1058 + 48))(v1025);
                    }
                  }
                  if ( *v1027 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1027)-- == 1;
                    if ( v8 )
                    {
                      v1059 = sub_18000CC40(v1027);
                      (*(void (__fastcall **)(int *))(v1059 + 48))(v1027);
                    }
                  }
                  if ( *v471 >= 0 )
                  {
                    v8 = (*(_QWORD *)v471)-- == 1;
                    if ( v8 )
                    {
                      v1060 = sub_18000CC40(v471);
                      (*(void (__fastcall **)(int *))(v1060 + 48))(v471);
                    }
                  }
LABEL_941:
                  if ( *v469 >= 0 )
                  {
                    v8 = (*(_QWORD *)v469)-- == 1;
                    if ( v8 )
                    {
                      v516 = sub_18000CC40(v469);
                      (*(void (__fastcall **)(int *))(v516 + 48))(v469);
                      v10 = v1440;
                      goto LABEL_2516;
                    }
                  }
                  goto LABEL_277;
                }
                if ( v1066 != 1 )
                {
                  if ( *v469 >= 0 )
                  {
                    v8 = (*(_QWORD *)v469)-- == 1;
                    if ( v8 )
                    {
                      v1067 = sub_18000CC40(v469);
                      (*(void (__fastcall **)(int *))(v1067 + 48))(v469);
                    }
                  }
// Original IDA line 8796 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                  v469 = (int *)qword_1803BDFE8;
// Original IDA line 8797 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                  if ( *(_DWORD *)qword_1803BDFE8 != -1 )
// Original IDA line 8798 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                    ++*(_DWORD *)qword_1803BDFE8;
                }
              }
              v1068 = sub_1803749A0(v469);
              v29 = *v469 < 0;
              v1442 = (_QWORD *)v1068;
              if ( !v29 )
              {
                v8 = (*(_QWORD *)v469)-- == 1;
                if ( v8 )
                {
                  v1069 = sub_18000CC40(v469);
                  (*(void (__fastcall **)(int *))(v1069 + 48))(v469);
                }
              }
              v1070 = (int *)v1442;
              v1053 = a1;
              if ( v1442 != nullptr )
              {
                v151 = 507;
                *(_DWORD *)(v1440 + 40) = 507;
                v1532[2] = v1456;
                v1520[0] = v1025;
                v1520[1] = v1027;
                v1532[0] = v471;
                v1532[1] = v1070;
// Original IDA line 8824 | qword_1803BE5F0 -> novels.services.agent_session#220 bytes [4310471, 4310499]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310473, 4310479], "value": "mode"}, {"tag": "a", "fileRange": [4310479, 4310492], "value": "instruction"}, {"tag": "a", "fileRange": [4310492, 4310499], "value": "model"}]}
                v1071 = sub_18037D650(a1, v1023, v1520, v1532, qword_1803BE5F0);
                v29 = *v1023 < 0;
                v1459 = (int *)v1071;
                if ( !v29 )
                {
                  v8 = (*(_QWORD *)v1023)-- == 1;
                  if ( v8 )
                  {
                    v1072 = sub_18000CC40(v1023);
                    (*(void (__fastcall **)(int *))(v1072 + 48))(v1023);
                  }
                }
                if ( *v1025 >= 0 )
                {
                  v8 = (*(_QWORD *)v1025)-- == 1;
                  if ( v8 )
                  {
                    v1073 = sub_18000CC40(v1025);
                    (*(void (__fastcall **)(int *))(v1073 + 48))(v1025);
                  }
                }
                if ( *v1027 >= 0 )
                {
                  v8 = (*(_QWORD *)v1027)-- == 1;
                  if ( v8 )
                  {
                    v1074 = sub_18000CC40(v1027);
                    (*(void (__fastcall **)(int *))(v1074 + 48))(v1027);
                  }
                }
                if ( *v471 >= 0 )
                {
                  v8 = (*(_QWORD *)v471)-- == 1;
                  if ( v8 )
                  {
                    v1075 = sub_18000CC40(v471);
                    (*(void (__fastcall **)(int *))(v1075 + 48))(v471);
                  }
                }
                if ( *v1070 >= 0 )
                {
                  v8 = (*(_QWORD *)v1070)-- == 1;
                  if ( v8 )
                  {
                    v1076 = sub_18000CC40(v1070);
                    (*(void (__fastcall **)(int *))(v1076 + 48))(v1070);
                  }
                }
                v1077 = v1459;
                v310 = a1;
                if ( v1459 != nullptr )
                {
                  v1078 = v1441;
                  v1457 = v1459;
// Original IDA line 8878 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                  v1079 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                  v1080 = (int *)v1079;
                  if ( v1079 == 0 )
                    goto LABEL_2062;
// Original IDA line 8882 | qword_1803BE530 -> novels.services.agent_session#196 bytes [4310053, 4310064]: "load_book"
                  v1081 = (int *)sub_18037F0C0(a1, v1079, qword_1803BE530);
                  if ( *v1080 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1080)-- == 1;
                    if ( v8 )
                    {
                      v1082 = sub_18000CC40(v1080);
                      (*(void (__fastcall **)(int *))(v1082 + 48))(v1080);
                    }
                  }
                  if ( v1081 == nullptr )
                  {
LABEL_2062:
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 514;
                    goto LABEL_2516;
                  }
// Original IDA line 8900 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                  v1083 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                  if ( v1083 == nullptr )
                  {
                    sub_18000CE30(a1, &v1438);
                    if ( *v1081 >= 0 )
                    {
                      v8 = (*(_QWORD *)v1081)-- == 1;
                      if ( v8 )
                      {
                        v1084 = sub_18000CC40(v1081);
                        (*(void (__fastcall **)(int *))(v1084 + 48))(v1081);
                      }
                    }
                    v10 = v1440;
                    v151 = 514;
                    goto LABEL_2516;
                  }
                  v151 = 514;
                  *(_DWORD *)(v1440 + 40) = 514;
                  v1085 = sub_180376770(a1, v1081, v1083);
                  if ( *v1081 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1081)-- == 1;
                    if ( v8 )
                    {
                      v1086 = sub_18000CC40(v1081);
                      (*(void (__fastcall **)(int *))(v1086 + 48))(v1081);
                    }
                  }
                  if ( *v1083 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1083)-- == 1;
                    if ( v8 )
                    {
                      v1087 = sub_18000CC40(v1083);
                      (*(void (__fastcall **)(int *))(v1087 + 48))(v1083);
                    }
                  }
                  v310 = a1;
                  if ( v1085 != 0 )
                  {
                    v1461 = (int *)v1085;
// Original IDA line 8942 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                    if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
                    {
                      sub_18000CE30(a1, &v1438);
                      v10 = v1440;
                      v151 = 515;
                      goto LABEL_2516;
                    }
// Original IDA line 8949 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                    v1088 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
                    v1089 = (int *)v1088;
                    if ( v1088 != 0 )
                    {
// Original IDA line 8953 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                      v23 = (int *)sub_18037F0C0(a1, v1088, qword_1803BE0B8);
                      if ( *v1089 >= 0 )
                      {
                        v8 = (*(_QWORD *)v1089)-- == 1;
                        if ( v8 )
                        {
                          v1090 = sub_18000CC40(v1089);
                          (*(void (__fastcall **)(int *))(v1090 + 48))(v1089);
                        }
                      }
                      if ( v23 != nullptr )
                      {
// Original IDA line 8965 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                        v173 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1078 + 16), *(_QWORD *)&qword_1803BE2D8);
                        if ( v173 == nullptr )
                        {
                          sub_18000CE30(a1, &v1438);
                          if ( *v23 >= 0 )
                          {
                            v8 = (*(_QWORD *)v23)-- == 1;
                            if ( v8 )
                            {
                              v1091 = sub_18000CC40(v23);
                              (*(void (__fastcall **)(int *))(v1091 + 48))(v23);
                            }
                          }
                          v10 = v1440;
                          v151 = 516;
                          goto LABEL_2516;
                        }
// Original IDA line 8982 | qword_1803BE600 -> novels.services.agent_session#222 bytes [4310508, 4310514]: "第 "
                        v1092 = (_DWORD *)qword_1803BE600;
// Original IDA line 8983 | qword_1803BE5F8 -> novels.services.agent_session#221 bytes [4310499, 4310508]: "reviser"
                        v1449 = (int *)qword_1803BE5F8;
                        v1093 = sub_180374A40(a1, 5);
                        v177 = (int *)v1093;
                        if ( *v1092 != -1 )
                          ++*v1092;
                        sub_18000CC80(v1093, 0, v1092);
// Original IDA line 8989 | qword_1803BE5A8 -> novels.services.agent_session#211 bytes [4310289, 4310304]: "chapterNumber"
                        v1094 = sub_180001B80(a1, v1077, qword_1803BE5A8);
                        v1095 = (int *)v1094;
                        if ( v1094 != 0 )
                        {
// Original IDA line 8993 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          v1096 = sub_180375520(a1, v1094, qword_1803BDFE8);
                          if ( *v1095 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1095)-- == 1;
                            if ( v8 )
                            {
                              v1097 = sub_18000CC40(v1095);
                              (*(void (__fastcall **)(int *))(v1097 + 48))(v1095);
                            }
                          }
                          if ( v1096 != 0 )
                          {
                            sub_18000CC80(v177, 1, v1096);
// Original IDA line 9006 | qword_1803BE608 -> novels.services.agent_session#223 bytes [4310514, 4310523]: " 章《"
                            v1098 = qword_1803BE608;
// Original IDA line 9007 | qword_1803BE608 -> novels.services.agent_session#223 bytes [4310514, 4310523]: " 章《"
                            if ( *(_DWORD *)qword_1803BE608 != -1 )
// Original IDA line 9008 | qword_1803BE608 -> novels.services.agent_session#223 bytes [4310514, 4310523]: " 章《"
                              ++*(_DWORD *)qword_1803BE608;
                            sub_18000CC80(v177, 2, v1098);
// Original IDA line 9010 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                            v1099 = sub_180001B80(a1, v1459, qword_1803BE350);
                            v1100 = (int *)v1099;
                            if ( v1099 != 0 )
                            {
// Original IDA line 9014 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                              v1101 = sub_180375520(a1, v1099, qword_1803BDFE8);
                              if ( *v1100 >= 0 )
                              {
                                v8 = (*(_QWORD *)v1100)-- == 1;
                                if ( v8 )
                                {
                                  v1102 = sub_18000CC40(v1100);
                                  (*(void (__fastcall **)(int *))(v1102 + 48))(v1100);
                                }
                              }
                              if ( v1101 != 0 )
                              {
                                sub_18000CC80(v177, 3, v1101);
// Original IDA line 9027 | qword_1803BE610 -> novels.services.agent_session#224 bytes [4310523, 4310537]: "》已修订"
                                v1103 = qword_1803BE610;
// Original IDA line 9028 | qword_1803BE610 -> novels.services.agent_session#224 bytes [4310523, 4310537]: "》已修订"
                                if ( *(_DWORD *)qword_1803BE610 != -1 )
// Original IDA line 9029 | qword_1803BE610 -> novels.services.agent_session#224 bytes [4310523, 4310537]: "》已修订"
                                  ++*(_DWORD *)qword_1803BE610;
                                sub_18000CC80(v177, 4, v1103);
// Original IDA line 9031 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                                v1104 = PyUnicode_Join(qword_1803BDFE8, v177);
                                v29 = *v177 < 0;
                                v1105 = (int *)v1104;
                                v1453 = (int *)v1104;
                                if ( !v29 )
                                {
                                  v8 = (*(_QWORD *)v177)-- == 1;
                                  if ( v8 )
                                  {
                                    v1106 = sub_18000CC40(v177);
                                    (*(void (__fastcall **)(int *))(v1106 + 48))(v177);
                                  }
                                }
                                if ( v1105 == nullptr )
                                {
                                  sub_18000CE30(a1, &v1438);
                                  if ( *v23 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v23)-- == 1;
                                    if ( v8 )
                                    {
                                      v1107 = sub_18000CC40(v23);
                                      (*(void (__fastcall **)(int *))(v1107 + 48))(v23);
                                    }
                                  }
                                  if ( *v173 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v173)-- == 1;
                                    if ( v8 )
                                    {
                                      v1108 = sub_18000CC40(v173);
                                      (*(void (__fastcall **)(int *))(v1108 + 48))(v173);
                                    }
                                  }
                                  v10 = v1440;
                                  v151 = 518;
                                  goto LABEL_2516;
                                }
// Original IDA line 9069 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                                v1442 = (_QWORD *)sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                                v1109 = (int *)v1442;
                                if ( v1442 == nullptr )
                                {
                                  sub_18000CE30(a1, &v1438);
                                  if ( *v23 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v23)-- == 1;
                                    if ( v8 )
                                    {
                                      v1110 = sub_18000CC40(v23);
                                      (*(void (__fastcall **)(int *))(v1110 + 48))(v23);
                                    }
                                  }
                                  if ( *v173 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v173)-- == 1;
                                    if ( v8 )
                                    {
                                      v1111 = sub_18000CC40(v173);
                                      (*(void (__fastcall **)(int *))(v1111 + 48))(v173);
                                    }
                                  }
                                  if ( *v1105 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1105)-- == 1;
                                    if ( v8 )
                                    {
                                      v1112 = sub_18000CC40(v1105);
                                      (*(void (__fastcall **)(int *))(v1112 + 48))(v1105);
                                    }
                                  }
                                  v10 = v1440;
                                  v151 = 519;
                                  goto LABEL_2516;
                                }
// Original IDA line 9105 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                                v1448 = (int *)sub_18037F0C0(a1, v1085, qword_1803BE350);
                                if ( v1448 == nullptr )
                                {
                                  sub_18000CE30(a1, &v1438);
                                  if ( *v23 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v23)-- == 1;
                                    if ( v8 )
                                    {
                                      v1113 = sub_18000CC40(v23);
                                      (*(void (__fastcall **)(int *))(v1113 + 48))(v23);
                                    }
                                  }
                                  if ( *v173 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v173)-- == 1;
                                    if ( v8 )
                                    {
                                      v1114 = sub_18000CC40(v173);
                                      (*(void (__fastcall **)(int *))(v1114 + 48))(v173);
                                    }
                                  }
                                  if ( *v1105 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1105)-- == 1;
                                    if ( v8 )
                                    {
                                      v1115 = sub_18000CC40(v1105);
                                      (*(void (__fastcall **)(int *))(v1115 + 48))(v1105);
                                    }
                                  }
                                  if ( *v1109 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1109)-- == 1;
                                    if ( v8 )
                                    {
                                      v1116 = sub_18000CC40(v1109);
                                      (*(void (__fastcall **)(int *))(v1116 + 48))(v1109);
                                    }
                                  }
                                  v10 = v1440;
                                  v151 = 520;
                                  goto LABEL_2516;
                                }
// Original IDA line 9149 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                                v1117 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                                v1118 = (int *)v1117;
                                v1119 = a1;
                                if ( v1117 != 0 )
                                {
// Original IDA line 9154 | qword_1803BE618 -> novels.services.agent_session#225 bytes [4310537, 4310552]: "STAGES_REVISE"
                                  v1121 = (int *)sub_18037F0C0(a1, v1117, qword_1803BE618);
                                  if ( *v1118 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1118)-- == 1;
                                    if ( v8 )
                                    {
                                      v1122 = sub_18000CC40(v1118);
                                      (*(void (__fastcall **)(int *))(v1122 + 48))(v1118);
                                    }
                                  }
                                  v1119 = a1;
                                  if ( v1121 != nullptr )
                                  {
                                    v1123 = (int *)sub_180001930(a1, v1121);
                                    if ( *v1121 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v1121)-- == 1;
                                      if ( v8 )
                                      {
                                        v1124 = sub_18000CC40(v1121);
                                        (*(void (__fastcall **)(int *))(v1124 + 48))(v1121);
                                      }
                                    }
                                    v1119 = a1;
                                    if ( v1123 != nullptr )
                                    {
                                      v1450 = (_QWORD *)sub_180372160(a1, 0);
                                      v1125 = v1450;
                                      v1126 = (int *)sub_18000D430(v1123);
                                      if ( v1126 != nullptr )
                                      {
                                        while ( 1 )
                                        {
                                          v1127 = v1439;
                                          v1128 = (int *)v1454;
                                          v1129 = v1439;
                                          v1439 = v1126;
                                          v1130 = (int *)v1454;
                                          if ( v1129 != nullptr && *v1127 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1127)-- == 1;
                                            if ( v8 )
                                            {
                                              v1131 = sub_18000CC40(v1127);
                                              (*(void (__fastcall **)(int *))(v1131 + 48))(v1129);
                                              v1128 = v1130;
                                            }
                                          }
                                          v1132 = *v1126;
                                          v1133 = v1128;
                                          v1454 = v1126;
                                          v1134 = v1132 + 1;
                                          if ( v1134 != 0 )
                                            *v1126 = v1134;
                                          if ( v1128 != nullptr && *v1130 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1130)-- == 1;
                                            if ( v8 )
                                            {
                                              v1135 = sub_18000CC40(v1128);
                                              (*(void (__fastcall **)(int *))(v1135 + 48))(v1133);
                                            }
                                          }
// Original IDA line 9217 | qword_1803BE378 -> novels.services.agent_session#141 bytes [4309146, 4309152]: "name"
                                          v1136 = qword_1803BE378;
                                          v1137 = PyDict_NewPresized(2);
                                          PyDict_SetItem(v1137, v1136, v1126);
// Original IDA line 9220 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"; qword_1803BE380 -> novels.services.agent_session#142 bytes [4309152, 4309160]: "status"
                                          PyDict_SetItem(v1137, qword_1803BE380, qword_1803BE2E0);
                                          v1138 = v1137;
                                          v1125 = v1450;
                                          if ( (unsigned __int8)sub_180372560(v1450, v1138) == 0
                                            || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                          {
                                            break;
                                          }
                                          v1126 = (int *)sub_18000D430(v1123);
                                          if ( v1126 == nullptr )
                                            goto LABEL_2169;
                                        }
                                      }
                                      else
                                      {
LABEL_2169:
                                        if ( (unsigned __int8)sub_1800012A0(a1) != 0 )
                                        {
                                          if ( *v1125 != -1 )
                                            ++*v1125;
                                          if ( *v1123 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1123)-- == 1;
                                            if ( v8 )
                                            {
                                              v1139 = sub_18000CC40(v1123);
                                              (*(void (__fastcall **)(int *))(v1139 + 48))(v1123);
                                            }
                                          }
                                          if ( (int)*v1125 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1125)-- == 1;
                                            if ( v8 )
                                            {
                                              v1140 = sub_18000CC40(v1125);
                                              (*(void (__fastcall **)(_DWORD *))(v1140 + 48))(v1125);
                                            }
                                          }
                                          v1141 = (int **)v1454;
                                          if ( v1454 != nullptr && *(int *)v1454 >= 0 )
                                          {
                                            v8 = (*v1454)-- == 1;
                                            v1142 = *v1141;
                                            if ( v8 )
                                            {
                                              v1143 = sub_18000CC40(v1141);
                                              (*(void (__fastcall **)(int **))(v1143 + 48))(v1141);
                                              LODWORD(v1142) = *(_DWORD *)v1141;
                                            }
                                            if ( (int)v1142 >= 0 )
                                            {
                                              v8 = *v1141 == (int *)1;
                                              *v1141 = (int *)((char *)*v1141 - 1);
                                              v1439 = *v1141;
                                              if ( v8 )
                                              {
                                                v1144 = sub_18000CC40(v1141);
                                                (*(void (__fastcall **)(int **))(v1144 + 48))(v1141);
                                              }
                                            }
                                          }
                                          v1145 = v1453;
                                          v151 = 516;
                                          *(_DWORD *)(v1440 + 40) = 516;
                                          v1521[0] = v1449;
                                          v1533[0] = v1442;
                                          v1533[1] = v1448;
                                          v1521[1] = v1145;
                                          v1533[2] = v1125;
// Original IDA line 9289 | qword_1803BE388 -> novels.services.agent_session#143 bytes [4309160, 4309189]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4309162, 4309170], "value": "bookId"}, {"tag": "a", "fileRange": [4309170, 4309181], "value": "bookTitle"}, {"tag": "a", "fileRange": [4309181, 4309189], "value": "stages"}]}
                                          v1146 = (int *)sub_18037D650(a1, v173, v1521, v1533, qword_1803BE388);
                                          if ( *v173 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v173)-- == 1;
                                            if ( v8 )
                                            {
                                              v1147 = sub_18000CC40(v173);
                                              (*(void (__fastcall **)(int *))(v1147 + 48))(v173);
                                            }
                                          }
                                          if ( *v1145 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1145)-- == 1;
                                            if ( v8 )
                                            {
                                              v1148 = sub_18000CC40(v1145);
                                              (*(void (__fastcall **)(int *))(v1148 + 48))(v1145);
                                            }
                                          }
                                          v1149 = v1442;
                                          if ( *(int *)v1442 >= 0 )
                                          {
                                            v8 = (*v1442)-- == 1;
                                            if ( v8 )
                                            {
                                              v1150 = sub_18000CC40(v1149);
                                              (*(void (__fastcall **)(_QWORD *))(v1150 + 48))(v1149);
                                            }
                                          }
                                          v1151 = v1448;
                                          if ( *v1448 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1448)-- == 1;
                                            if ( v8 )
                                            {
                                              v1152 = sub_18000CC40(v1151);
                                              (*(void (__fastcall **)(int *))(v1152 + 48))(v1151);
                                            }
                                          }
                                          if ( (int)*v1125 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1125)-- == 1;
                                            if ( v8 )
                                            {
                                              v1153 = sub_18000CC40(v1125);
                                              (*(void (__fastcall **)(_DWORD *))(v1153 + 48))(v1125);
                                            }
                                          }
                                          if ( v1146 == nullptr )
                                            goto LABEL_2201;
                                          v10 = v1440;
                                          *(_DWORD *)(v1440 + 40) = 516;
                                          v1154 = (int *)sub_180376770(a1, v23, v1146);
                                          if ( *v23 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v23)-- == 1;
                                            if ( v8 )
                                            {
                                              v1155 = sub_18000CC40(v23);
                                              (*(void (__fastcall **)(int *))(v1155 + 48))(v23);
                                            }
                                          }
                                          if ( *v1146 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1146)-- == 1;
                                            if ( v8 )
                                            {
                                              v1156 = sub_18000CC40(v1146);
                                              (*(void (__fastcall **)(int *))(v1156 + 48))(v1146);
                                            }
                                          }
                                          if ( v1154 == nullptr )
                                            goto LABEL_2475;
                                          if ( *v1154 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1154)-- == 1;
                                            if ( v8 )
                                            {
                                              v1157 = sub_18000CC40(v1154);
                                              (*(void (__fastcall **)(int *))(v1157 + 48))(v1154);
                                            }
                                          }
// Original IDA line 9371 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                                          v1158 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                                          v1159 = (int *)v1158;
                                          if ( v1158 == 0 )
                                          {
                                            sub_18000CE30(a1, &v1438);
                                            v151 = 526;
                                            goto LABEL_2516;
                                          }
                                          v1160 = v1445;
                                          v151 = 526;
                                          v1161 = *(_QWORD *)(v1445 + 16);
                                          *(_DWORD *)(v10 + 40) = 526;
// Original IDA line 9383 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                                          v1162 = (int *)sub_18037E6C0(a1, v1158, qword_1803BE0D8, v1161);
                                          if ( *v1159 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1159)-- == 1;
                                            if ( v8 )
                                            {
                                              v1163 = sub_18000CC40(v1159);
                                              (*(void (__fastcall **)(int *))(v1163 + 48))(v1159);
                                            }
                                          }
                                          if ( v1162 == nullptr )
                                            goto LABEL_2475;
                                          if ( *v1162 >= 0 )
                                          {
                                            v8 = (*(_QWORD *)v1162)-- == 1;
                                            if ( v8 )
                                            {
                                              v1164 = sub_18000CC40(v1162);
                                              (*(void (__fastcall **)(int *))(v1164 + 48))(v1162);
                                            }
                                          }
// Original IDA line 9404 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                                          v1165 = qword_1803BE2C8;
                                          v23 = (int *)PyDict_NewPresized(2);
                                          PyDict_SetItem(v23, v1165, Py_TrueStruct);
// Original IDA line 9407 | qword_1803BE570 -> novels.services.agent_session#204 bytes [4310200, 4310209]: "chapter"
                                          v1166 = qword_1803BE570;
// Original IDA line 9408 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                                          v1448 = (int *)qword_1803BE2F8;
                                          v70 = (int *)PyDict_NewPresized(3);
                                          PyDict_SetItem(v70, v1166, v1459);
// Original IDA line 9411 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
                                          v1442 = (_QWORD *)qword_1803BE348;
// Original IDA line 9412 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                                          v1167 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                                          v1168 = (int *)v1167;
                                          if ( v1167 != 0 )
                                          {
// Original IDA line 9416 | qword_1803BDFC8 -> novels.services.agent_session#23 bytes [4307250, 4307268]: "get_book_payload"
                                            v1169 = (int *)sub_18037F0C0(a1, v1167, qword_1803BDFC8);
                                            if ( *v1168 >= 0 )
                                            {
                                              v8 = (*(_QWORD *)v1168)-- == 1;
                                              if ( v8 )
                                              {
                                                v1170 = sub_18000CC40(v1168);
                                                (*(void (__fastcall **)(int *))(v1170 + 48))(v1168);
                                              }
                                            }
                                            if ( v1169 != nullptr )
                                            {
// Original IDA line 9428 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                                              v1171 = sub_18037F0C0(a1, *(_QWORD *)(v1160 + 16), qword_1803BDFC0);
                                              v1172 = (int *)v1171;
                                              if ( v1171 == 0 )
                                              {
                                                sub_18000CE30(a1, &v1438);
                                                if ( *v1169 >= 0 )
                                                {
                                                  v8 = (*(_QWORD *)v1169)-- == 1;
                                                  if ( v8 )
                                                  {
                                                    v1173 = sub_18000CC40(v1169);
                                                    (*(void (__fastcall **)(int *))(v1173 + 48))(v1169);
                                                  }
                                                }
                                                v151 = 531;
                                                goto LABEL_1176;
                                              }
                                              v151 = 531;
                                              *(_DWORD *)(v10 + 40) = 531;
                                              v1174 = (int *)sub_180376770(a1, v1169, v1171);
                                              if ( *v1169 >= 0 )
                                              {
                                                v8 = (*(_QWORD *)v1169)-- == 1;
                                                if ( v8 )
                                                {
                                                  v1175 = sub_18000CC40(v1169);
                                                  (*(void (__fastcall **)(int *))(v1175 + 48))(v1169);
                                                }
                                              }
                                              if ( *v1172 >= 0 )
                                              {
                                                v8 = (*(_QWORD *)v1172)-- == 1;
                                                if ( v8 )
                                                {
                                                  v1176 = sub_18000CC40(v1172);
                                                  (*(void (__fastcall **)(int *))(v1176 + 48))(v1172);
                                                }
                                              }
                                              if ( v1174 == nullptr )
                                                goto LABEL_1586;
                                              PyDict_SetItem(v70, v1442, v1174);
                                              if ( *v1174 >= 0 )
                                              {
                                                v8 = (*(_QWORD *)v1174)-- == 1;
                                                if ( v8 )
                                                {
                                                  v1177 = sub_18000CC40(v1174);
                                                  (*(void (__fastcall **)(int *))(v1177 + 48))(v1174);
                                                }
                                              }
                                              v151 = 532;
                                              goto LABEL_1591;
                                            }
                                          }
                                          sub_18000CE30(a1, &v1438);
                                          v151 = 531;
                                          goto LABEL_1176;
                                        }
                                      }
                                      sub_18000CE30(a1, &v1438);
                                      v1178 = v1438;
                                      sub_18000CE70(&v1438);
                                      if ( *v1123 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1123)-- == 1;
                                        if ( v8 )
                                        {
                                          v1179 = sub_18000CC40(v1123);
                                          (*(void (__fastcall **)(int *))(v1179 + 48))(v1123);
                                        }
                                      }
                                      if ( (int)*v1125 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1125)-- == 1;
                                        if ( v8 )
                                        {
                                          v1180 = sub_18000CC40(v1125);
                                          (*(void (__fastcall **)(_DWORD *))(v1180 + 48))(v1125);
                                        }
                                      }
                                      v1181 = (int **)v1454;
                                      if ( v1454 != nullptr && *(int *)v1454 >= 0 )
                                      {
                                        v8 = (*v1454)-- == 1;
                                        v1439 = *v1181;
                                        if ( v8 )
                                        {
                                          v1182 = sub_18000CC40(v1181);
                                          (*(void (__fastcall **)(int **))(v1182 + 48))(v1181);
                                        }
                                      }
                                      v1438 = v1178;
                                      sub_18000CE70(&v1438);
                                      if ( v1181 != nullptr && *(int *)v1181 >= 0 )
                                      {
                                        v8 = *v1181 == (int *)1;
                                        *v1181 = (int *)((char *)*v1181 - 1);
                                        v1439 = *v1181;
                                        if ( v8 )
                                        {
                                          v1183 = sub_18000CC40(v1181);
                                          (*(void (__fastcall **)(int **))(v1183 + 48))(v1181);
                                        }
                                      }
                                      v1438 = v1178;
                                      v10 = v1440;
                                      v151 = 521;
                                      goto LABEL_2516;
                                    }
                                  }
                                }
                                sub_18000CE30(v1119, &v1438);
                                v1120 = v1438;
                                sub_18000CE70(&v1438);
                                v1438 = v1120;
                                v10 = v1440;
                                v151 = 521;
                                goto LABEL_2516;
                              }
                            }
                          }
                        }
                        v151 = 518;
                        goto LABEL_309;
                      }
                    }
                    sub_18000CE30(a1, &v1438);
                    v10 = v1440;
                    v151 = 516;
                    goto LABEL_2516;
                  }
                }
                goto LABEL_2060;
              }
LABEL_2002:
              sub_18000CE30(v1053, &v1438);
              if ( *v1023 >= 0 )
              {
                v8 = (*(_QWORD *)v1023)-- == 1;
                if ( v8 )
                {
                  v1054 = sub_18000CC40(v1023);
                  (*(void (__fastcall **)(int *))(v1054 + 48))(v1023);
                }
              }
              if ( *v1025 >= 0 )
              {
                v8 = (*(_QWORD *)v1025)-- == 1;
                if ( v8 )
                {
                  v1055 = sub_18000CC40(v1025);
                  (*(void (__fastcall **)(int *))(v1055 + 48))(v1025);
                }
              }
              if ( *v1027 >= 0 )
              {
                v8 = (*(_QWORD *)v1027)-- == 1;
                if ( v8 )
                {
                  v485 = sub_18000CC40(v1027);
                  v486 = v1027;
                  goto LABEL_894;
                }
              }
LABEL_895:
              if ( *v471 >= 0 )
              {
                v8 = (*(_QWORD *)v471)-- == 1;
                if ( v8 )
                {
                  v487 = sub_18000CC40(v471);
                  (*(void (__fastcall **)(int *))(v487 + 48))(v471);
                  v10 = v1440;
                  goto LABEL_2516;
                }
              }
              goto LABEL_277;
            }
          }
          sub_18000CE30(a1, &v1438);
          v151 = 507;
          goto LABEL_2516;
        }
      }
      sub_18000CE30(a1, &v1438);
      v151 = 502;
      goto LABEL_2516;
    }
// Original IDA line 9616 | qword_1803BE620 -> novels.services.agent_session#226 bytes [4310552, 4310565]: "write_batch"
    v1184 = sub_1803836D0(v45, qword_1803BE620);
    if ( v1184 == -1 )
    {
      sub_18000CE30(a1, &v1438);
      v151 = 536;
      goto LABEL_2516;
    }
    if ( v1184 != 1 )
    {
// Original IDA line 9625 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
      v1305 = qword_1803BE2C8;
      v23 = (int *)PyDict_NewPresized(2);
      PyDict_SetItem(v23, v1305, Py_FalseStruct);
// Original IDA line 9628 | qword_1803BE2D0 -> novels.services.agent_session#120 bytes [4308747, 4308756]: "message"
      v1306 = qword_1803BE2D0;
// Original IDA line 9629 | qword_1803BE6B0 -> novels.services.agent_session#244 bytes [4310826, 4310848]: "不支持的动作: "
      v1307 = (_DWORD *)qword_1803BE6B0;
      v1308 = sub_180374A40(a1, 2);
      v1309 = (int *)v1308;
      if ( *v1307 != -1 )
        ++*v1307;
      sub_18000CC80(v1308, 0, v1307);
// Original IDA line 9635 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
      v1310 = sub_180375520(a1, v1463, qword_1803BDFE8);
      if ( v1310 != 0 )
      {
        sub_18000CC80(v1309, 1, v1310);
// Original IDA line 9639 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
        v1312 = (int *)PyUnicode_Join(qword_1803BDFE8, v1309);
        if ( *v1309 >= 0 )
        {
          v8 = (*(_QWORD *)v1309)-- == 1;
          if ( v8 )
          {
            v1313 = sub_18000CC40(v1309);
            (*(void (__fastcall **)(int *))(v1313 + 48))(v1309);
          }
        }
        if ( v1312 != nullptr )
        {
          PyDict_SetItem(v23, v1306, v1312);
          if ( *v1312 >= 0 )
          {
            v8 = (*(_QWORD *)v1312)-- == 1;
            if ( v8 )
            {
              v1431 = sub_18000CC40(v1312);
              (*(void (__fastcall **)(int *))(v1431 + 48))(v1312);
            }
          }
          goto LABEL_150;
        }
        sub_18000CE30(a1, &v1438);
        v151 = 562;
      }
      else
      {
        sub_18000CE30(a1, &v1438);
        v151 = 562;
        if ( *v1309 >= 0 )
        {
          v8 = (*(_QWORD *)v1309)-- == 1;
          if ( v8 )
          {
            v1311 = sub_18000CC40(v1309);
            (*(void (__fastcall **)(int *))(v1311 + 48))(v1309);
          }
        }
      }
      goto LABEL_2512;
    }
// Original IDA line 9682 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
    v1185 = sub_18037F0C0(a1, *(_QWORD *)(v149 + 16), qword_1803BDFC0);
    v1186 = (int *)v1185;
    if ( v1185 != 0 )
    {
      v1187 = sub_180001510(v1185);
      if ( *v1186 >= 0 )
      {
        v8 = (*(_QWORD *)v1186)-- == 1;
        if ( v8 )
        {
          v1188 = sub_18000CC40(v1186);
          (*(void (__fastcall **)(int *))(v1188 + 48))(v1186);
        }
      }
      if ( v1187 != -1 )
      {
        if ( v1187 == 0 )
          goto LABEL_2269;
// Original IDA line 9700 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
        v1189 = sub_18037F0C0(a1, v5, qword_1803BDFD8);
        v1190 = (int *)v1189;
        if ( v1189 == 0 )
        {
          sub_18000CE30(a1, &v1438);
          v151 = 539;
          goto LABEL_2516;
        }
        v151 = 539;
        *(_DWORD *)(v10 + 40) = 539;
// Original IDA line 9710 | qword_1803BE628 -> novels.services.agent_session#227 bytes [4310565, 4310574]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310567, 4310574], "value": "count"}]}
        v177 = (int *)sub_1803770D0(a1, v1189, qword_1803BE628);
        if ( *v1190 >= 0 )
        {
          v8 = (*(_QWORD *)v1190)-- == 1;
          if ( v8 )
          {
            v1191 = sub_18000CC40(v1190);
            (*(void (__fastcall **)(int *))(v1191 + 48))(v1190);
          }
        }
        if ( v177 == nullptr )
          goto LABEL_2475;
        v1192 = sub_180001510(v177);
        if ( v1192 == -1 )
        {
          v435 = a1;
          goto LABEL_798;
        }
        if ( v1192 != 1 )
        {
          if ( *v177 >= 0 )
          {
            v8 = (*(_QWORD *)v177)-- == 1;
            if ( v8 )
            {
              v1193 = sub_18000CC40(v177);
              (*(void (__fastcall **)(int *))(v1193 + 48))(v177);
            }
          }
// Original IDA line 9739 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
          v1194 = sub_18037F0C0(a1, v5, qword_1803BDFD8);
          v1195 = (int *)v1194;
          if ( v1194 == 0 )
            goto LABEL_2475;
          *(_DWORD *)(v10 + 40) = 539;
// Original IDA line 9744 | qword_1803BE630 -> novels.services.agent_session#228 bytes [4310574, 4310586]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310576, 4310586], "value": "chapters"}]}
          v177 = (int *)sub_1803770D0(a1, v1194, qword_1803BE630);
          if ( *v1195 >= 0 )
          {
            v8 = (*(_QWORD *)v1195)-- == 1;
            if ( v8 )
            {
              v1196 = sub_18000CC40(v1195);
              (*(void (__fastcall **)(int *))(v1196 + 48))(v1195);
            }
          }
          if ( v177 == nullptr )
            goto LABEL_2475;
          v1197 = sub_180001510(v177);
          if ( v1197 == -1 )
          {
            v435 = a1;
LABEL_798:
            sub_18000CE30(v435, &v1438);
            if ( *v177 < 0 )
              goto LABEL_2516;
            v8 = (*(_QWORD *)v177)-- == 1;
            if ( !v8 )
              goto LABEL_2516;
            v436 = sub_18000CC40(v177);
            v437 = v177;
            goto LABEL_2515;
          }
          if ( v1197 != 1 )
          {
            if ( *v177 >= 0 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v1198 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v1198 + 48))(v177);
              }
            }
// Original IDA line 9782 | qword_1803BE638 -> novels.services.agent_session#229 bytes [4310586, 4310588]: 5
            v177 = (int *)qword_1803BE638;
// Original IDA line 9783 | qword_1803BE638 -> novels.services.agent_session#229 bytes [4310586, 4310588]: 5
            if ( *(_DWORD *)qword_1803BE638 != -1 )
// Original IDA line 9784 | qword_1803BE638 -> novels.services.agent_session#229 bytes [4310586, 4310588]: 5
              ++*(_DWORD *)qword_1803BE638;
          }
        }
        v1199 = PyNumber_Long(v177);
        if ( *v177 >= 0 )
        {
          v8 = (*(_QWORD *)v177)-- == 1;
          if ( v8 )
          {
            v1200 = sub_18000CC40(v177);
            (*(void (__fastcall **)(int *))(v1200 + 48))(v177);
          }
        }
        if ( v1199 == 0 )
          goto LABEL_2475;
        v1481 = (int *)v1199;
// Original IDA line 9800 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
        v1201 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
        v1202 = (int *)v1201;
        if ( v1201 != 0 )
        {
// Original IDA line 9804 | qword_1803BE640 -> novels.services.agent_session#230 bytes [4310588, 4310610]: "write_batch_chapters"
          v177 = (int *)sub_18037F0C0(a1, v1201, qword_1803BE640);
          if ( *v1202 >= 0 )
          {
            v8 = (*(_QWORD *)v1202)-- == 1;
            if ( v8 )
            {
              v1203 = sub_18000CC40(v1202);
              (*(void (__fastcall **)(int *))(v1203 + 48))(v1202);
            }
          }
          if ( v177 != nullptr )
          {
// Original IDA line 9816 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
            v23 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
            if ( v23 == nullptr )
            {
              sub_18000CE30(a1, &v1438);
              if ( *v177 >= 0 )
              {
                v8 = (*(_QWORD *)v177)-- == 1;
                if ( v8 )
                {
                  v1204 = sub_18000CC40(v177);
                  (*(void (__fastcall **)(int *))(v1204 + 48))(v177);
                }
              }
              v151 = 541;
              goto LABEL_2516;
            }
// Original IDA line 9832 | qword_1803BE230 -> novels.services.agent_session#100 bytes [4308395, 4308400]: "max"
            v1205 = sub_18000D570(qword_1803BE230);
            v1206 = qword_1803AB258;
// Original IDA line 9834 | qword_1803BE648 -> novels.services.agent_session#231 bytes [4310610, 4310615]: "min"
            v1207 = sub_18000D570(qword_1803BE648);
// Original IDA line 9835 | qword_1803BE650 -> novels.services.agent_session#232 bytes [4310615, 4310617]: 50
            v1208 = qword_1803BE650;
            v151 = 542;
            *(_DWORD *)(v1440 + 40) = 542;
            v1522[0] = v1208;
            v1522[1] = v1199;
            v1209 = (int *)sub_180377750(a1, v1207, v1522);
            v316 = a1;
            if ( v1209 == nullptr )
              goto LABEL_572;
            *(_DWORD *)(v1440 + 40) = 542;
            v1523[0] = v1206;
            v1523[1] = v1209;
            v1210 = (int *)sub_180377750(a1, v1205, v1523);
            if ( *v1209 >= 0 )
            {
              v8 = (*(_QWORD *)v1209)-- == 1;
              if ( v8 )
              {
                v1211 = sub_18000CC40(v1209);
                (*(void (__fastcall **)(int *))(v1211 + 48))(v1209);
              }
            }
            v316 = a1;
            if ( v1210 == nullptr )
              goto LABEL_572;
            v1212 = v1456;
            v900 = v1440;
            v151 = 540;
            *(_DWORD *)(v1440 + 40) = 540;
            v1524[1] = v1212;
            v1506 = v23;
            v1524[0] = v1210;
// Original IDA line 9867 | qword_1803BE658 -> novels.services.agent_session#233 bytes [4310617, 4310633]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310619, 4310626], "value": "count"}, {"tag": "a", "fileRange": [4310626, 4310633], "value": "model"}]}
            v1213 = sub_18037D030(a1, v177, &v1506, v1524, qword_1803BE658);
            v29 = *v177 < 0;
            v1214 = (int *)v1213;
            v1442 = (_QWORD *)v1213;
            if ( !v29 )
            {
              v8 = (*(_QWORD *)v177)-- == 1;
              if ( v8 )
              {
                v1215 = sub_18000CC40(v177);
                (*(void (__fastcall **)(int *))(v1215 + 48))(v177);
              }
            }
            if ( *v23 >= 0 )
            {
              v8 = (*(_QWORD *)v23)-- == 1;
              if ( v8 )
              {
                v1216 = sub_18000CC40(v23);
                (*(void (__fastcall **)(int *))(v1216 + 48))(v23);
              }
            }
            if ( *v1210 >= 0 )
            {
              v8 = (*(_QWORD *)v1210)-- == 1;
              if ( v8 )
              {
                v1217 = sub_18000CC40(v1210);
                (*(void (__fastcall **)(int *))(v1217 + 48))(v1210);
              }
            }
            if ( v1214 == nullptr )
            {
LABEL_2327:
              sub_18000CE30(a1, &v1438);
              v10 = v900;
              goto LABEL_2516;
            }
            v1457 = v1214;
// Original IDA line 9906 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
            v1218 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
            v1219 = (int *)v1218;
            v1220 = a1;
            if ( v1218 != 0 )
            {
// Original IDA line 9911 | qword_1803BE530 -> novels.services.agent_session#196 bytes [4310053, 4310064]: "load_book"
              v1221 = (int *)sub_18037F0C0(a1, v1218, qword_1803BE530);
              if ( *v1219 >= 0 )
              {
                v8 = (*(_QWORD *)v1219)-- == 1;
                if ( v8 )
                {
                  v1222 = sub_18000CC40(v1219);
                  (*(void (__fastcall **)(int *))(v1222 + 48))(v1219);
                }
              }
              v1220 = a1;
              if ( v1221 != nullptr )
              {
// Original IDA line 9924 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                v1223 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BDFC0);
                v1224 = (int *)v1223;
                if ( v1223 == 0 )
                {
                  sub_18000CE30(a1, &v1438);
                  if ( *v1221 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1221)-- == 1;
                    if ( v8 )
                    {
                      v1225 = sub_18000CC40(v1221);
                      (*(void (__fastcall **)(int *))(v1225 + 48))(v1221);
                      v151 = 545;
                      v10 = v900;
                      goto LABEL_2516;
                    }
                  }
LABEL_2330:
                  v151 = 545;
                  v10 = v900;
                  goto LABEL_2516;
                }
                v151 = 545;
                *(_DWORD *)(v900 + 40) = 545;
                v1226 = sub_180376770(a1, v1221, v1223);
                if ( *v1221 >= 0 )
                {
                  v8 = (*(_QWORD *)v1221)-- == 1;
                  if ( v8 )
                  {
                    v1227 = sub_18000CC40(v1221);
                    (*(void (__fastcall **)(int *))(v1227 + 48))(v1221);
                  }
                }
                if ( *v1224 >= 0 )
                {
                  v8 = (*(_QWORD *)v1224)-- == 1;
                  if ( v8 )
                  {
                    v1228 = sub_18000CC40(v1224);
                    (*(void (__fastcall **)(int *))(v1228 + 48))(v1224);
                  }
                }
                if ( v1226 == 0 )
                  goto LABEL_2327;
                v1461 = (int *)v1226;
// Original IDA line 9970 | qword_1803BE258 -> novels.services.agent_session#105 bytes [4308444, 4308459]: "pendingAction"
                if ( (unsigned __int8)sub_18037F3A0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE258, Py_NoneStruct) == 0 )
                {
                  sub_18000CE30(a1, &v1438);
                  v151 = 546;
                  v10 = v900;
                  goto LABEL_2516;
                }
// Original IDA line 9977 | qword_1803BE660 -> novels.services.agent_session#234 bytes [4310633, 4310656]: "批量写作完成："
                v1229 = (_DWORD *)qword_1803BE660;
                v1230 = sub_180374A40(a1, 5);
                v901 = (int *)v1230;
                if ( *v1229 != -1 )
                  ++*v1229;
                sub_18000CC80(v1230, 0, v1229);
// Original IDA line 9983 | qword_1803BE2E0 -> novels.services.agent_session#122 bytes [4308771, 4308782]: "completed"
                v1231 = sub_180001B80(a1, v1214, qword_1803BE2E0);
                v1232 = (int *)v1231;
                if ( v1231 != 0 )
                {
// Original IDA line 9987 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                  v1233 = sub_180375520(a1, v1231, qword_1803BDFE8);
                  if ( *v1232 >= 0 )
                  {
                    v8 = (*(_QWORD *)v1232)-- == 1;
                    if ( v8 )
                    {
                      v1234 = sub_18000CC40(v1232);
                      (*(void (__fastcall **)(int *))(v1234 + 48))(v1232);
                    }
                  }
                  if ( v1233 != 0 )
                  {
                    sub_18000CC80(v901, 1, v1233);
// Original IDA line 10000 | qword_1803BE668 -> novels.services.agent_session#235 bytes [4310656, 4310658]: "/"
                    v1235 = qword_1803BE668;
// Original IDA line 10001 | qword_1803BE668 -> novels.services.agent_session#235 bytes [4310656, 4310658]: "/"
                    if ( *(_DWORD *)qword_1803BE668 != -1 )
// Original IDA line 10002 | qword_1803BE668 -> novels.services.agent_session#235 bytes [4310656, 4310658]: "/"
                      ++*(_DWORD *)qword_1803BE668;
                    sub_18000CC80(v901, 2, v1235);
// Original IDA line 10004 | qword_1803BE670 -> novels.services.agent_session#236 bytes [4310658, 4310665]: "total"
                    v1236 = sub_180001B80(a1, v1214, qword_1803BE670);
                    v1237 = (int *)v1236;
                    if ( v1236 != 0 )
                    {
// Original IDA line 10008 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                      v1238 = sub_180375520(a1, v1236, qword_1803BDFE8);
                      if ( *v1237 >= 0 )
                      {
                        v8 = (*(_QWORD *)v1237)-- == 1;
                        if ( v8 )
                        {
                          v1239 = sub_18000CC40(v1237);
                          (*(void (__fastcall **)(int *))(v1239 + 48))(v1237);
                        }
                      }
                      if ( v1238 != 0 )
                      {
                        sub_18000CC80(v901, 3, v1238);
// Original IDA line 10021 | qword_1803BE678 -> novels.services.agent_session#237 bytes [4310665, 4310671]: " 章"
                        v1240 = qword_1803BE678;
// Original IDA line 10022 | qword_1803BE678 -> novels.services.agent_session#237 bytes [4310665, 4310671]: " 章"
                        if ( *(_DWORD *)qword_1803BE678 != -1 )
// Original IDA line 10023 | qword_1803BE678 -> novels.services.agent_session#237 bytes [4310665, 4310671]: " 章"
                          ++*(_DWORD *)qword_1803BE678;
                        sub_18000CC80(v901, 4, v1240);
// Original IDA line 10025 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                        v1241 = PyUnicode_Join(qword_1803BDFE8, v901);
                        if ( *v901 >= 0 )
                        {
                          v8 = (*(_QWORD *)v901)-- == 1;
                          if ( v8 )
                          {
                            v1242 = sub_18000CC40(v901);
                            (*(void (__fastcall **)(int *))(v1242 + 48))(v901);
                          }
                        }
                        if ( v1241 == 0 )
                        {
                          sub_18000CE30(a1, &v1438);
                          v10 = v1440;
                          v151 = 547;
                          goto LABEL_2516;
                        }
                        v1460 = (int *)v1241;
// Original IDA line 10043 | qword_1803BDFD8 -> novels.services.agent_session#25 bytes [4307288, 4307293]: "get"
                        v1243 = (int *)sub_18037F0C0(a1, v1214, qword_1803BDFD8);
                        if ( v1243 == nullptr )
                        {
                          sub_18000CE30(a1, &v1438);
                          v10 = v1440;
                          v151 = 548;
                          goto LABEL_2516;
                        }
                        v151 = 548;
                        *(_DWORD *)(v1440 + 40) = 548;
// Original IDA line 10053 | qword_1803BE680 -> novels.services.agent_session#238 bytes [4310671, 4310688]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310673, 4310688], "value": "stoppedReason"}]}
                        v177 = (int *)sub_1803770D0(a1, v1243, qword_1803BE680);
                        if ( *v1243 >= 0 )
                        {
                          v8 = (*(_QWORD *)v1243)-- == 1;
                          if ( v8 )
                          {
                            v1244 = sub_18000CC40(v1243);
                            (*(void (__fastcall **)(int *))(v1244 + 48))(v1243);
                          }
                        }
                        if ( v177 == nullptr )
                          goto LABEL_2465;
                        v1245 = sub_180001510(v177);
                        if ( v1245 == -1 )
                          goto LABEL_565;
                        if ( *v177 >= 0 )
                        {
                          v8 = (*(_QWORD *)v177)-- == 1;
                          if ( v8 )
                          {
                            v1246 = sub_18000CC40(v177);
                            (*(void (__fastcall **)(int *))(v1246 + 48))(v177);
                          }
                        }
                        if ( v1245 != 0 )
                        {
// Original IDA line 10079 | qword_1803BE688 -> novels.services.agent_session#239 bytes [4310688, 4310778]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4310690, 4310701], "value": "cancelled"}, {"tag": "a", "fileRange": [4310701, 4310717], "value": "target_reached"}, {"tag": "a", "fileRange": [4310717, 4310739], "value": "consecutive_fai... (truncated literal)
                          v1247 = sub_18036FB80(a1, qword_1803BE688);
// Original IDA line 10080 | qword_1803BE690 -> novels.services.agent_session#240 bytes [4310778, 4310783]: "（"
                          v1248 = (_DWORD *)qword_1803BE690;
                          v1470 = (int *)v1247;
                          v1488 = (int *)v1241;
                          v1249 = sub_180374A40(a1, 3);
                          v177 = (int *)v1249;
                          if ( *v1248 != -1 )
                            ++*v1248;
                          sub_18000CC80(v1249, 0, v1248);
// Original IDA line 10088 | qword_1803BE698 -> novels.services.agent_session#241 bytes [4310783, 4310798]: "stoppedReason"
                          v1250 = (int *)sub_180001B80(a1, v1214, qword_1803BE698);
                          v1251 = a1;
                          if ( v1250 == nullptr )
                            goto LABEL_2407;
// Original IDA line 10092 | qword_1803BE698 -> novels.services.agent_session#241 bytes [4310783, 4310798]: "stoppedReason"
                          v1252 = (int *)sub_180001B80(a1, v1214, qword_1803BE698);
                          if ( v1252 == nullptr )
                          {
                            sub_18000CE30(a1, &v1438);
                            if ( *v1250 >= 0 )
                            {
                              v8 = (*(_QWORD *)v1250)-- == 1;
                              if ( v8 )
                              {
                                v1253 = sub_18000CC40(v1250);
                                (*(void (__fastcall **)(int *))(v1253 + 48))(v1250);
                                v151 = 550;
                                goto LABEL_316;
                              }
                            }
LABEL_2408:
                            v151 = 550;
                            goto LABEL_316;
                          }
                          v1254 = (int *)sub_18036F8E0(a1, v1470, v1250);
                          if ( v1254 == nullptr && (unsigned __int8)sub_18005BAD0(a1) == 0 )
                          {
                            v1254 = v1252;
                            if ( *v1252 != -1 )
                              ++*v1252;
                          }
                          if ( *v1250 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1250)-- == 1;
                            if ( v8 )
                            {
                              v1255 = sub_18000CC40(v1250);
                              (*(void (__fastcall **)(int *))(v1255 + 48))(v1250);
                            }
                          }
                          if ( *v1252 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1252)-- == 1;
                            if ( v8 )
                            {
                              v1256 = sub_18000CC40(v1252);
                              (*(void (__fastcall **)(int *))(v1256 + 48))(v1252);
                            }
                          }
                          v1251 = a1;
                          if ( v1254 == nullptr )
                          {
LABEL_2407:
                            sub_18000CE30(v1251, &v1438);
                            goto LABEL_2408;
                          }
// Original IDA line 10143 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          v1257 = sub_180375520(a1, v1254, qword_1803BDFE8);
                          if ( *v1254 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1254)-- == 1;
                            if ( v8 )
                            {
                              v1258 = sub_18000CC40(v1254);
                              (*(void (__fastcall **)(int *))(v1258 + 48))(v1254);
                            }
                          }
                          if ( v1257 == 0 )
                          {
                            v1251 = a1;
                            goto LABEL_2407;
                          }
                          sub_18000CC80(v177, 1, v1257);
// Original IDA line 10159 | qword_1803BE1C8 -> novels.services.agent_session#87 bytes [4308228, 4308233]: "）"
                          v1259 = qword_1803BE1C8;
// Original IDA line 10160 | qword_1803BE1C8 -> novels.services.agent_session#87 bytes [4308228, 4308233]: "）"
                          if ( *(_DWORD *)qword_1803BE1C8 != -1 )
// Original IDA line 10161 | qword_1803BE1C8 -> novels.services.agent_session#87 bytes [4308228, 4308233]: "）"
                            ++*(_DWORD *)qword_1803BE1C8;
                          sub_18000CC80(v177, 2, v1259);
// Original IDA line 10163 | qword_1803BDFE8 -> novels.services.agent_session#27 bytes [4307299, 4307300]: ""
                          v1260 = (int *)PyUnicode_Join(qword_1803BDFE8, v177);
                          if ( *v177 >= 0 )
                          {
                            v8 = (*(_QWORD *)v177)-- == 1;
                            if ( v8 )
                            {
                              v1261 = sub_18000CC40(v177);
                              (*(void (__fastcall **)(int *))(v1261 + 48))(v177);
                            }
                          }
                          if ( v1260 == nullptr )
                          {
                            sub_18000CE30(a1, &v1438);
                            v10 = v1440;
                            v151 = 550;
                            goto LABEL_2516;
                          }
                          sub_180382D00(&v1488, v1260);
                          if ( *v1260 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1260)-- == 1;
                            if ( v8 )
                            {
                              v1262 = sub_18000CC40(v1260);
                              (*(void (__fastcall **)(int *))(v1262 + 48))(v1260);
                            }
                          }
                          v1460 = v1488;
                        }
                        v1263 = v1445;
// Original IDA line 10193 | qword_1803BE070 -> novels.services.agent_session#44 bytes [4307561, 4307571]: "messages"
                        v1264 = sub_18037F0C0(a1, *(_QWORD *)(v1445 + 16), qword_1803BE070);
                        v1265 = (int *)v1264;
                        if ( v1264 != 0 )
                        {
// Original IDA line 10197 | qword_1803BE0B8 -> novels.services.agent_session#53 bytes [4307652, 4307660]: "append"
                          v177 = (int *)sub_18037F0C0(a1, v1264, qword_1803BE0B8);
                          if ( *v1265 >= 0 )
                          {
                            v8 = (*(_QWORD *)v1265)-- == 1;
                            if ( v8 )
                            {
                              v1266 = sub_18000CC40(v1265);
                              (*(void (__fastcall **)(int *))(v1266 + 48))(v1265);
                            }
                          }
                          if ( v177 != nullptr )
                          {
// Original IDA line 10209 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                            v1267 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), *(_QWORD *)&qword_1803BE2D8);
                            if ( v1267 != nullptr )
                            {
// Original IDA line 10212 | qword_1803BE540 -> novels.services.agent_session#198 bytes [4310093, 4310101]: "writer"
                              v1269 = qword_1803BE540;
// Original IDA line 10213 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                              v1270 = (int *)sub_18037F0C0(a1, *(_QWORD *)(v1263 + 16), qword_1803BDFC0);
                              if ( v1270 != nullptr )
                              {
// Original IDA line 10216 | qword_1803BE350 -> novels.services.agent_session#136 bytes [4309067, 4309074]: "title"
                                v1273 = (int *)sub_18037F0C0(a1, v1226, qword_1803BE350);
                                if ( v1273 != nullptr )
                                {
                                  v151 = 551;
                                  *(_DWORD *)(v1440 + 40) = 551;
                                  v1526[1] = v1460;
                                  v1526[0] = v1269;
                                  v1525[0] = v1270;
                                  v1525[1] = v1273;
// Original IDA line 10225 | qword_1803BE6A0 -> novels.services.agent_session#242 bytes [4310798, 4310819]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4310800, 4310808], "value": "bookId"}, {"tag": "a", "fileRange": [4310808, 4310819], "value": "bookTitle"}]}
                                  v1277 = (int *)sub_18037D650(a1, v1267, v1526, v1525, qword_1803BE6A0);
                                  if ( *v1267 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1267)-- == 1;
                                    if ( v8 )
                                    {
                                      v1278 = sub_18000CC40(v1267);
                                      (*(void (__fastcall **)(int *))(v1278 + 48))(v1267);
                                    }
                                  }
                                  if ( *v1270 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1270)-- == 1;
                                    if ( v8 )
                                    {
                                      v1279 = sub_18000CC40(v1270);
                                      (*(void (__fastcall **)(int *))(v1279 + 48))(v1270);
                                    }
                                  }
                                  if ( *v1273 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1273)-- == 1;
                                    if ( v8 )
                                    {
                                      v1280 = sub_18000CC40(v1273);
                                      (*(void (__fastcall **)(int *))(v1280 + 48))(v1273);
                                    }
                                  }
                                  v314 = a1;
                                  if ( v1277 == nullptr )
                                    goto LABEL_566;
                                  *(_DWORD *)(v1440 + 40) = 551;
                                  v1281 = (int *)sub_180376770(a1, v177, v1277);
                                  if ( *v177 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v177)-- == 1;
                                    if ( v8 )
                                    {
                                      v1282 = sub_18000CC40(v177);
                                      (*(void (__fastcall **)(int *))(v1282 + 48))(v177);
                                    }
                                  }
                                  if ( *v1277 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1277)-- == 1;
                                    if ( v8 )
                                    {
                                      v1283 = sub_18000CC40(v1277);
                                      (*(void (__fastcall **)(int *))(v1283 + 48))(v1277);
                                    }
                                  }
                                  if ( v1281 == nullptr )
                                    goto LABEL_2465;
                                  if ( *v1281 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1281)-- == 1;
                                    if ( v8 )
                                    {
                                      v1284 = sub_18000CC40(v1281);
                                      (*(void (__fastcall **)(int *))(v1284 + 48))(v1281);
                                    }
                                  }
// Original IDA line 10287 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
                                  v1285 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF10);
                                  v1286 = (int *)v1285;
                                  if ( v1285 == 0 )
                                  {
                                    sub_18000CE30(a1, &v1438);
                                    v10 = v1440;
                                    v151 = 552;
                                    goto LABEL_2516;
                                  }
                                  v1287 = v1445;
                                  v151 = 552;
                                  v10 = v1440;
                                  v1288 = *(_QWORD *)(v1445 + 16);
                                  *(_DWORD *)(v1440 + 40) = 552;
// Original IDA line 10301 | qword_1803BE0D8 -> novels.services.agent_session#57 bytes [4307693, 4307707]: "save_session"
                                  v1289 = (int *)sub_18037E6C0(a1, v1285, qword_1803BE0D8, v1288);
                                  if ( *v1286 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1286)-- == 1;
                                    if ( v8 )
                                    {
                                      v1290 = sub_18000CC40(v1286);
                                      (*(void (__fastcall **)(int *))(v1290 + 48))(v1286);
                                    }
                                  }
                                  if ( v1289 == nullptr )
                                    goto LABEL_2475;
                                  if ( *v1289 >= 0 )
                                  {
                                    v8 = (*(_QWORD *)v1289)-- == 1;
                                    if ( v8 )
                                    {
                                      v1291 = sub_18000CC40(v1289);
                                      (*(void (__fastcall **)(int *))(v1291 + 48))(v1289);
                                    }
                                  }
// Original IDA line 10322 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
                                  v1292 = qword_1803BE2C8;
                                  v23 = (int *)PyDict_NewPresized(2);
                                  PyDict_SetItem(v23, v1292, Py_TrueStruct);
// Original IDA line 10325 | qword_1803BE6A8 -> novels.services.agent_session#243 bytes [4310819, 4310826]: "batch"
                                  v1293 = qword_1803BE6A8;
// Original IDA line 10326 | qword_1803BE2F8 -> novels.services.agent_session#125 bytes [4308841, 4308847]: "data"
                                  v1448 = (int *)qword_1803BE2F8;
                                  v70 = (int *)PyDict_NewPresized(3);
                                  PyDict_SetItem(v70, v1293, v1442);
// Original IDA line 10329 | qword_1803BE348 -> novels.services.agent_session#135 bytes [4309061, 4309067]: "book"
                                  v1442 = (_QWORD *)qword_1803BE348;
// Original IDA line 10330 | qword_1803BDF18 -> novels.services.agent_session#1 bytes [4306937, 4306947]: "pipeline"
                                  v1294 = sub_18037F0C0(a1, *(_QWORD *)(v1441 + 16), qword_1803BDF18);
                                  v1295 = (int *)v1294;
                                  if ( v1294 != 0 )
                                  {
// Original IDA line 10334 | qword_1803BDFC8 -> novels.services.agent_session#23 bytes [4307250, 4307268]: "get_book_payload"
                                    v1296 = (int *)sub_18037F0C0(a1, v1294, qword_1803BDFC8);
                                    if ( *v1295 >= 0 )
                                    {
                                      v8 = (*(_QWORD *)v1295)-- == 1;
                                      if ( v8 )
                                      {
                                        v1297 = sub_18000CC40(v1295);
                                        (*(void (__fastcall **)(int *))(v1297 + 48))(v1295);
                                      }
                                    }
                                    if ( v1296 != nullptr )
                                    {
// Original IDA line 10346 | qword_1803BDFC0 -> novels.services.agent_session#22 bytes [4307242, 4307250]: "bookId"
                                      v1298 = sub_18037F0C0(a1, *(_QWORD *)(v1287 + 16), qword_1803BDFC0);
                                      v1299 = (int *)v1298;
                                      if ( v1298 == 0 )
                                      {
                                        sub_18000CE30(a1, &v1438);
                                        if ( *v1296 >= 0 )
                                        {
                                          v8 = (*(_QWORD *)v1296)-- == 1;
                                          if ( v8 )
                                          {
                                            v1300 = sub_18000CC40(v1296);
                                            (*(void (__fastcall **)(int *))(v1300 + 48))(v1296);
                                          }
                                        }
                                        v151 = 557;
                                        goto LABEL_1176;
                                      }
                                      v151 = 557;
                                      *(_DWORD *)(v10 + 40) = 557;
                                      v1301 = (int *)sub_180376770(a1, v1296, v1298);
                                      if ( *v1296 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1296)-- == 1;
                                        if ( v8 )
                                        {
                                          v1302 = sub_18000CC40(v1296);
                                          (*(void (__fastcall **)(int *))(v1302 + 48))(v1296);
                                        }
                                      }
                                      if ( *v1299 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1299)-- == 1;
                                        if ( v8 )
                                        {
                                          v1303 = sub_18000CC40(v1299);
                                          (*(void (__fastcall **)(int *))(v1303 + 48))(v1299);
                                        }
                                      }
                                      if ( v1301 == nullptr )
                                        goto LABEL_1586;
                                      PyDict_SetItem(v70, v1442, v1301);
                                      if ( *v1301 >= 0 )
                                      {
                                        v8 = (*(_QWORD *)v1301)-- == 1;
                                        if ( v8 )
                                        {
                                          v1304 = sub_18000CC40(v1301);
                                          (*(void (__fastcall **)(int *))(v1304 + 48))(v1301);
                                        }
                                      }
                                      v151 = 558;
                                      goto LABEL_1591;
                                    }
                                  }
                                  sub_18000CE30(a1, &v1438);
                                  v151 = 557;
                                  goto LABEL_1176;
                                }
                                sub_18000CE30(a1, &v1438);
                                if ( *v177 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v177)-- == 1;
                                  if ( v8 )
                                  {
                                    v1274 = sub_18000CC40(v177);
                                    (*(void (__fastcall **)(int *))(v1274 + 48))(v177);
                                  }
                                }
                                if ( *v1267 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1267)-- == 1;
                                  if ( v8 )
                                  {
                                    v1275 = sub_18000CC40(v1267);
                                    (*(void (__fastcall **)(int *))(v1275 + 48))(v1267);
                                  }
                                }
                                if ( *v1270 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1270)-- == 1;
                                  if ( v8 )
                                  {
                                    v1276 = sub_18000CC40(v1270);
                                    (*(void (__fastcall **)(int *))(v1276 + 48))(v1270);
                                    v10 = v1440;
                                    v151 = 551;
                                    goto LABEL_2516;
                                  }
                                }
                              }
                              else
                              {
                                sub_18000CE30(a1, &v1438);
                                if ( *v177 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v177)-- == 1;
                                  if ( v8 )
                                  {
                                    v1271 = sub_18000CC40(v177);
                                    (*(void (__fastcall **)(int *))(v1271 + 48))(v177);
                                  }
                                }
                                if ( *v1267 >= 0 )
                                {
                                  v8 = (*(_QWORD *)v1267)-- == 1;
                                  if ( v8 )
                                  {
                                    v1272 = sub_18000CC40(v1267);
                                    (*(void (__fastcall **)(int *))(v1272 + 48))(v1267);
                                    v10 = v1440;
                                    v151 = 551;
                                    goto LABEL_2516;
                                  }
                                }
                              }
                            }
                            else
                            {
                              sub_18000CE30(a1, &v1438);
                              if ( *v177 >= 0 )
                              {
                                v8 = (*(_QWORD *)v177)-- == 1;
                                if ( v8 )
                                {
                                  v1268 = sub_18000CC40(v177);
                                  (*(void (__fastcall **)(int *))(v1268 + 48))(v177);
                                }
                              }
                            }
                            v10 = v1440;
                            v151 = 551;
                            goto LABEL_2516;
                          }
                        }
                        sub_18000CE30(a1, &v1438);
                        v10 = v1440;
                        v151 = 551;
                        goto LABEL_2516;
                      }
                    }
                  }
                }
                v151 = 547;
                sub_18000CE30(a1, &v1438);
                if ( *v901 < 0 )
                  goto LABEL_1688;
                v8 = (*(_QWORD *)v901)-- == 1;
                if ( !v8 )
                  goto LABEL_1688;
                goto LABEL_1692;
              }
            }
            sub_18000CE30(v1220, &v1438);
            goto LABEL_2330;
          }
        }
        sub_18000CE30(a1, &v1438);
        v151 = 540;
        goto LABEL_2516;
      }
    }
    sub_18000CE30(a1, &v1438);
    v151 = 537;
    goto LABEL_2516;
  }
// Original IDA line 10511 | qword_1803BE2C8 -> novels.services.agent_session#119 bytes [4308738, 4308747]: "success"
  v43 = qword_1803BE2C8;
  v23 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v23, v43, Py_FalseStruct);
// Original IDA line 10514 | qword_1803BE2D0 -> novels.services.agent_session#120 bytes [4308747, 4308756]: "message"
  PyDict_SetItem(v23, qword_1803BE2D0, v40);
LABEL_150:
  v11 = (int *)v1441;
  v20 = v5;
LABEL_151:
  sub_180001D80(a1, v22);
  v95 = v1445;
  if ( *(int *)v1445 >= 0 )
  {
    v8 = (*(_QWORD *)v1445)-- == 1;
    if ( v8 )
    {
      v96 = sub_18000CC40(v95);
      (*(void (__fastcall **)(__int64))(v96 + 48))(v95);
    }
  }
  if ( *v5 >= 0 )
  {
    v8 = (*(_QWORD *)v5)-- == 1;
    if ( v8 )
    {
      v97 = sub_18000CC40(v20);
      (*(void (__fastcall **)(int *))(v97 + 48))(v20);
    }
  }
  v98 = v1452;
  if ( v1452 != nullptr && *v1452 >= 0 )
  {
    v8 = (*(_QWORD *)v1452)-- == 1;
    if ( v8 )
    {
      v99 = sub_18000CC40(v98);
      (*(void (__fastcall **)(int *))(v99 + 48))(v98);
    }
  }
  v100 = v1473;
  if ( v1473 != nullptr && *v1473 >= 0 )
  {
    v8 = (*(_QWORD *)v1473)-- == 1;
    if ( v8 )
    {
      v101 = sub_18000CC40(v100);
      (*(void (__fastcall **)(int *))(v101 + 48))(v100);
    }
  }
  v102 = v1474;
  if ( *v1474 >= 0 )
  {
    v8 = (*(_QWORD *)v1474)-- == 1;
    if ( v8 )
    {
      v103 = sub_18000CC40(v102);
      (*(void (__fastcall **)(int *))(v103 + 48))(v102);
    }
  }
  v104 = v1466;
  if ( v1466 != nullptr && *v1466 >= 0 )
  {
    v8 = (*(_QWORD *)v1466)-- == 1;
    if ( v8 )
    {
      v105 = sub_18000CC40(v104);
      (*(void (__fastcall **)(int *))(v105 + 48))(v104);
    }
  }
  v106 = v1461;
  if ( v1461 != nullptr && *v1461 >= 0 )
  {
    v8 = (*(_QWORD *)v1461)-- == 1;
    if ( v8 )
    {
      v107 = sub_18000CC40(v106);
      (*(void (__fastcall **)(int *))(v107 + 48))(v106);
    }
  }
  v108 = v1475;
  if ( v1475 != nullptr && *v1475 >= 0 )
  {
    v8 = (*(_QWORD *)v1475)-- == 1;
    if ( v8 )
    {
      v109 = sub_18000CC40(v108);
      (*(void (__fastcall **)(int *))(v109 + 48))(v108);
    }
  }
  v110 = v1476;
  if ( v1476 != nullptr && *v1476 >= 0 )
  {
    v8 = (*(_QWORD *)v1476)-- == 1;
    if ( v8 )
    {
      v111 = sub_18000CC40(v110);
      (*(void (__fastcall **)(int *))(v111 + 48))(v110);
    }
  }
  v112 = v1477;
  if ( v1477 != nullptr && *v1477 >= 0 )
  {
    v8 = (*(_QWORD *)v1477)-- == 1;
    if ( v8 )
    {
      v113 = sub_18000CC40(v112);
      (*(void (__fastcall **)(int *))(v113 + 48))(v112);
    }
  }
  v114 = v1478;
  if ( v1478 != nullptr && *v1478 >= 0 )
  {
    v8 = (*(_QWORD *)v1478)-- == 1;
    if ( v8 )
    {
      v115 = sub_18000CC40(v114);
      (*(void (__fastcall **)(int *))(v115 + 48))(v114);
    }
  }
  v116 = v1479;
  if ( v1479 != nullptr && *v1479 >= 0 )
  {
    v8 = (*(_QWORD *)v1479)-- == 1;
    if ( v8 )
    {
      v117 = sub_18000CC40(v116);
      (*(void (__fastcall **)(int *))(v117 + 48))(v116);
    }
  }
  v118 = v1457;
  if ( v1457 != nullptr && *v1457 >= 0 )
  {
    v8 = (*(_QWORD *)v1457)-- == 1;
    if ( v8 )
    {
      v119 = sub_18000CC40(v118);
      (*(void (__fastcall **)(int *))(v119 + 48))(v118);
    }
  }
  v120 = v1480;
  if ( v1480 != nullptr && *v1480 >= 0 )
  {
    v8 = (*(_QWORD *)v1480)-- == 1;
    if ( v8 )
    {
      v121 = sub_18000CC40(v120);
      (*(void (__fastcall **)(int *))(v121 + 48))(v120);
    }
  }
  v122 = v1483;
  if ( v1483 != nullptr && *v1483 >= 0 )
  {
    v8 = (*(_QWORD *)v1483)-- == 1;
    if ( v8 )
    {
      v123 = sub_18000CC40(v122);
      (*(void (__fastcall **)(int *))(v123 + 48))(v122);
    }
  }
  v124 = v1467;
  if ( v1467 != nullptr && *v1467 >= 0 )
  {
    v8 = (*(_QWORD *)v1467)-- == 1;
    if ( v8 )
    {
      v125 = sub_18000CC40(v124);
      (*(void (__fastcall **)(int *))(v125 + 48))(v124);
    }
  }
  v126 = v1462;
  if ( v1462 != nullptr && *v1462 >= 0 )
  {
    v8 = (*(_QWORD *)v1462)-- == 1;
    if ( v8 )
    {
      v127 = sub_18000CC40(v126);
      (*(void (__fastcall **)(int *))(v127 + 48))(v126);
    }
  }
  v128 = v1460;
  if ( v1460 != nullptr && *v1460 >= 0 )
  {
    v8 = (*(_QWORD *)v1460)-- == 1;
    if ( v8 )
    {
      v129 = sub_18000CC40(v128);
      (*(void (__fastcall **)(int *))(v129 + 48))(v128);
    }
  }
  v130 = v1468;
  if ( v1468 != nullptr && *v1468 >= 0 )
  {
    v8 = (*(_QWORD *)v1468)-- == 1;
    if ( v8 )
    {
      v131 = sub_18000CC40(v130);
      (*(void (__fastcall **)(int *))(v131 + 48))(v130);
    }
  }
  v132 = v1469;
  if ( v1469 != nullptr && *v1469 >= 0 )
  {
    v8 = (*(_QWORD *)v1469)-- == 1;
    if ( v8 )
    {
      v133 = sub_18000CC40(v132);
      (*(void (__fastcall **)(int *))(v133 + 48))(v132);
    }
  }
  v134 = v1464;
  if ( v1464 != nullptr && *(int *)v1464 >= 0 )
  {
    v8 = (*v1464)-- == 1;
    if ( v8 )
    {
      v135 = sub_18000CC40(v134);
      (*(void (__fastcall **)(_QWORD *))(v135 + 48))(v134);
    }
  }
  v136 = v1481;
  if ( v1481 != nullptr && *v1481 >= 0 )
  {
    v8 = (*(_QWORD *)v1481)-- == 1;
    if ( v8 )
    {
      v137 = sub_18000CC40(v136);
      (*(void (__fastcall **)(int *))(v137 + 48))(v136);
    }
  }
  v138 = v1470;
  if ( v1470 != nullptr && *v1470 >= 0 )
  {
    v8 = (*(_QWORD *)v1470)-- == 1;
    if ( v8 )
    {
      v139 = sub_18000CC40(v138);
      (*(void (__fastcall **)(int *))(v139 + 48))(v138);
    }
  }
  if ( *v11 >= 0 )
  {
    v8 = (*(_QWORD *)v11)-- == 1;
    if ( v8 )
    {
      v140 = sub_18000CC40(v11);
      (*(void (__fastcall **)(int *))(v140 + 48))(v11);
    }
  }
  v141 = v1482;
  if ( *v1482 >= 0 )
  {
    v8 = (*(_QWORD *)v1482)-- == 1;
    if ( v8 )
    {
      v142 = sub_18000CC40(v141);
      (*(void (__fastcall **)(int *))(v142 + 48))(v141);
    }
  }
  v143 = v1456;
  if ( *v1456 >= 0 )
  {
    v8 = (*(_QWORD *)v1456)-- == 1;
    if ( v8 )
    {
      v144 = sub_18000CC40(v143);
      (*(void (__fastcall **)(int *))(v144 + 48))(v143);
    }
  }
  return v23;
}
