// write access to const memory has been detected, the output may be wrong!
int *__fastcall sub_1801F5CB0(__int64 a1, __int64 a2, _QWORD *a3)
{
  __int64 v5; // rax
  __int64 v6; // rcx
  __int64 v7; // rdi
  __int64 v8; // rsi
  int *v9; // rax
  _QWORD *v10; // r14
  int *v11; // r15
  __int64 v12; // rax
  __int64 v13; // r12
  __int64 v14; // rax
  int *v15; // rbx
  unsigned int v16; // esi
  int *v17; // r12
  __int64 v18; // r9
  __int64 v19; // rdi
  int *v20; // rdx
  int *v21; // r9
  __int64 v22; // rdx
  __int64 v23; // rax
  int *v24; // r15
  __int64 v25; // rbx
  __int64 v26; // rax
  int *v27; // rbx
  int v28; // edi
  __int64 v29; // rax
  int *v30; // rbx
  int v31; // edi
  __int64 v32; // rax
  int *v33; // rbx
  int *v34; // rdi
  __int64 v35; // rdx
  __int64 v36; // rdx
  int *v37; // r14
  __int64 v38; // rcx
  __int64 v39; // rdx
  __int64 v40; // rax
  int *v41; // rbx
  int v42; // edi
  __int64 v43; // rax
  __int64 v44; // rax
  int *v45; // r15
  __int64 v46; // rax
  int *v47; // rdi
  int *v48; // rbx
  __int64 v49; // rax
  __int64 v50; // rax
  int *v51; // rdi
  __int64 v52; // rax
  __int64 v53; // r14
  __int64 v54; // rax
  __int64 v55; // rax
  __int64 v56; // rax
  __int64 v57; // rax
  __int64 v58; // rax
  __int64 v59; // rax
  __int64 v60; // rax
  __int64 v61; // rax
  __int64 v62; // rbx
  __int64 v63; // rdi
  __int64 v64; // rax
  int *v65; // rbx
  __int64 v66; // rax
  bool v67; // sf
  int *v68; // rdi
  __int64 v69; // rax
  int *v70; // rsi
  __int64 v71; // r14
  __int64 v72; // rax
  int *v73; // r12
  int *v74; // rbx
  __int64 v75; // rdi
  int *v76; // rcx
  int *v77; // r15
  int *v78; // rdx
  int *v79; // rdi
  int *v80; // r14
  __int64 v81; // rax
  int v82; // eax
  int *v83; // rdi
  int v84; // eax
  __int64 v85; // rax
  int IsInstance; // eax
  __int64 v87; // rax
  int *v88; // rdi
  __int64 v89; // rcx
  __int64 v90; // r14
  int *v91; // rbx
  __int64 v92; // rax
  int v93; // eax
  __int64 v94; // rax
  int *v95; // rdi
  __int64 v96; // rax
  int *v97; // rbx
  __int64 v98; // rax
  __int64 v99; // rax
  _QWORD *v100; // rax
  int v101; // eax
  __int64 v102; // rcx
  __int64 v103; // rax
  int *v104; // rbx
  __int64 v105; // rcx
  int *v106; // rdi
  __int64 v107; // rax
  __int64 v108; // r14
  __int64 v109; // rax
  _QWORD *v110; // rbx
  __int64 v111; // rax
  int v112; // eax
  __int64 v113; // rax
  __int64 v114; // rax
  __int64 v115; // rax
  int v116; // eax
  __int64 v117; // rax
  __int64 v118; // r12
  int *v119; // r14
  int *v120; // r15
  int *v121; // rdi
  int *v122; // rdx
  int *v123; // rcx
  int *v124; // rbx
  int *v125; // rsi
  __int64 v126; // rax
  int v127; // eax
  int *v128; // rbx
  int v129; // eax
  __int64 v130; // rax
  __int64 v131; // rax
  int *v132; // rbx
  int v133; // esi
  __int64 v134; // rax
  __int64 v135; // rax
  __int64 v136; // rax
  int **v137; // rbx
  int *v138; // rax
  __int64 v139; // rax
  __int64 v140; // rax
  int *v141; // rbx
  bool v142; // zf
  __int64 v143; // rax
  __int64 v144; // r15
  __int64 v145; // rax
  __int64 v146; // rax
  int *v147; // rdi
  __int64 v148; // rax
  __int64 v149; // rax
  __int64 v150; // rax
  __int64 v151; // rax
  __int64 v152; // rax
  int *v153; // rbx
  int v154; // eax
  __int64 v155; // rax
  int *v156; // r14
  __int64 v157; // rax
  int *v158; // rsi
  __int64 v159; // rax
  int *v160; // rcx
  int *v161; // rdi
  __int64 v162; // rax
  int v163; // eax
  __int64 v164; // rax
  __int64 v165; // rax
  int v166; // esi
  __int64 v167; // rax
  __int64 v168; // rax
  int *v169; // rcx
  int *v170; // r15
  __int64 v171; // rax
  int *v172; // r14
  int *v173; // rbx
  int v174; // eax
  int v175; // edx
  int v176; // r8d
  __int64 v177; // rax
  int *v178; // rbx
  __int64 v179; // rax
  int *v180; // rdi
  __int64 v181; // rax
  __int64 v182; // rax
  int *v183; // rbx
  __int64 v184; // r14
  __int64 v185; // rax
  int v186; // eax
  __int64 v187; // rdi
  int *v188; // rbx
  __int64 v189; // r8
  __int64 v190; // rax
  __int64 v191; // rax
  __int64 v192; // rdi
  __int64 v193; // rax
  __int64 v194; // rax
  int *v195; // rdi
  int v196; // edx
  int v197; // r8d
  __int64 v198; // rax
  int *v199; // rbx
  __int64 v200; // rax
  int *v201; // r14
  __int64 v202; // rax
  int *v203; // rbx
  int *v204; // r8
  __int64 v205; // rax
  __int64 v206; // r14
  __int64 v207; // rax
  __int64 v208; // rdi
  __int64 v209; // rbx
  __int64 v210; // rsi
  __int64 v211; // rax
  int *v212; // rbx
  __int64 v213; // rcx
  __int64 v214; // rdi
  __int64 v215; // rax
  __int64 v216; // rsi
  int v217; // ecx
  __int64 v218; // rax
  int *v219; // rbx
  __int64 v220; // rax
  _QWORD *v221; // rcx
  __int64 v222; // rax
  int *v223; // r12
  int *v224; // rcx
  int *v225; // rdx
  int *v226; // rbx
  int *v227; // r15
  int *v228; // rdi
  __int64 v229; // rax
  int v230; // eax
  int *v231; // rbx
  int v232; // eax
  __int64 v233; // rax
  int *v234; // rbx
  int *v235; // rdi
  __int64 v236; // rax
  __int64 v237; // r14
  __int64 v238; // rax
  __int64 v239; // rax
  __int64 v240; // rax
  int *v241; // rbx
  char v242; // di
  __int64 v243; // rax
  int v244; // eax
  int *v245; // rdi
  __int64 v246; // rax
  _QWORD *v247; // rdi
  __int64 v248; // rax
  int *v249; // rbx
  __int64 v250; // rax
  _QWORD *v251; // rbx
  __int64 v252; // rax
  int *v253; // rax
  __int64 v254; // rdi
  __int64 v255; // rax
  int *v256; // rbx
  __int64 v257; // rdi
  __int64 v258; // rax
  int *v259; // rbx
  __int64 v260; // rax
  __int64 v261; // rax
  int *v262; // rbx
  int *v263; // r12
  int *v264; // rcx
  int *v265; // rdx
  int *v266; // rbx
  int *v267; // r14
  int *v268; // rdi
  __int64 v269; // rax
  int v270; // eax
  int *v271; // rbx
  int v272; // eax
  __int64 v273; // rax
  int *v274; // rbx
  int *v275; // rdi
  __int64 v276; // rax
  __int64 v277; // r15
  __int64 v278; // rax
  __int64 v279; // rax
  int v280; // eax
  __int64 v281; // rbx
  int *v282; // r14
  __int64 v283; // r15
  int *v284; // rbx
  int *v285; // rdi
  __int64 v286; // rax
  int *v287; // rbx
  __int64 v288; // rax
  __int64 v289; // rax
  int *v290; // rdi
  __int64 v291; // rax
  __int64 v292; // rax
  _QWORD *v293; // r14
  __int64 v294; // rax
  int *v295; // rdi
  __int64 v296; // rax
  int *v297; // rdi
  __int64 v298; // rax
  __int64 v299; // rbx
  __int64 v300; // rax
  __int64 v301; // rdi
  __int64 v302; // rbx
  __int64 v303; // rsi
  int *v304; // r12
  __int64 v305; // rax
  int *v306; // rbx
  int *v307; // r15
  __int64 v308; // rax
  __int64 v309; // rdi
  __int64 v310; // rbx
  __int64 v311; // rdi
  __int64 v312; // rax
  __int64 v313; // r12
  __int64 v314; // rax
  int *v315; // rbx
  __int64 v316; // rax
  __int64 v317; // rax
  __int64 v318; // rax
  int *v319; // rbx
  __int64 v320; // rax
  __int64 v321; // rax
  __int64 v322; // rcx
  __int64 v323; // rax
  int *v324; // r15
  __int64 v325; // rax
  int *v326; // rbx
  __int64 v327; // r9
  int *v328; // rdi
  __int64 v329; // rax
  __int64 v330; // rcx
  __int64 v331; // rax
  _DWORD *v332; // rdi
  __int64 v333; // rax
  int *v334; // rbx
  __int64 v335; // rax
  __int64 v336; // r8
  __int64 v337; // rax
  int *v338; // rdi
  int *v339; // rsi
  __int64 v340; // rax
  __int64 v341; // rdi
  __int64 v342; // rax
  __int64 v343; // rax
  __int64 v344; // rdi
  __int64 v345; // rax
  __int64 v346; // rdx
  int *v347; // rdi
  __int64 v348; // rax
  __int64 v349; // rax
  __int64 v350; // rdx
  int *v351; // rbx
  __int64 v352; // rcx
  __int64 v353; // rdx
  __int64 v354; // rax
  int *v355; // rdi
  __int64 v356; // rax
  __int64 v357; // rax
  int *v358; // r15
  __int64 v359; // rax
  _DWORD *v360; // rbx
  __int64 v361; // rax
  int *v362; // rbx
  __int64 v363; // rax
  __int64 v364; // rdx
  __int64 v365; // r9
  __int64 v366; // rax
  int v367; // edi
  __int64 v368; // rax
  __int64 v369; // r14
  _QWORD *v370; // rdi
  int *v371; // rbx
  __int64 v372; // rax
  int *v373; // rdi
  __int64 v374; // rcx
  __int64 v375; // rax
  int v376; // esi
  __int64 v377; // rax
  __int64 v378; // rax
  __int64 v379; // rdi
  __int64 v380; // rax
  __int64 v381; // rbx
  int *v382; // r12
  __int64 v383; // rax
  int *v384; // rbx
  __int64 v385; // rcx
  __int64 v386; // r14
  __int64 v387; // rax
  __int64 v388; // rbx
  __int64 v389; // rsi
  __int64 v390; // rdi
  int *v391; // rbx
  __int64 v392; // rax
  int *v393; // rcx
  __int64 v394; // rax
  int *v395; // rbx
  int *v396; // rsi
  __int64 v397; // rax
  __int64 v398; // rbx
  unsigned int v399; // edi
  __int64 v400; // rbx
  __int64 v401; // rax
  int *v402; // rbx
  int *v403; // r12
  __int64 v404; // rax
  int *v405; // rdi
  _QWORD *v406; // r14
  __int64 v407; // rbx
  __int64 v408; // rax
  int *v409; // rsi
  __int64 v410; // rax
  __int64 v411; // rax
  __int64 v412; // rax
  int v413; // ecx
  int v414; // eax
  __int64 v415; // rax
  int v416; // ecx
  int v417; // eax
  __int64 v418; // rax
  __int64 v419; // rax
  int *v420; // rbx
  int v421; // eax
  __int64 v422; // rax
  __int64 v423; // rax
  int *v424; // rsi
  __int64 v425; // rax
  _DWORD *v426; // r12
  __int64 v427; // rax
  __int64 v428; // rax
  __int64 v429; // rax
  int *v430; // rbx
  int v431; // eax
  __int64 v432; // rax
  __int64 v433; // rax
  __int64 v434; // rdi
  __int64 v435; // rax
  int v436; // eax
  __int64 v437; // rax
  int *v438; // r15
  _DWORD *v439; // r14
  int *v440; // rcx
  int *v441; // rdx
  _QWORD *v442; // rbx
  int *v443; // rsi
  __int64 v444; // rax
  int v445; // eax
  int *v446; // rbx
  int v447; // eax
  __int64 v448; // rax
  int *v449; // rbx
  int *v450; // rsi
  __int64 v451; // rax
  int v452; // ebx
  __int64 v453; // rax
  __int64 v454; // rdx
  int *v455; // rbx
  int *v456; // rsi
  __int64 v457; // rax
  int *v458; // r14
  __int64 v459; // rax
  __int64 v460; // rax
  int *v461; // r15
  __int64 v462; // rax
  __int64 v463; // rax
  _QWORD *v464; // rbx
  __int64 v465; // rax
  __int64 v466; // r14
  _QWORD *v467; // rdi
  __int64 v468; // rdx
  int *v469; // rax
  int *v470; // rbx
  __int64 v471; // rdx
  __int64 v472; // rax
  __int64 v473; // rax
  int *v474; // rdi
  __int64 v475; // rax
  int v476; // esi
  __int64 v477; // rax
  __int64 v478; // rax
  __int64 v479; // rsi
  __int64 v480; // rax
  int v481; // eax
  __int64 v482; // rax
  int v483; // eax
  __int64 v484; // rdi
  _DWORD *v485; // rdi
  __int64 v486; // rax
  int *v487; // rbx
  __int64 v488; // rax
  __int64 v489; // rax
  __int64 v490; // r8
  __int64 v491; // rax
  __int64 v492; // rax
  _QWORD *v493; // rcx
  __int64 v494; // rax
  _QWORD *v495; // rbx
  __int64 v496; // rax
  int *v497; // rsi
  __int64 v498; // r14
  __int64 v499; // rax
  __int64 v500; // rcx
  __int64 v501; // rax
  int v502; // esi
  __int64 v503; // rax
  __int64 v504; // rdx
  int *v505; // rbx
  __int64 v506; // rax
  int *v507; // rsi
  __int64 v508; // rax
  int *v509; // r14
  __int64 v510; // rax
  __int64 v511; // rax
  char v512; // bl
  __int64 v513; // rax
  _QWORD *v514; // rsi
  __int64 v515; // rax
  __int64 v516; // rax
  __int64 v517; // rax
  __int64 v518; // r8
  __int64 v519; // rax
  __int64 v520; // rax
  __int64 v521; // rax
  __int64 v522; // rdx
  __int64 v523; // rax
  int *v524; // rbx
  int *v525; // r12
  __int64 v526; // rax
  __int64 v527; // rsi
  int *v528; // r14
  __int64 v529; // rbx
  int *v530; // r15
  int *v531; // rdi
  int *v532; // rdx
  int *v533; // rcx
  int *v534; // rbx
  int *v535; // rsi
  __int64 v536; // rax
  int v537; // eax
  int *v538; // rbx
  int v539; // eax
  __int64 v540; // rax
  __int64 v541; // rax
  __int64 v542; // rbx
  __int64 v543; // rax
  __int64 v544; // rax
  __int64 v545; // rax
  int *v546; // rbx
  int *v547; // rax
  __int64 v548; // rax
  __int64 v549; // rax
  int *v550; // rbx
  __int64 v551; // rax
  __int64 v552; // rax
  __int64 v553; // rax
  int *v554; // r14
  __int64 v555; // rsi
  _DWORD *v556; // r15
  int *v557; // rdi
  __int64 v558; // rax
  __int64 v559; // rsi
  int *v560; // rdi
  __int64 v561; // rax
  int *v562; // rcx
  __int64 v563; // rax
  __int64 v564; // rax
  __int64 v565; // rax
  __int64 v566; // rdi
  int *v567; // rbx
  __int64 v568; // rax
  __int64 v569; // rdx
  __int64 v570; // rbx
  int *v571; // rbx
  __int64 v572; // rbx
  int *v573; // rdi
  __int64 v574; // rax
  __int64 v575; // rbx
  int *v576; // r14
  __int64 v577; // rax
  __int64 v578; // rax
  __int64 v579; // rcx
  int *v580; // rdi
  __int64 v581; // rax
  __int64 v582; // rbx
  __int64 v583; // rax
  int *v584; // rbx
  int *v585; // r12
  __int64 v586; // rax
  int *v587; // r15
  __int64 v588; // rbx
  __int64 v589; // rax
  int *v590; // rdi
  __int64 v591; // rax
  __int64 v592; // rax
  __int64 v593; // rax
  int v594; // eax
  int *v595; // rdi
  int v596; // eax
  __int64 v597; // rax
  __int64 v598; // rax
  int v599; // eax
  int *v600; // rbx
  int v601; // eax
  __int64 v602; // rax
  __int64 v603; // rax
  int v604; // eax
  __int64 v605; // r14
  __int64 v606; // rcx
  _QWORD *v607; // rdi
  __int64 v608; // rdx
  int *v609; // rax
  int *v610; // rbx
  __int64 v611; // rdx
  __int64 v612; // rax
  __int64 v613; // rax
  int *v614; // rdi
  __int64 v615; // rax
  int v616; // esi
  __int64 v617; // rax
  __int64 v618; // rax
  __int64 v619; // rdi
  __int64 v620; // rax
  int *v621; // rbx
  __int64 v622; // rax
  int v623; // eax
  __int64 v624; // rax
  __int64 v625; // rax
  __int64 v626; // rax
  int *v627; // rdi
  __int64 v628; // r14
  __int64 v629; // rax
  __int64 v630; // rax
  int v631; // edi
  __int64 v632; // rax
  __int64 v633; // rdx
  int *v634; // rbx
  __int64 v635; // rax
  int *v636; // rdi
  __int64 v637; // rax
  int *v638; // r14
  __int64 v639; // rax
  __int64 v640; // rax
  char v641; // bl
  __int64 v642; // rax
  int *v643; // rbx
  __int64 v644; // rax
  __int64 v645; // rax
  __int64 v646; // r15
  __int64 v647; // r14
  __int64 v648; // rax
  __int64 v649; // rax
  int *v650; // rbx
  int *v651; // rdi
  __int64 v652; // rax
  int *v653; // rbx
  __int64 v654; // rax
  __int64 v655; // rax
  int *v656; // r15
  __int64 v657; // rax
  __int64 v658; // r14
  _DWORD *v659; // rdi
  __int64 v660; // rax
  int *v661; // rbx
  __int64 v662; // rdx
  __int64 v663; // rax
  int *v664; // rdi
  __int64 v665; // rcx
  int *v666; // rsi
  __int64 v667; // rax
  __int64 v668; // rdi
  __int64 v669; // rax
  __int64 v670; // rax
  int *v671; // rdi
  __int64 v672; // rax
  int *v673; // r12
  __int64 v674; // rax
  __int64 v675; // rdi
  __int64 v676; // rax
  __int64 v677; // r14
  __int64 v678; // rax
  int *v679; // rbx
  __int64 v680; // rax
  __int64 v681; // rax
  _QWORD *v682; // rbx
  int v683; // eax
  int *v684; // rbx
  int v685; // eax
  _QWORD *v686; // r15
  __int64 v687; // rdx
  __int64 v688; // rdx
  __int64 v689; // rax
  int *v690; // rbx
  __int64 v691; // rcx
  int *v692; // rdi
  __int64 v693; // rax
  int *v694; // rbx
  __int64 v695; // rax
  __int64 v696; // rax
  __int64 v697; // rdx
  char v698; // al
  __int64 v699; // rcx
  __int64 v700; // rdx
  __int64 v701; // r8
  __int64 v702; // rax
  int *v703; // rdi
  int *v704; // rbx
  __int64 v705; // rax
  __int64 v706; // r8
  __int64 v707; // rax
  int *v708; // rdi
  __int64 v709; // rax
  __int64 v710; // rax
  __int64 v711; // rdx
  int *v712; // r14
  __int64 v713; // rbx
  __int64 v714; // rdi
  int *v715; // r15
  int *v716; // rdi
  int *v717; // rbx
  __int64 v718; // rax
  __int64 v719; // rax
  __int64 v720; // rax
  __int64 v721; // rbx
  __int64 v722; // rdi
  __int64 v723; // rdx
  __int64 v724; // rax
  int *v725; // rbx
  __int64 v726; // rdx
  __int64 v727; // rax
  int *v728; // rcx
  _QWORD *v729; // r13
  _QWORD *v730; // r15
  _QWORD *v731; // rsi
  int *v732; // rdi
  int *v733; // rbx
  int *v734; // rcx
  int *v735; // rcx
  _QWORD *v736; // rcx
  int *v737; // rcx
  _QWORD *v738; // rcx
  int *v739; // rcx
  int *v740; // rcx
  int *v741; // rcx
  int *v742; // rcx
  _QWORD *v743; // rcx
  int *v744; // rcx
  int *v745; // rcx
  int *v746; // rcx
  int *v747; // rcx
  int *v748; // rcx
  _QWORD *v749; // rcx
  int *v750; // rcx
  int *v751; // rcx
  _QWORD *v752; // rcx
  _QWORD *v753; // rcx
  int *v754; // rcx
  int *v755; // rcx
  _QWORD *v756; // rcx
  _QWORD *v757; // rcx
  int *v758; // rcx
  int *v759; // rcx
  _QWORD *v760; // rcx
  _QWORD *v761; // rcx
  _QWORD *v762; // rcx
  _QWORD *v763; // rcx
  int *v764; // rcx
  _QWORD *v765; // rcx
  int *v766; // rcx
  int *v767; // rcx
  int *v768; // rcx
  int *v769; // rcx
  int *v770; // rcx
  int *v771; // rcx
  __int64 v773; // rbx
  __int64 v774; // rax
  __int64 v775; // rax
  int *v776; // rdi
  __int64 v777; // rax
  __int64 v778; // rax
  __int64 v779; // rax
  int *v780; // rsi
  int *v781; // r12
  __int64 v782; // rdx
  __int64 v783; // r8
  __int64 v784; // r9
  __int64 v785; // rcx
  int *v786; // rcx
  __int64 v787; // rbx
  _QWORD *v788; // rcx
  _QWORD *v789; // rcx
  int *v790; // rcx
  int *v791; // rcx
  int *v792; // rcx
  int *v793; // rcx
  int *v794; // rcx
  int *v795; // rcx
  int *v796; // rcx
  int *v797; // rcx
  int *v798; // rcx
  int *v799; // rcx
  _QWORD *v800; // rcx
  int *v801; // rcx
  int *v802; // rcx
  _QWORD *v803; // rcx
  _QWORD *v804; // rcx
  int *v805; // rcx
  int *v806; // rcx
  _QWORD *v807; // rcx
  _QWORD *v808; // rcx
  int *v809; // rcx
  int *v810; // rcx
  _QWORD *v811; // rcx
  _QWORD *v812; // rcx
  _QWORD *v813; // rcx
  _QWORD *v814; // rcx
  int *v815; // rcx
  int *v816; // rcx
  _QWORD *v817; // rcx
  int *v818; // rcx
  int *v819; // rcx
  int *v820; // rcx
  int *v821; // rcx
  int *v822; // rcx
  int *v823; // rcx
  int *v824; // rcx
  __int64 v825; // [rsp+20h] [rbp-1C0h]
  __int64 v826; // [rsp+20h] [rbp-1C0h]
  int *v827; // [rsp+20h] [rbp-1C0h]
  __int64 v828; // [rsp+28h] [rbp-1B8h]
  __int64 v829; // [rsp+28h] [rbp-1B8h]
  int *v830; // [rsp+28h] [rbp-1B8h]
  int *v831; // [rsp+30h] [rbp-1B0h]
  int *v832; // [rsp+38h] [rbp-1A8h]
  int *v833; // [rsp+40h] [rbp-1A0h]
  int *v834; // [rsp+48h] [rbp-198h]
  int *v835; // [rsp+50h] [rbp-190h]
  int *v836; // [rsp+58h] [rbp-188h]
  _QWORD *v837; // [rsp+60h] [rbp-180h]
  int *v838; // [rsp+68h] [rbp-178h]
  _QWORD *v839; // [rsp+70h] [rbp-170h]
  int *v840; // [rsp+78h] [rbp-168h]
  int *v841; // [rsp+88h] [rbp-158h]
  int *v842; // [rsp+90h] [rbp-150h]
  int *v843; // [rsp+98h] [rbp-148h]
  int *v844; // [rsp+A0h] [rbp-140h]
  _QWORD *v845; // [rsp+B0h] [rbp-130h]
  int *v846; // [rsp+B8h] [rbp-128h]
  int *v847; // [rsp+C0h] [rbp-120h]
  int *v848; // [rsp+C8h] [rbp-118h]
  int *v849; // [rsp+D0h] [rbp-110h]
  int *v850; // [rsp+D8h] [rbp-108h]
  _QWORD *v851; // [rsp+E0h] [rbp-100h]
  int *v852; // [rsp+E8h] [rbp-F8h]
  int *v853; // [rsp+F0h] [rbp-F0h]
  _QWORD *v854; // [rsp+F8h] [rbp-E8h]
  _QWORD *v855; // [rsp+100h] [rbp-E0h]
  int *v856; // [rsp+108h] [rbp-D8h]
  int *v857; // [rsp+110h] [rbp-D0h]
  _QWORD *v858; // [rsp+118h] [rbp-C8h]
  _QWORD *v859; // [rsp+120h] [rbp-C0h]
  int *v860; // [rsp+128h] [rbp-B8h]
  int *v861; // [rsp+130h] [rbp-B0h]
  _QWORD *v862; // [rsp+138h] [rbp-A8h]
  _QWORD v863[2]; // [rsp+160h] [rbp-80h] BYREF
  __int64 v864; // [rsp+170h] [rbp-70h]
  int *v865; // [rsp+178h] [rbp-68h]
  int *v866; // [rsp+180h] [rbp-60h]
  _QWORD *v867; // [rsp+188h] [rbp-58h]
  _QWORD *v868; // [rsp+190h] [rbp-50h]
  int *v869; // [rsp+198h] [rbp-48h]
  _QWORD *v870; // [rsp+1A8h] [rbp-38h]
  int *v871; // [rsp+1B0h] [rbp-30h]
  _QWORD *v872; // [rsp+1B8h] [rbp-28h]
  _QWORD *v873; // [rsp+1C0h] [rbp-20h]
  int *v874; // [rsp+1C8h] [rbp-18h]
  int *v875; // [rsp+1D0h] [rbp-10h]
  int *v876; // [rsp+1D8h] [rbp-8h]
  _QWORD *v877; // [rsp+1E0h] [rbp+0h]
  int *v878; // [rsp+1E8h] [rbp+8h]
  _QWORD *v879; // [rsp+1F0h] [rbp+10h]
  int *v880; // [rsp+1F8h] [rbp+18h]
  _QWORD *v881; // [rsp+200h] [rbp+20h]
  _QWORD *v882; // [rsp+208h] [rbp+28h]
  int *v883; // [rsp+210h] [rbp+30h]
  int *v884; // [rsp+218h] [rbp+38h]
  _QWORD *v885; // [rsp+220h] [rbp+40h]
  _QWORD *v886; // [rsp+228h] [rbp+48h]
  _QWORD *v887; // [rsp+230h] [rbp+50h]
  int *v888; // [rsp+238h] [rbp+58h]
  int *v889; // [rsp+240h] [rbp+60h]
  int *v890; // [rsp+248h] [rbp+68h]
  _QWORD *v891; // [rsp+250h] [rbp+70h]
  _QWORD *v892; // [rsp+258h] [rbp+78h]
  _QWORD *v893; // [rsp+260h] [rbp+80h]
  _QWORD *v894; // [rsp+268h] [rbp+88h]
  _QWORD *v895; // [rsp+270h] [rbp+90h]
  int *v896; // [rsp+278h] [rbp+98h]
  int *v897; // [rsp+280h] [rbp+A0h]
  int *v898; // [rsp+288h] [rbp+A8h]
  int *v899; // [rsp+290h] [rbp+B0h]
  int *v900; // [rsp+298h] [rbp+B8h]
  int *v901; // [rsp+2A0h] [rbp+C0h]
  int *v902; // [rsp+2A8h] [rbp+C8h]
  int *v903; // [rsp+2B0h] [rbp+D0h]
  int *v904; // [rsp+2B8h] [rbp+D8h]
  int *v905; // [rsp+2C0h] [rbp+E0h]
  _QWORD *v906; // [rsp+2C8h] [rbp+E8h]
  int *v907; // [rsp+2D0h] [rbp+F0h]
  int *v908; // [rsp+2D8h] [rbp+F8h]
  int *v909; // [rsp+2E0h] [rbp+100h]
  int *v910; // [rsp+2E8h] [rbp+108h]
  int *v911; // [rsp+2F0h] [rbp+110h]
  int *v912; // [rsp+2F8h] [rbp+118h]
  int *v913; // [rsp+300h] [rbp+120h]
  int *v914; // [rsp+308h] [rbp+128h]
  _QWORD *v915; // [rsp+310h] [rbp+130h]
  int *v916; // [rsp+318h] [rbp+138h]
  int *v917; // [rsp+320h] [rbp+140h]
  __int64 v918; // [rsp+328h] [rbp+148h] BYREF
  _QWORD *v919; // [rsp+330h] [rbp+150h]
  int *v920; // [rsp+338h] [rbp+158h]
  int *v921; // [rsp+340h] [rbp+160h]
  int *v922; // [rsp+348h] [rbp+168h]
  int *v923; // [rsp+350h] [rbp+170h]
  __int64 v924; // [rsp+358h] [rbp+178h] BYREF
  int *v925; // [rsp+360h] [rbp+180h]
  int *v926; // [rsp+368h] [rbp+188h]
  _QWORD *v927; // [rsp+370h] [rbp+190h]
  int *v928; // [rsp+378h] [rbp+198h]
  int *v929; // [rsp+380h] [rbp+1A0h]
  _QWORD *v930; // [rsp+388h] [rbp+1A8h]
  int *v931; // [rsp+390h] [rbp+1B0h] BYREF
  __int64 v932; // [rsp+398h] [rbp+1B8h] BYREF
  int *v933; // [rsp+3A0h] [rbp+1C0h]
  __int64 v934; // [rsp+3A8h] [rbp+1C8h] BYREF
  int *v935; // [rsp+3B0h] [rbp+1D0h]
  _QWORD *v936; // [rsp+3B8h] [rbp+1D8h]
  int *v937; // [rsp+3C0h] [rbp+1E0h]
  int *v938; // [rsp+3C8h] [rbp+1E8h]
  __int64 v939; // [rsp+3D0h] [rbp+1F0h] BYREF
  int *v940; // [rsp+3D8h] [rbp+1F8h] BYREF
  __int64 v941; // [rsp+3E0h] [rbp+200h] BYREF
  __int64 v942; // [rsp+3E8h] [rbp+208h] BYREF
  __int64 v943; // [rsp+3F0h] [rbp+210h] BYREF
  _QWORD v944[2]; // [rsp+3F8h] [rbp+218h] BYREF
  _QWORD v945[2]; // [rsp+408h] [rbp+228h] BYREF
  _QWORD v946[2]; // [rsp+418h] [rbp+238h] BYREF
  _QWORD v947[2]; // [rsp+428h] [rbp+248h] BYREF
  _QWORD v948[2]; // [rsp+438h] [rbp+258h] BYREF
  _QWORD v949[2]; // [rsp+448h] [rbp+268h] BYREF
  _QWORD v950[3]; // [rsp+458h] [rbp+278h] BYREF
  _QWORD v951[3]; // [rsp+470h] [rbp+290h] BYREF
  _QWORD v952[3]; // [rsp+488h] [rbp+2A8h] BYREF
  _QWORD v953[3]; // [rsp+4A0h] [rbp+2C0h] BYREF
  _QWORD v954[3]; // [rsp+4B8h] [rbp+2D8h] BYREF
  _QWORD v955[3]; // [rsp+4D0h] [rbp+2F0h] BYREF
  _QWORD v956[3]; // [rsp+4E8h] [rbp+308h] BYREF
  _QWORD v957[4]; // [rsp+500h] [rbp+320h] BYREF
  _QWORD v958[6]; // [rsp+520h] [rbp+340h] BYREF

  v5 = sub_18036D7E0(*a3);
  v6 = a3[1];
  v7 = v5;
  v871 = (int *)v5;
  v8 = sub_18036D7E0(v6);
  v893 = (_QWORD *)v8;
  v910 = (int *)a3[2];
  v920 = (int *)a3[3];
  v921 = (int *)a3[4];
  v916 = (int *)a3[5];
  v914 = (int *)a3[6];
  v922 = (int *)a3[7];
  v9 = (int *)a3[8];
  v905 = nullptr;
  v882 = nullptr;
  v875 = nullptr;
  v923 = v9;
  v915 = (_QWORD *)sub_18036D6D0();
  v907 = nullptr;
  v10 = (_QWORD *)sub_18036D6D0();
  v867 = v10;
  v908 = nullptr;
  v11 = nullptr;
  v909 = nullptr;
  v876 = nullptr;
  v883 = nullptr;
  v865 = nullptr;
  v881 = nullptr;
  v904 = nullptr;
  v884 = nullptr;
  v901 = nullptr;
  v899 = nullptr;
  v913 = nullptr;
  v895 = nullptr;
  v890 = nullptr;
  v866 = nullptr;
  v891 = nullptr;
  v892 = nullptr;
  v12 = sub_18036D6D0();
  v13 = qword_1803BDD90;
  v902 = (int *)v12;
  v863[0] = 0;
  v912 = nullptr;
  v872 = nullptr;
  v877 = nullptr;
  v903 = nullptr;
  v898 = nullptr;
  v885 = nullptr;
  v906 = nullptr;
  v873 = nullptr;
  v886 = nullptr;
  v911 = nullptr;
  v868 = nullptr;
  v917 = nullptr;
  v889 = nullptr;
  v900 = nullptr;
  v896 = nullptr;
  v887 = nullptr;
  v897 = nullptr;
  v878 = nullptr;
  v874 = nullptr;
  v880 = nullptr;
  v879 = nullptr;
  v894 = nullptr;
  v864 = qword_1803BDD90;
  if ( qword_1803BDD90 != 0 )
  {
    if ( *(__int64 *)qword_1803BDD90 <= 1 && *(_QWORD *)(qword_1803BDD90 + 16) == 0 )
      goto LABEL_9;
    if ( qword_1803BDD90 != 0 && *(int *)qword_1803BDD90 >= 0 )
    {
      v142 = (*(_QWORD *)qword_1803BDD90)-- == 1;
      if ( v142 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v13 + 8) + 48LL))(v13);
    }
  }
  v13 = sub_18036CD10(a1, *(_QWORD *)&qword_1803BDDE0, qword_1803CA5C0, 336);
  v864 = v13;
  qword_1803BDD90 = v13;
LABEL_9:
  sub_180001D00(a1, v13);
  v14 = sub_18037F0C0(a1, *(_QWORD *)(v7 + 16), qword_1803BDF10);
  v15 = (int *)v14;
  if ( v14 == 0 )
  {
    v863[0] = *(_QWORD *)(a1 + 96);
    v16 = 71;
    *(_QWORD *)(a1 + 96) = 0;
    v17 = nullptr;
    goto LABEL_513;
  }
  v18 = *(_QWORD *)(v8 + 16);
  v16 = 71;
  *(_DWORD *)(v13 + 40) = 71;
  v19 = sub_18037E6C0(a1, v14, qword_1803BDF80, v18);
  if ( *v15 >= 0 )
  {
    v142 = (*(_QWORD *)v15)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v15 + 1) + 48LL))(v15);
  }
  if ( v19 == 0 )
  {
    v863[0] = *(_QWORD *)(a1 + 96);
    v17 = nullptr;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_513;
  }
  v20 = v871;
  v16 = 72;
  v21 = v920;
  v10[2] = v19;
  v22 = *((_QWORD *)v20 + 2);
  *(_DWORD *)(v13 + 40) = 72;
  v23 = sub_18037E6C0(a1, v22, *(_QWORD *)&qword_1803BDF88, v21);
  v933 = (int *)v23;
  v24 = (int *)v23;
  if ( v23 == 0 )
  {
    v863[0] = *(_QWORD *)(a1 + 96);
    v11 = nullptr;
    *(_QWORD *)(a1 + 96) = 0;
    v17 = nullptr;
    goto LABEL_513;
  }
  v25 = qword_1803BDF90;
  v908 = (int *)v23;
  v909 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v909, v25, v922);
  PyDict_SetItem(v909, qword_1803BDF98, v923);
  v26 = sub_18037F0C0(a1, v10[2], qword_1803BDFA0);
  v27 = (int *)v26;
  if ( v26 == 0 )
  {
    v863[0] = *(_QWORD *)(a1 + 96);
    v16 = 77;
    *(_QWORD *)(a1 + 96) = 0;
    v11 = nullptr;
    v17 = nullptr;
    goto LABEL_513;
  }
  v28 = sub_1803836D0(v26, qword_1803BDFA8);
  if ( *v27 >= 0 )
  {
    v142 = (*(_QWORD *)v27)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v27 + 1) + 48LL))(v27);
  }
  if ( v28 == -1 )
  {
    v863[0] = *(_QWORD *)(a1 + 96);
    v16 = 77;
    *(_QWORD *)(a1 + 96) = 0;
    v11 = nullptr;
    v17 = nullptr;
    goto LABEL_513;
  }
  if ( v28 == 1 )
  {
    v29 = sub_18037F0C0(a1, v10[2], qword_1803BDFB0);
    v30 = (int *)v29;
    if ( v29 == 0 )
    {
      sub_18000CE30(a1, v863);
      v11 = v865;
      v16 = 77;
      v17 = v865;
      goto LABEL_513;
    }
    v31 = sub_180001510(v29);
    if ( v31 == -1 )
    {
      sub_18000CE30(a1, v863);
      if ( *v30 >= 0 )
      {
        v142 = (*(_QWORD *)v30)-- == 1;
        if ( v142 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v30 + 1) + 48LL))(v30);
      }
      v11 = v865;
      v16 = 77;
      v17 = v865;
      goto LABEL_513;
    }
    if ( *v30 >= 0 )
    {
      v142 = (*(_QWORD *)v30)-- == 1;
      if ( v142 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v30 + 1) + 48LL))(v30);
    }
    if ( v31 != 0 )
    {
      v32 = v10[2];
      v33 = v871;
      v16 = 78;
      v34 = v910;
      v35 = *((_QWORD *)v871 + 2);
      *(_DWORD *)(v13 + 40) = 78;
      v950[0] = v32;
      v950[1] = v34;
      v950[2] = v24;
      v37 = (int *)sub_18037E980(a1, v35, *(_QWORD *)&qword_1803BDFB8, v950);
      v38 = a1;
      if ( v37 == nullptr )
        goto LABEL_38;
      sub_180001D80(a1, v36);
LABEL_1287:
      v736 = v915;
      if ( *(int *)v915 >= 0 )
      {
        v142 = (*v915)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v736[1] + 48LL))();
      }
      v737 = v907;
      if ( v907 != nullptr && *v907 >= 0 )
      {
        v142 = (*(_QWORD *)v907)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v737 + 1) + 48LL))();
      }
      v738 = v867;
      if ( *(int *)v867 >= 0 )
      {
        v142 = (*v867)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v738[1] + 48LL))();
      }
      if ( *v24 >= 0 )
      {
        v142 = (*(_QWORD *)v24)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v908 + 1) + 48LL))();
      }
      v739 = v909;
      if ( v909 != nullptr && *v909 >= 0 )
      {
        v142 = (*(_QWORD *)v909)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v739 + 1) + 48LL))();
      }
      v740 = v876;
      if ( v876 != nullptr && *v876 >= 0 )
      {
        v142 = (*(_QWORD *)v876)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v740 + 1) + 48LL))();
      }
      v741 = v883;
      if ( v883 != nullptr && *v883 >= 0 )
      {
        v142 = (*(_QWORD *)v883)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v741 + 1) + 48LL))();
      }
      v742 = v865;
      if ( v865 != nullptr && *v865 >= 0 )
      {
        v142 = (*(_QWORD *)v865)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v742 + 1) + 48LL))();
      }
      v743 = v881;
      if ( v881 != nullptr && *(int *)v881 >= 0 )
      {
        v142 = (*v881)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v743[1] + 48LL))();
      }
      v744 = v904;
      if ( v904 != nullptr && *v904 >= 0 )
      {
        v142 = (*(_QWORD *)v904)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v744 + 1) + 48LL))();
      }
      v745 = v884;
      if ( v884 != nullptr && *v884 >= 0 )
      {
        v142 = (*(_QWORD *)v884)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v745 + 1) + 48LL))();
      }
      v746 = v901;
      if ( v901 != nullptr && *v901 >= 0 )
      {
        v142 = (*(_QWORD *)v901)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v746 + 1) + 48LL))();
      }
      v747 = v899;
      if ( v899 != nullptr && *v899 >= 0 )
      {
        v142 = (*(_QWORD *)v899)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v747 + 1) + 48LL))();
      }
      v748 = v913;
      if ( v913 != nullptr && *v913 >= 0 )
      {
        v142 = (*(_QWORD *)v913)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v748 + 1) + 48LL))();
      }
      v749 = v895;
      if ( v895 != nullptr && *(int *)v895 >= 0 )
      {
        v142 = (*v895)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v749[1] + 48LL))();
      }
      v750 = v890;
      if ( v890 != nullptr && *v890 >= 0 )
      {
        v142 = (*(_QWORD *)v890)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v750 + 1) + 48LL))();
      }
      v751 = v866;
      if ( v866 != nullptr && *v866 >= 0 )
      {
        v142 = (*(_QWORD *)v866)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v751 + 1) + 48LL))();
      }
      v752 = v891;
      if ( v891 != nullptr && *(int *)v891 >= 0 )
      {
        v142 = (*v891)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v752[1] + 48LL))();
      }
      v753 = v892;
      if ( v892 != nullptr && *(int *)v892 >= 0 )
      {
        v142 = (*v892)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v753[1] + 48LL))();
      }
      v754 = v902;
      if ( *v902 >= 0 )
      {
        v142 = (*(_QWORD *)v902)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v754 + 1) + 48LL))();
      }
      v755 = v912;
      if ( v912 != nullptr && *v912 >= 0 )
      {
        v142 = (*(_QWORD *)v912)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v755 + 1) + 48LL))();
      }
      v756 = v872;
      if ( v872 != nullptr && *(int *)v872 >= 0 )
      {
        v142 = (*v872)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v756[1] + 48LL))();
      }
      v757 = v877;
      if ( v877 != nullptr && *(int *)v877 >= 0 )
      {
        v142 = (*v877)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v757[1] + 48LL))();
      }
      v758 = v903;
      if ( v903 != nullptr && *v903 >= 0 )
      {
        v142 = (*(_QWORD *)v903)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v758 + 1) + 48LL))();
      }
      v759 = v898;
      if ( v898 != nullptr && *v898 >= 0 )
      {
        v142 = (*(_QWORD *)v898)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v759 + 1) + 48LL))();
      }
      v760 = v885;
      if ( v885 != nullptr && *(int *)v885 >= 0 )
      {
        v142 = (*v885)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v760[1] + 48LL))();
      }
      v761 = v906;
      if ( v906 != nullptr && *(int *)v906 >= 0 )
      {
        v142 = (*v906)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v761[1] + 48LL))();
      }
      v762 = v873;
      if ( v873 != nullptr && *(int *)v873 >= 0 )
      {
        v142 = (*v873)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v762[1] + 48LL))();
      }
      v763 = v886;
      if ( v886 != nullptr && *(int *)v886 >= 0 )
      {
        v142 = (*v886)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v763[1] + 48LL))();
      }
      v764 = v911;
      if ( v911 != nullptr && *v911 >= 0 )
      {
        v142 = (*(_QWORD *)v911)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v764 + 1) + 48LL))();
      }
      if ( *v33 >= 0 )
      {
        v142 = (*(_QWORD *)v33)-- == 1;
        if ( v142 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v33 + 1) + 48LL))(v33);
      }
      v765 = v893;
      if ( *(int *)v893 >= 0 )
      {
        v142 = (*v893)-- == 1;
        if ( v142 )
          (*(void (**)(void))(v765[1] + 48LL))();
      }
      if ( *v34 >= 0 )
      {
        v142 = (*(_QWORD *)v34)-- == 1;
        if ( v142 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v34 + 1) + 48LL))(v34);
      }
      v766 = v920;
      if ( *v920 >= 0 )
      {
        v142 = (*(_QWORD *)v920)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v766 + 1) + 48LL))();
      }
      v767 = v921;
      if ( *v921 >= 0 )
      {
        v142 = (*(_QWORD *)v921)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v767 + 1) + 48LL))();
      }
      v768 = v916;
      if ( *v916 >= 0 )
      {
        v142 = (*(_QWORD *)v916)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v768 + 1) + 48LL))();
      }
      v769 = v914;
      if ( *v914 >= 0 )
      {
        v142 = (*(_QWORD *)v914)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v769 + 1) + 48LL))();
      }
      v770 = v922;
      if ( *v922 >= 0 )
      {
        v142 = (*(_QWORD *)v922)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v770 + 1) + 48LL))();
      }
      v771 = v923;
      if ( *v923 >= 0 )
      {
        v142 = (*(_QWORD *)v923)-- == 1;
        if ( v142 )
          (*(void (**)(void))(*((_QWORD *)v771 + 1) + 48LL))();
      }
      return v37;
    }
  }
  if ( Py_NoneStruct != -1 )
    ++Py_NoneStruct;
  v39 = v10[2];
  v876 = (int *)Py_NoneStruct;
  v40 = sub_18037F0C0(a1, v39, qword_1803BDFC0);
  v41 = (int *)v40;
  if ( v40 == 0 )
  {
    sub_18000CE30(a1, v863);
    v11 = v865;
    v16 = 80;
    v17 = v865;
    goto LABEL_513;
  }
  v42 = sub_180001510(v40);
  if ( v42 == -1 )
  {
    sub_18000CE30(a1, v863);
    if ( *v41 >= 0 )
    {
      v142 = (*(_QWORD *)v41)-- == 1;
      if ( v142 )
      {
        v43 = sub_18000CC40(v41);
        (*(void (__fastcall **)(int *))(v43 + 48))(v41);
      }
    }
    v11 = v865;
    v16 = 80;
    v17 = v865;
    goto LABEL_513;
  }
  if ( *v41 >= 0 )
  {
    v142 = (*(_QWORD *)v41)-- == 1;
    if ( v142 )
    {
      v44 = sub_18000CC40(v41);
      (*(void (__fastcall **)(int *))(v44 + 48))(v41);
    }
  }
  v45 = v871;
  if ( v42 != 0 )
  {
    v46 = sub_18037F0C0(a1, *((_QWORD *)v871 + 2), qword_1803BDF18);
    v47 = (int *)v46;
    if ( v46 == 0 )
      goto LABEL_72;
    v48 = (int *)sub_18037F0C0(a1, v46, qword_1803BDFC8);
    if ( *v47 >= 0 )
    {
      v142 = (*(_QWORD *)v47)-- == 1;
      if ( v142 )
      {
        v49 = sub_18000CC40(v47);
        (*(void (__fastcall **)(int *))(v49 + 48))(v47);
      }
    }
    if ( v48 == nullptr )
      goto LABEL_72;
    v50 = sub_18037F0C0(a1, v10[2], qword_1803BDFC0);
    v51 = (int *)v50;
    if ( v50 == 0 )
    {
      sub_18000CE30(a1, v863);
      if ( *v48 >= 0 )
      {
        v142 = (*(_QWORD *)v48)-- == 1;
        if ( v142 )
        {
          v52 = sub_18000CC40(v48);
          (*(void (__fastcall **)(int *))(v52 + 48))(v48);
        }
      }
      goto LABEL_73;
    }
    *(_DWORD *)(v13 + 40) = 82;
    v53 = sub_180376770(a1, v48, v50);
    if ( *v48 >= 0 )
    {
      v142 = (*(_QWORD *)v48)-- == 1;
      if ( v142 )
      {
        v54 = sub_18000CC40(v48);
        (*(void (__fastcall **)(int *))(v54 + 48))(v48);
      }
    }
    if ( *v51 >= 0 )
    {
      v142 = (*(_QWORD *)v51)-- == 1;
      if ( v142 )
      {
        v55 = sub_18000CC40(v51);
        (*(void (__fastcall **)(int *))(v55 + 48))(v51);
      }
    }
    if ( v53 == 0 )
    {
LABEL_72:
      sub_18000CE30(a1, v863);
LABEL_73:
      v924 = v863[0];
      sub_18000CE70(v863);
      v932 = sub_18000CCC0(a1);
      v57 = sub_18000CE80(&v924);
      if ( v57 != 0 )
        v58 = sub_18000CC90(v57, v13, 82);
      else
        v58 = sub_18038E2A0(v13, 82);
      sub_180001300(&v924, v58);
      sub_18000D050(a1, &v924);
      v59 = sub_18000CC40(*(_QWORD *)(a1 + 248));
      if ( (unsigned int)sub_18000CE90(a1, v59, PyExc_Exception) == 0 )
      {
        v16 = 81;
        if ( (unsigned __int8)sub_180373290(a1, v863) != 0 )
          v16 = 0;
        v61 = sub_18000CE80(v863);
        if ( v61 != 0 && *(_QWORD *)(v61 + 24) == v13 )
          *(_DWORD *)(v13 + 40) = *(_DWORD *)(v61 + 36);
        v62 = v863[0];
        sub_18000CE70(v863);
        sub_18000CCE0(a1, &v932);
        v11 = v865;
        v17 = v865;
        goto LABEL_511;
      }
      v876 = (int *)Py_NoneStruct;
      if ( Py_NoneStruct != -1 )
        ++Py_NoneStruct;
      if ( Py_NoneStruct >= 0 && --Py_NoneStruct == 1 )
      {
        v60 = sub_18000CC40(Py_NoneStruct);
        (*(void (__fastcall **)(_QWORD))(v60 + 48))(Py_NoneStruct);
      }
      sub_18000CCE0(a1, &v932);
      v53 = (__int64)v876;
      goto LABEL_90;
    }
    v876 = (int *)v53;
    if ( Py_NoneStruct >= 0 && --Py_NoneStruct == 1 )
    {
      v56 = sub_18000CC40(Py_NoneStruct);
      (*(void (__fastcall **)(_QWORD))(v56 + 48))(Py_NoneStruct);
    }
  }
  else
  {
    v53 = Py_NoneStruct;
  }
LABEL_90:
  v937 = (int *)v53;
  v63 = sub_1801F3CE0(a1);
  if ( v63 == 0 )
  {
    sub_180373340(a1, v863, qword_1803BDFD0);
    v11 = v865;
    v16 = 86;
    v17 = v865;
    goto LABEL_512;
  }
  v64 = sub_18037F0C0(a1, *((_QWORD *)v45 + 2), qword_1803BDF20);
  v65 = (int *)v64;
  if ( v64 == 0 )
  {
    sub_18000CE30(a1, v863);
    v11 = v865;
    v16 = 86;
    v17 = v865;
    goto LABEL_512;
  }
  v16 = 86;
  *(_DWORD *)(v13 + 40) = 86;
  v66 = sub_180376770(a1, v63, v64);
  v67 = *v65 < 0;
  v68 = (int *)v66;
  v888 = (int *)v66;
  if ( !v67 )
  {
    v142 = (*(_QWORD *)v65)-- == 1;
    if ( v142 )
    {
      v69 = sub_18000CC40(v65);
      (*(void (__fastcall **)(int *))(v69 + 48))(v65);
    }
  }
  v38 = a1;
  if ( v68 == nullptr )
  {
LABEL_38:
    sub_18000CE30(v38, v863);
    v11 = v865;
    v17 = v865;
    goto LABEL_512;
  }
  v883 = v68;
  v70 = v68;
  v71 = sub_180372160(a1, 0);
  v905 = (int *)v71;
  if ( v921 == (int *)Py_NoneStruct )
    goto LABEL_233;
  v72 = sub_180001930(a1, v921);
  v73 = (int *)v72;
  if ( v72 == 0 )
  {
    sub_18000CE30(a1, v863);
    v11 = v865;
    v16 = 89;
    v17 = v865;
    goto LABEL_512;
  }
  v74 = (int *)sub_18000D430(v72);
  if ( v74 == nullptr )
    goto LABEL_164;
  v75 = (__int64)v881;
  v16 = 91;
  do
  {
    v76 = v889;
    v77 = (int *)v75;
    v78 = v865;
    v79 = v889;
    v889 = v74;
    v80 = v865;
    if ( v79 != nullptr && *v76 >= 0 )
    {
      v142 = (*(_QWORD *)v76)-- == 1;
      if ( v142 )
      {
        v81 = sub_18000CC40(v76);
        (*(void (__fastcall **)(int *))(v81 + 48))(v79);
        v78 = v80;
      }
    }
    v82 = *v74;
    v83 = v78;
    v865 = v74;
    v84 = v82 + 1;
    if ( v84 != 0 )
      *v74 = v84;
    if ( v78 != nullptr && *v80 >= 0 )
    {
      v142 = (*(_QWORD *)v80)-- == 1;
      if ( v142 )
      {
        v85 = sub_18000CC40(v78);
        (*(void (__fastcall **)(int *))(v85 + 48))(v83);
      }
    }
    IsInstance = PyObject_IsInstance(v74, PyDict_Type);
    if ( IsInstance == -1 )
    {
      sub_18000CE30(a1, v863);
      v16 = 90;
      goto LABEL_253;
    }
    if ( IsInstance != 0 )
    {
      v87 = sub_18037F0C0(a1, v74, qword_1803BDFD8);
      v88 = (int *)v87;
      v89 = a1;
      if ( v87 == 0 )
        goto LABEL_178;
      v90 = v864;
      *(_DWORD *)(v864 + 40) = 91;
      v91 = (int *)sub_1803770D0(a1, v87, qword_1803BDFE0);
      if ( *v88 >= 0 )
      {
        v142 = (*(_QWORD *)v88)-- == 1;
        if ( v142 )
        {
          v92 = sub_18000CC40(v88);
          (*(void (__fastcall **)(int *))(v92 + 48))(v88);
        }
      }
      if ( v91 == nullptr )
      {
LABEL_177:
        sub_18000CE30(a1, v863);
        goto LABEL_253;
      }
      v93 = sub_180001510(v91);
      if ( v93 == -1 )
      {
        sub_18000CE30(a1, v863);
        if ( *v91 >= 0 )
        {
          v142 = (*(_QWORD *)v91)-- == 1;
          if ( v142 )
          {
            v115 = sub_18000CC40(v91);
            (*(void (__fastcall **)(int *))(v115 + 48))(v91);
          }
        }
        goto LABEL_253;
      }
      if ( v93 != 1 )
      {
        if ( *v91 >= 0 )
        {
          v142 = (*(_QWORD *)v91)-- == 1;
          if ( v142 )
          {
            v94 = sub_18000CC40(v91);
            (*(void (__fastcall **)(int *))(v94 + 48))(v91);
          }
        }
        v91 = (int *)qword_1803BDFE8;
        if ( *(_DWORD *)qword_1803BDFE8 != -1 )
          ++*(_DWORD *)qword_1803BDFE8;
      }
      v95 = (int *)sub_1803749A0(v91);
      if ( *v91 >= 0 )
      {
        v142 = (*(_QWORD *)v91)-- == 1;
        if ( v142 )
        {
          v96 = sub_18000CC40(v91);
          (*(void (__fastcall **)(int *))(v96 + 48))(v91);
        }
      }
      v89 = a1;
      if ( v95 == nullptr )
        goto LABEL_178;
      v97 = (int *)sub_18037F0C0(a1, v95, qword_1803BDFF0);
      if ( *v95 >= 0 )
      {
        v142 = (*(_QWORD *)v95)-- == 1;
        if ( v142 )
        {
          v98 = sub_18000CC40(v95);
          (*(void (__fastcall **)(int *))(v98 + 48))(v95);
        }
      }
      v89 = a1;
      if ( v97 == nullptr )
      {
LABEL_178:
        sub_18000CE30(v89, v863);
        goto LABEL_253;
      }
      *(_DWORD *)(v90 + 40) = 91;
      v75 = sub_180376210(a1, v97);
      if ( *v97 >= 0 )
      {
        v142 = (*(_QWORD *)v97)-- == 1;
        if ( v142 )
        {
          v99 = sub_18000CC40(v97);
          (*(void (__fastcall **)(int *))(v99 + 48))(v97);
        }
      }
      if ( v75 == 0 )
        goto LABEL_177;
      v100 = v881;
      v881 = (_QWORD *)v75;
    }
    else
    {
      v101 = sub_180001510(v74);
      if ( v101 == -1 )
        goto LABEL_182;
      v102 = qword_1803BDFE8;
      if ( v101 == 1 )
        v102 = (__int64)v74;
      v103 = sub_1803749A0(v102);
      v104 = (int *)v103;
      v105 = a1;
      if ( v103 == 0 )
        goto LABEL_181;
      v106 = (int *)sub_18037F0C0(a1, v103, qword_1803BDFF0);
      if ( *v104 >= 0 )
      {
        v142 = (*(_QWORD *)v104)-- == 1;
        if ( v142 )
        {
          v107 = sub_18000CC40(v104);
          (*(void (__fastcall **)(int *))(v107 + 48))(v104);
        }
      }
      v105 = a1;
      if ( v106 == nullptr )
      {
LABEL_181:
        sub_18000CE30(v105, v863);
        v16 = 93;
        goto LABEL_253;
      }
      *(_DWORD *)(v864 + 40) = 93;
      v108 = sub_180376210(a1, v106);
      if ( *v106 >= 0 )
      {
        v142 = (*(_QWORD *)v106)-- == 1;
        if ( v142 )
        {
          v109 = sub_18000CC40(v106);
          (*(void (__fastcall **)(int *))(v109 + 48))(v106);
        }
      }
      if ( v108 == 0 )
      {
LABEL_182:
        sub_18000CE30(a1, v863);
        v16 = 93;
        goto LABEL_253;
      }
      v100 = v881;
      v75 = v108;
      v881 = (_QWORD *)v108;
    }
    v110 = v100;
    if ( v100 != nullptr && *v77 >= 0 )
    {
      v142 = (*(_QWORD *)v77)-- == 1;
      if ( v142 )
      {
        v111 = sub_18000CC40(v100);
        (*(void (__fastcall **)(_QWORD *))(v111 + 48))(v110);
      }
    }
    v112 = sub_180001510(v75);
    if ( v112 == -1 )
    {
      sub_18000CE30(a1, v863);
      v16 = 94;
      goto LABEL_253;
    }
    v71 = (__int64)v905;
    if ( v112 != 0 )
    {
      if ( v905 == nullptr )
      {
        sub_180370FE0(a1, v863, qword_1803BDFF8);
        sub_18000D1D0(a1, v863[0]);
        v16 = 95;
        goto LABEL_253;
      }
      sub_1803725B0(v905, v75);
    }
    if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
      goto LABEL_252;
    v74 = (int *)sub_18000D430(v73);
  }
  while ( v74 != nullptr );
  v68 = v888;
  v70 = v888;
LABEL_164:
  if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
  {
LABEL_252:
    sub_18000CE30(a1, v863);
    v16 = 89;
LABEL_253:
    v62 = v863[0];
    sub_18000CE70(v863);
    v11 = v865;
    if ( v865 != nullptr && *v865 >= 0 )
    {
      v142 = (*(_QWORD *)v865)-- == 1;
      if ( v142 )
      {
        v151 = sub_18000CC40(v11);
        (*(void (__fastcall **)(int *))(v151 + 48))(v11);
      }
    }
    if ( *v73 >= 0 )
    {
      v142 = (*(_QWORD *)v73)-- == 1;
      if ( v142 )
      {
        v152 = sub_18000CC40(v73);
        (*(void (__fastcall **)(int *))(v152 + 48))(v73);
      }
    }
    goto LABEL_510;
  }
  v11 = v865;
  if ( v865 != nullptr && *v865 >= 0 )
  {
    v142 = (*(_QWORD *)v865)-- == 1;
    if ( v142 )
    {
      v113 = sub_18000CC40(v11);
      (*(void (__fastcall **)(int *))(v113 + 48))(v11);
    }
  }
  if ( *v73 >= 0 )
  {
    v142 = (*(_QWORD *)v73)-- == 1;
    if ( v142 )
    {
      v114 = sub_18000CC40(v73);
      (*(void (__fastcall **)(int *))(v114 + 48))(v73);
    }
  }
  if ( v71 == 0 )
  {
    sub_180370FE0(a1, v863, qword_1803BDFF8);
    sub_18000D1D0(a1, v863[0]);
    v17 = v866;
    v16 = 96;
    goto LABEL_512;
  }
  v116 = sub_180001510(v71);
  if ( v116 == -1 )
  {
    sub_18000CE30(a1, v863);
    v17 = v866;
    v16 = 96;
    goto LABEL_512;
  }
  if ( v116 == 0 )
  {
    v141 = v70;
    v150 = sub_180372160(a1, 0);
    v67 = *v68 < 0;
    v70 = (int *)v150;
    v883 = (int *)v150;
    if ( !v67 )
    {
      v142 = (*(_QWORD *)v68)-- == 1;
      goto LABEL_230;
    }
LABEL_232:
    v13 = v864;
LABEL_233:
    v935 = v904;
    v936 = v881;
    v897 = v70;
    v144 = sub_1801F39B0(a1);
    if ( v144 == 0 )
    {
      sub_180373340(a1, v863, qword_1803BE008);
      v11 = v865;
      v16 = 102;
      v17 = v866;
      goto LABEL_512;
    }
    v10 = v867;
    v153 = (int *)sub_18037F0C0(a1, v867[2], qword_1803BDFA0);
    if ( v153 == nullptr )
    {
      sub_18000CE30(a1, v863);
      v11 = v865;
      v16 = 103;
      v17 = v866;
      goto LABEL_513;
    }
    v154 = sub_180001510(v70);
    if ( v154 == -1 )
    {
      sub_18000CE30(a1, v863);
      if ( *v153 >= 0 )
      {
        v142 = (*(_QWORD *)v153)-- == 1;
        if ( v142 )
        {
          v155 = sub_18000CC40(v153);
          (*(void (__fastcall **)(int *))(v155 + 48))(v153);
        }
      }
      v11 = v865;
      v16 = 105;
      v17 = v866;
      goto LABEL_513;
    }
    v156 = (int *)Py_NoneStruct;
    if ( v154 == 1 )
      v156 = v70;
    v157 = sub_18037F0C0(a1, v867[2], qword_1803BDFA0);
    v158 = (int *)v157;
    if ( v157 != 0 )
    {
      v161 = (int *)sub_180383520(v157, qword_1803BDFA8);
      if ( *v158 >= 0 )
      {
        v142 = (*(_QWORD *)v158)-- == 1;
        if ( v142 )
        {
          v162 = sub_18000CC40(v158);
          (*(void (__fastcall **)(int *))(v162 + 48))(v158);
        }
      }
      if ( v161 != nullptr )
      {
        v163 = sub_180001510(v161);
        if ( v163 == -1 )
        {
          sub_18000CE30(a1, v863);
          if ( *v153 >= 0 )
          {
            v142 = (*(_QWORD *)v153)-- == 1;
            if ( v142 )
            {
              v164 = sub_18000CC40(v153);
              (*(void (__fastcall **)(int *))(v164 + 48))(v153);
            }
          }
          if ( *v161 < 0 )
            goto LABEL_274;
          v142 = (*(_QWORD *)v161)-- == 1;
          if ( !v142 )
            goto LABEL_274;
          v159 = sub_18000CC40(v161);
          v160 = v161;
          goto LABEL_273;
        }
        if ( v163 != 1 )
          goto LABEL_294;
        if ( *v161 >= 0 )
        {
          v142 = (*(_QWORD *)v161)-- == 1;
          if ( v142 )
          {
            v165 = sub_18000CC40(v161);
            (*(void (__fastcall **)(int *))(v165 + 48))(v161);
          }
        }
        v161 = (int *)sub_18037F0C0(a1, v867[2], qword_1803BDFB0);
        if ( v161 != nullptr )
        {
LABEL_294:
          v166 = sub_180001510(v161);
          if ( *v161 >= 0 )
          {
            v142 = (*(_QWORD *)v161)-- == 1;
            if ( v142 )
            {
              v167 = sub_18000CC40(v161);
              (*(void (__fastcall **)(int *))(v167 + 48))(v161);
            }
          }
          if ( v166 != -1 )
          {
            v168 = Py_FalseStruct;
            v169 = v876;
            v142 = v166 == 0;
            v16 = 102;
            if ( !v142 )
              v168 = Py_TrueStruct;
            *(_DWORD *)(v13 + 40) = 102;
            v944[1] = v168;
            v945[1] = v169;
            v945[0] = v153;
            v944[0] = v156;
            v170 = (int *)sub_18037D650(a1, v144, v945, v944, qword_1803BE010);
            if ( *v153 >= 0 )
            {
              v142 = (*(_QWORD *)v153)-- == 1;
              if ( v142 )
              {
                v171 = sub_18000CC40(v153);
                (*(void (__fastcall **)(int *))(v171 + 48))(v153);
              }
            }
            if ( v170 == nullptr )
            {
LABEL_1261:
              sub_18000CE30(a1, v863);
              v11 = v865;
              v17 = v866;
              goto LABEL_512;
            }
            v172 = v905;
            v173 = v170;
            v884 = v170;
            if ( v905 == nullptr )
            {
              sub_180370FE0(a1, v863, qword_1803BDFF8);
              sub_18000D1D0(a1, v863[0]);
              v11 = v865;
              v16 = 108;
              v17 = v866;
              goto LABEL_512;
            }
            v174 = sub_180001510(v905);
            if ( v174 == -1 )
            {
              sub_18000CE30(a1, v863);
              v11 = v865;
              v16 = 108;
              v17 = v866;
              goto LABEL_512;
            }
            if ( v174 != 0 )
            {
              v16 = 109;
              v175 = qword_1803BE018;
              v176 = qword_1803CA5B8;
              v828 = qword_1803BE028;
              v825 = qword_1803BE020;
              *(_DWORD *)(v13 + 40) = 109;
              v177 = sub_1803717E0(a1, v175, v176, Py_NoneStruct, v825, v828);
              v178 = (int *)v177;
              if ( v177 == 0 )
                goto LABEL_1261;
              if ( (unsigned int)sub_18000CC50(v177, PyModule_Type) != 0 )
                v179 = sub_180371D80(a1, (_DWORD)v178, qword_1803CA5B8, qword_1803BE030, qword_1803AB250);
              else
                v179 = sub_180371890(a1, v178, qword_1803BE030);
              v180 = (int *)v179;
              if ( *v178 >= 0 )
              {
                v142 = (*(_QWORD *)v178)-- == 1;
                if ( v142 )
                {
                  v181 = sub_18000CC40(v178);
                  (*(void (__fastcall **)(int *))(v181 + 48))(v178);
                }
              }
              if ( v180 == nullptr )
                goto LABEL_1261;
              v901 = v180;
              v182 = sub_18037F0C0(a1, *((_QWORD *)v871 + 2), qword_1803BDF20);
              v183 = (int *)v182;
              if ( v182 == 0 )
              {
                sub_18000CE30(a1, v863);
                v11 = v865;
                v16 = 111;
                v17 = v866;
                goto LABEL_512;
              }
              v16 = 111;
              *(_DWORD *)(v13 + 40) = 111;
              v946[0] = v182;
              v946[1] = v172;
              v184 = sub_180377750(a1, v180, v946);
              if ( *v183 >= 0 )
              {
                v142 = (*(_QWORD *)v183)-- == 1;
                if ( v142 )
                {
                  v185 = sub_18000CC40(v183);
                  (*(void (__fastcall **)(int *))(v185 + 48))(v183);
                }
              }
              if ( v184 == 0 )
                goto LABEL_1261;
              v899 = (int *)v184;
              v186 = sub_180001510(v184);
              if ( v186 == -1 )
              {
                sub_18000CE30(a1, v863);
                v11 = v865;
                v16 = 112;
                v17 = v866;
                goto LABEL_512;
              }
              if ( v186 != 0 )
              {
                v187 = sub_180375520(a1, v170, qword_1803BDFE8);
                if ( v187 == 0 )
                {
LABEL_326:
                  sub_18000CE30(a1, v863);
                  v11 = v865;
                  v16 = 113;
                  v17 = v866;
                  goto LABEL_512;
                }
                v188 = (int *)sub_180374A40(a1, 3);
                sub_18000CC80(v188, 0, v187);
                v189 = qword_1803BE038;
                if ( *(_DWORD *)qword_1803BE038 != -1 )
                  ++*(_DWORD *)qword_1803BE038;
                sub_18000CC80(v188, 1, v189);
                v190 = sub_180375520(a1, v184, qword_1803BDFE8);
                if ( v190 == 0 )
                {
                  sub_18000CE30(a1, v863);
                  v16 = 113;
                  if ( *v188 >= 0 )
                  {
                    v142 = (*(_QWORD *)v188)-- == 1;
                    if ( v142 )
                    {
                      v191 = sub_18000CC40(v188);
                      (*(void (__fastcall **)(int *))(v191 + 48))(v188);
                      v11 = v865;
                      v17 = v866;
                      goto LABEL_512;
                    }
                  }
                  goto LABEL_1271;
                }
                sub_18000CC80(v188, 2, v190);
                v192 = PyUnicode_Join(qword_1803BDFE8, v188);
                if ( *v188 >= 0 )
                {
                  v142 = (*(_QWORD *)v188)-- == 1;
                  if ( v142 )
                  {
                    v193 = sub_18000CC40(v188);
                    (*(void (__fastcall **)(int *))(v193 + 48))(v188);
                  }
                }
                if ( v192 == 0 )
                  goto LABEL_326;
                v67 = *v170 < 0;
                v173 = (int *)v192;
                v884 = (int *)v192;
                if ( !v67 )
                {
                  v142 = (*(_QWORD *)v170)-- == 1;
                  if ( v142 )
                  {
                    v194 = sub_18000CC40(v170);
                    (*(void (__fastcall **)(int *))(v194 + 48))(v170);
                  }
                }
              }
              else
              {
                v173 = v170;
              }
            }
            v195 = v173;
            v16 = 114;
            v196 = qword_1803BE040;
            v197 = qword_1803CA5B8;
            v926 = v899;
            v925 = v901;
            v829 = qword_1803AB250;
            v826 = qword_1803BE048;
            *(_DWORD *)(v13 + 40) = 114;
            v198 = sub_1803717E0(a1, v196, v197, Py_NoneStruct, v826, v829);
            v199 = (int *)v198;
            if ( v198 == 0 )
              goto LABEL_1261;
            if ( (unsigned int)sub_18000CC50(v198, PyModule_Type) != 0 )
              v200 = sub_180371D80(a1, (_DWORD)v199, qword_1803CA5B8, qword_1803BE050, qword_1803AB250);
            else
              v200 = sub_180371890(a1, v199, qword_1803BE050);
            v67 = *v199 < 0;
            v201 = (int *)v200;
            v889 = (int *)v200;
            if ( !v67 )
            {
              v142 = (*(_QWORD *)v199)-- == 1;
              if ( v142 )
              {
                v202 = sub_18000CC40(v199);
                (*(void (__fastcall **)(int *))(v202 + 48))(v199);
              }
            }
            if ( v201 == nullptr )
              goto LABEL_1261;
            v203 = v884;
            v16 = 116;
            v204 = v884;
            v913 = v201;
            *(_DWORD *)(v13 + 40) = 116;
            v205 = sub_180376770(a1, v201, v204);
            v938 = (int *)v205;
            v206 = v205;
            if ( v205 == 0 )
              goto LABEL_1261;
            v67 = *v195 < 0;
            v884 = (int *)v205;
            if ( !v67 )
            {
              v142 = (*(_QWORD *)v195)-- == 1;
              if ( v142 )
              {
                v207 = sub_18000CC40(v203);
                (*(void (__fastcall **)(int *))(v207 + 48))(v203);
              }
            }
            v208 = qword_1803BE058;
            v209 = qword_1803BE060;
            v210 = PyDict_NewPresized(2);
            PyDict_SetItem(v210, v208, v209);
            PyDict_SetItem(v210, qword_1803BE068, v206);
            v882 = (_QWORD *)sub_180372160(a1, 1);
            sub_1800B46F0(v882, 0, v210);
            v10 = v867;
            v211 = sub_18037F0C0(a1, v867[2], qword_1803BE070);
            v212 = (int *)v211;
            v213 = a1;
            if ( v211 == 0 )
              goto LABEL_355;
            v214 = sub_180001B80(a1, v211, qword_1803BE078);
            if ( *v212 >= 0 )
            {
              v142 = (*(_QWORD *)v212)-- == 1;
              if ( v142 )
              {
                v215 = sub_18000CC40(v212);
                (*(void (__fastcall **)(int *))(v215 + 48))(v212);
              }
            }
            if ( v214 == 0 )
            {
              v213 = a1;
LABEL_355:
              sub_18000CE30(v213, v863);
              v16 = 118;
LABEL_356:
              v11 = v865;
              v17 = v866;
              goto LABEL_513;
            }
            v216 = v214;
            v895 = (_QWORD *)v214;
            v217 = v914 != (int *)Py_NoneStruct;
            if ( v914 != (int *)Py_NoneStruct )
            {
              v217 = sub_180384830(v914, qword_1803BE080);
              if ( v217 == -1 )
              {
                sub_18000CE30(a1, v863);
                v11 = v865;
                v16 = 119;
                v17 = v866;
                goto LABEL_513;
              }
            }
            if ( v217 == 1 )
            {
              v890 = (int *)sub_180372160(a1, 0);
              v866 = (int *)qword_1803AB250;
              if ( *(_DWORD *)qword_1803AB250 != -1 )
                ++*(_DWORD *)qword_1803AB250;
              v16 = 122;
              *(_DWORD *)(v13 + 40) = 122;
              v218 = sub_180376770(a1, PyReversed_Type, v214);
              v219 = (int *)v218;
              if ( v218 == 0 )
                goto LABEL_369;
              v220 = sub_180001930(a1, v218);
              v67 = *v219 < 0;
              v221 = (_QWORD *)v220;
              v868 = (_QWORD *)v220;
              if ( !v67 )
              {
                v142 = (*(_QWORD *)v219)-- == 1;
                if ( v142 )
                {
                  v222 = sub_18000CC40(v219);
                  (*(void (__fastcall **)(int *))(v222 + 48))(v219);
                  v221 = v868;
                }
              }
              if ( v221 == nullptr )
              {
LABEL_369:
                sub_18000CE30(a1, v863);
                v11 = v865;
                v17 = v866;
                goto LABEL_513;
              }
              v223 = (int *)sub_18000D430(v221);
              if ( v223 != nullptr )
              {
                while ( 1 )
                {
                  v224 = v900;
                  v225 = v865;
                  v226 = v900;
                  v227 = (int *)v891;
                  v228 = v865;
                  v900 = v223;
                  if ( v226 != nullptr && *v224 >= 0 )
                  {
                    v142 = (*(_QWORD *)v224)-- == 1;
                    if ( v142 )
                    {
                      v229 = sub_18000CC40(v224);
                      (*(void (__fastcall **)(int *))(v229 + 48))(v226);
                      v225 = v228;
                    }
                  }
                  v230 = *v223;
                  v231 = v225;
                  v865 = v223;
                  v232 = v230 + 1;
                  if ( v232 != 0 )
                    *v223 = v232;
                  if ( v225 != nullptr && *v228 >= 0 )
                  {
                    v142 = (*(_QWORD *)v228)-- == 1;
                    if ( v142 )
                    {
                      v233 = sub_18000CC40(v225);
                      (*(void (__fastcall **)(int *))(v233 + 48))(v231);
                    }
                  }
                  v234 = (int *)sub_18037F0C0(a1, v223, qword_1803BDFD8);
                  if ( v234 == nullptr )
                    break;
                  *(_DWORD *)(v864 + 40) = 123;
                  v235 = (int *)sub_180378210(a1, v234, qword_1803BE088);
                  if ( *v234 >= 0 )
                  {
                    v142 = (*(_QWORD *)v234)-- == 1;
                    if ( v142 )
                    {
                      v236 = sub_18000CC40(v234);
                      (*(void (__fastcall **)(int *))(v236 + 48))(v234);
                    }
                  }
                  if ( v235 == nullptr )
                    break;
                  v237 = sub_1803749A0(v235);
                  if ( *v235 >= 0 )
                  {
                    v142 = (*(_QWORD *)v235)-- == 1;
                    if ( v142 )
                    {
                      v238 = sub_18000CC40(v235);
                      (*(void (__fastcall **)(int *))(v238 + 48))(v235);
                    }
                  }
                  if ( v237 == 0 )
                    break;
                  v891 = (_QWORD *)v237;
                  if ( v227 != nullptr && *v227 >= 0 )
                  {
                    v142 = (*(_QWORD *)v227)-- == 1;
                    if ( v142 )
                    {
                      v239 = sub_18000CC40(v227);
                      (*(void (__fastcall **)(int *))(v239 + 48))(v227);
                    }
                  }
                  if ( v866 == nullptr )
                  {
                    sub_180370FE0(a1, v863, qword_1803BE090);
                    sub_18000D1D0(a1, v863[0]);
                    v16 = 124;
                    goto LABEL_411;
                  }
                  v931 = v866;
                  v240 = sub_180375370(a1, v237);
                  v241 = (int *)v240;
                  if ( v240 == 0 )
                    goto LABEL_421;
                  v242 = sub_180382B70(&v931, v240);
                  if ( *v241 >= 0 )
                  {
                    v142 = (*(_QWORD *)v241)-- == 1;
                    if ( v142 )
                    {
                      v243 = sub_18000CC40(v241);
                      (*(void (__fastcall **)(int *))(v243 + 48))(v241);
                    }
                  }
                  if ( v242 == 0 )
                  {
LABEL_421:
                    sub_18000CE30(a1, v863);
                    v16 = 124;
                    goto LABEL_411;
                  }
                  v866 = v931;
                  v244 = sub_180385620(v931, v914);
                  if ( v244 == -1 )
                  {
                    sub_18000CE30(a1, v863);
                    v16 = 125;
                    goto LABEL_411;
                  }
                  if ( v244 == 1 )
                    goto LABEL_424;
                  if ( v890 == nullptr )
                  {
                    sub_180370FE0(a1, v863, qword_1803BE098);
                    sub_18000D1D0(a1, v863[0]);
                    v16 = 127;
                    goto LABEL_411;
                  }
                  sub_1803725B0(v890, v223);
                  if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                    goto LABEL_410;
                  v223 = (int *)sub_18000D430(v868);
                  if ( v223 == nullptr )
                    goto LABEL_409;
                }
                sub_18000CE30(a1, v863);
                v16 = 123;
                goto LABEL_411;
              }
LABEL_409:
              if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
              {
LABEL_410:
                sub_18000CE30(a1, v863);
LABEL_411:
                v62 = v863[0];
                sub_18000CE70(v863);
                v245 = v900;
                if ( v900 != nullptr && *v900 >= 0 )
                {
                  v142 = (*(_QWORD *)v900)-- == 1;
                  if ( v142 )
                  {
                    v246 = sub_18000CC40(v245);
                    (*(void (__fastcall **)(int *))(v246 + 48))(v245);
                  }
                }
                v247 = v868;
                if ( *(int *)v868 >= 0 )
                {
                  v142 = (*v868)-- == 1;
                  if ( v142 )
                  {
                    v248 = sub_18000CC40(v247);
                    (*(void (__fastcall **)(_QWORD *))(v248 + 48))(v247);
                  }
                }
                v11 = v865;
                goto LABEL_510;
              }
LABEL_424:
              v249 = v900;
              if ( v900 != nullptr && *v900 >= 0 )
              {
                v142 = (*(_QWORD *)v900)-- == 1;
                if ( v142 )
                {
                  v250 = sub_18000CC40(v249);
                  (*(void (__fastcall **)(int *))(v250 + 48))(v249);
                }
              }
              v251 = v868;
              if ( *(int *)v868 >= 0 )
              {
                v142 = (*v868)-- == 1;
                if ( v142 )
                {
                  v252 = sub_18000CC40(v251);
                  (*(void (__fastcall **)(_QWORD *))(v252 + 48))(v251);
                }
              }
              v253 = v890;
              if ( v890 == nullptr )
              {
                sub_180370FE0(a1, v863, qword_1803BE098);
                sub_18000D1D0(a1, v863[0]);
                v11 = v865;
                v16 = 128;
                v17 = v866;
                goto LABEL_512;
              }
              v254 = v864;
              v16 = 128;
              *(_DWORD *)(v864 + 40) = 128;
              v255 = sub_180376770(a1, PyReversed_Type, v253);
              v256 = (int *)v255;
              if ( v255 == 0 )
              {
                sub_18000CE30(a1, v863);
                v11 = v865;
                v10 = v867;
                v17 = v866;
                goto LABEL_514;
              }
              v257 = sub_180372690(a1, v255);
              if ( *v256 >= 0 )
              {
                v142 = (*(_QWORD *)v256)-- == 1;
                if ( v142 )
                {
                  v258 = sub_18000CC40(v256);
                  (*(void (__fastcall **)(int *))(v258 + 48))(v256);
                }
              }
              if ( v257 == 0 )
                goto LABEL_1261;
              v259 = (int *)v895;
              v216 = v257;
              v895 = (_QWORD *)v257;
              if ( *v259 >= 0 )
              {
                v142 = (*(_QWORD *)v259)-- == 1;
                if ( v142 )
                {
                  v260 = sub_18000CC40(v259);
                  (*(void (__fastcall **)(int *))(v260 + 48))(v259);
                }
              }
              v10 = v867;
            }
            v17 = v866;
            v930 = v891;
            v928 = v890;
            v929 = v866;
            v927 = (_QWORD *)v216;
            v261 = sub_180001930(a1, v216);
            v869 = (int *)v261;
            v262 = (int *)v261;
            if ( v261 == 0 )
            {
              sub_18000CE30(a1, v863);
              v11 = v865;
              v16 = 129;
              goto LABEL_513;
            }
            v263 = (int *)sub_18000D430(v261);
            if ( v263 != nullptr )
            {
              v16 = 132;
              while ( 1 )
              {
                v264 = v896;
                v265 = v865;
                v266 = v896;
                v267 = (int *)v892;
                v268 = v865;
                v896 = v263;
                if ( v266 != nullptr && *v264 >= 0 )
                {
                  v142 = (*(_QWORD *)v264)-- == 1;
                  if ( v142 )
                  {
                    v269 = sub_18000CC40(v264);
                    (*(void (__fastcall **)(int *))(v269 + 48))(v266);
                    v265 = v268;
                  }
                }
                v270 = *v263;
                v271 = v265;
                v865 = v263;
                v11 = v263;
                v272 = v270 + 1;
                if ( v272 != 0 )
                  *v263 = v272;
                if ( v265 != nullptr && *v268 >= 0 )
                {
                  v142 = (*(_QWORD *)v268)-- == 1;
                  if ( v142 )
                  {
                    v273 = sub_18000CC40(v265);
                    (*(void (__fastcall **)(int *))(v273 + 48))(v271);
                  }
                }
                v274 = (int *)sub_18037F0C0(a1, v263, qword_1803BDFD8);
                if ( v274 == nullptr )
                  break;
                *(_DWORD *)(v864 + 40) = 130;
                v275 = (int *)sub_180378210(a1, v274, qword_1803BE0A0);
                if ( *v274 >= 0 )
                {
                  v142 = (*(_QWORD *)v274)-- == 1;
                  if ( v142 )
                  {
                    v276 = sub_18000CC40(v274);
                    (*(void (__fastcall **)(int *))(v276 + 48))(v274);
                  }
                }
                if ( v275 == nullptr )
                  break;
                v277 = sub_1803749A0(v275);
                if ( *v275 >= 0 )
                {
                  v142 = (*(_QWORD *)v275)-- == 1;
                  if ( v142 )
                  {
                    v278 = sub_18000CC40(v275);
                    (*(void (__fastcall **)(int *))(v278 + 48))(v275);
                  }
                }
                if ( v277 == 0 )
                {
                  sub_18000CE30(a1, v863);
                  v16 = 130;
                  goto LABEL_502;
                }
                v892 = (_QWORD *)v277;
                if ( v267 != nullptr && *v267 >= 0 )
                {
                  v142 = (*(_QWORD *)v267)-- == 1;
                  if ( v142 )
                  {
                    v279 = sub_18000CC40(v267);
                    (*(void (__fastcall **)(int *))(v279 + 48))(v267);
                  }
                }
                v280 = PySequence_Contains(qword_1803BE0A8, v277);
                if ( v280 == -1 )
                {
                  sub_18000CE30(a1, v863);
                  v16 = 131;
                  goto LABEL_502;
                }
                if ( v280 == 1 )
                {
                  if ( v882 == nullptr )
                  {
                    sub_180370FE0(a1, v863, qword_1803BE070);
                    sub_18000D1D0(a1, v863[0]);
LABEL_502:
                    v11 = v263;
                    goto LABEL_503;
                  }
                  v281 = qword_1803BE058;
                  v282 = (int *)PyDict_NewPresized(2);
                  PyDict_SetItem(v282, v281, v277);
                  v283 = qword_1803BE068;
                  v284 = (int *)sub_18037F0C0(a1, v263, qword_1803BDFD8);
                  if ( v284 == nullptr )
                    goto LABEL_499;
                  *(_DWORD *)(v864 + 40) = 132;
                  v285 = (int *)sub_180378210(a1, v284, qword_1803BE088);
                  if ( *v284 >= 0 )
                  {
                    v142 = (*(_QWORD *)v284)-- == 1;
                    if ( v142 )
                    {
                      v286 = sub_18000CC40(v284);
                      (*(void (__fastcall **)(int *))(v286 + 48))(v284);
                    }
                  }
                  if ( v285 == nullptr )
                    goto LABEL_499;
                  v287 = (int *)sub_1803749A0(v285);
                  if ( *v285 >= 0 )
                  {
                    v142 = (*(_QWORD *)v285)-- == 1;
                    if ( v142 )
                    {
                      v288 = sub_18000CC40(v285);
                      (*(void (__fastcall **)(int *))(v288 + 48))(v285);
                    }
                  }
                  if ( v287 == nullptr )
                  {
LABEL_499:
                    sub_18000CE30(a1, v863);
                    if ( *v282 >= 0 )
                    {
                      v142 = (*(_QWORD *)v282)-- == 1;
                      if ( v142 )
                      {
                        v294 = sub_18000CC40(v282);
                        (*(void (__fastcall **)(int *))(v294 + 48))(v282);
                      }
                    }
                    goto LABEL_502;
                  }
                  PyDict_SetItem(v282, v283, v287);
                  if ( *v287 >= 0 )
                  {
                    v142 = (*(_QWORD *)v287)-- == 1;
                    if ( v142 )
                    {
                      v289 = sub_18000CC40(v287);
                      (*(void (__fastcall **)(int *))(v289 + 48))(v287);
                    }
                  }
                  sub_180372560(v882, v282);
                }
                if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                {
                  sub_18000CE30(a1, v863);
                  v16 = 129;
                  goto LABEL_502;
                }
                v262 = v869;
                v263 = (int *)sub_18000D430(v869);
                if ( v263 == nullptr )
                  goto LABEL_489;
              }
              sub_18000CE30(a1, v863);
              v16 = 130;
              goto LABEL_503;
            }
LABEL_489:
            v11 = v865;
            v919 = v892;
            v900 = v865;
            if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
            {
              sub_18000CE30(a1, v863);
              v16 = 129;
LABEL_503:
              v62 = v863[0];
              sub_18000CE70(v863);
              v295 = v896;
              if ( v896 != nullptr && *v896 >= 0 )
              {
                v142 = (*(_QWORD *)v896)-- == 1;
                if ( v142 )
                {
                  v296 = sub_18000CC40(v295);
                  (*(void (__fastcall **)(int *))(v296 + 48))(v295);
                }
              }
              v297 = v869;
              if ( *v869 >= 0 )
              {
                v142 = (*(_QWORD *)v869)-- == 1;
                if ( v142 )
                {
                  v298 = sub_18000CC40(v297);
                  (*(void (__fastcall **)(int *))(v298 + 48))(v297);
                }
              }
              goto LABEL_510;
            }
            v290 = v896;
            if ( v896 != nullptr && *v896 >= 0 )
            {
              v142 = (*(_QWORD *)v896)-- == 1;
              if ( v142 )
              {
                v291 = sub_18000CC40(v290);
                (*(void (__fastcall **)(int *))(v291 + 48))(v290);
              }
            }
            if ( *v262 >= 0 )
            {
              v142 = (*(_QWORD *)v262)-- == 1;
              if ( v142 )
              {
                v292 = sub_18000CC40(v262);
                (*(void (__fastcall **)(int *))(v292 + 48))(v262);
              }
            }
            v293 = v882;
            if ( v882 == nullptr )
            {
              sub_180370FE0(a1, v863, qword_1803BE070);
              sub_18000D1D0(a1, v863[0]);
              v17 = v866;
              v16 = 133;
              goto LABEL_512;
            }
            v301 = qword_1803BE058;
            v302 = qword_1803BE0B0;
            v303 = PyDict_NewPresized(2);
            PyDict_SetItem(v303, v301, v302);
            v304 = v910;
            PyDict_SetItem(v303, qword_1803BE068, v910);
            sub_180372560(v293, v303);
            v10 = v867;
            v305 = sub_18037F0C0(a1, v867[2], qword_1803BE070);
            v306 = (int *)v305;
            if ( v305 == 0 )
            {
              sub_18000CE30(a1, v863);
              v17 = v866;
              v16 = 135;
              goto LABEL_513;
            }
            v307 = (int *)sub_18037F0C0(a1, v305, qword_1803BE0B8);
            if ( *v306 >= 0 )
            {
              v142 = (*(_QWORD *)v306)-- == 1;
              if ( v142 )
              {
                v308 = sub_18000CC40(v306);
                (*(void (__fastcall **)(int *))(v308 + 48))(v306);
              }
            }
            if ( v307 == nullptr )
            {
              sub_18000CE30(a1, v863);
              v11 = v865;
              v16 = 135;
              v17 = v866;
              goto LABEL_513;
            }
            v309 = qword_1803BE058;
            v310 = qword_1803BE0B0;
            v37 = (int *)PyDict_NewPresized(3);
            PyDict_SetItem(v37, v309, v310);
            PyDict_SetItem(v37, qword_1803BE068, v304);
            v311 = qword_1803BE0C0;
            v312 = sub_1801F4170(a1);
            if ( v312 == 0 )
            {
              sub_180373340(a1, v863, qword_1803BE0C8);
              v16 = 135;
LABEL_533:
              if ( *v307 >= 0 )
              {
                v142 = (*(_QWORD *)v307)-- == 1;
                if ( v142 )
                {
                  v316 = sub_18000CC40(v307);
                  (*(void (__fastcall **)(int *))(v316 + 48))(v307);
                }
              }
              if ( *v37 >= 0 )
              {
                v142 = (*(_QWORD *)v37)-- == 1;
                if ( v142 )
                  goto LABEL_538;
              }
              goto LABEL_1271;
            }
            v313 = v864;
            v16 = 135;
            *(_DWORD *)(v864 + 40) = 135;
            v314 = sub_180376210(a1, v312);
            v315 = (int *)v314;
            if ( v314 == 0 )
            {
              sub_18000CE30(a1, v863);
              goto LABEL_533;
            }
            PyDict_SetItem(v37, v311, v314);
            if ( *v315 >= 0 )
            {
              v142 = (*(_QWORD *)v315)-- == 1;
              if ( v142 )
              {
                v318 = sub_18000CC40(v315);
                (*(void (__fastcall **)(int *))(v318 + 48))(v315);
              }
            }
            *(_DWORD *)(v313 + 40) = 135;
            v319 = (int *)sub_180376770(a1, v307, v37);
            if ( *v307 >= 0 )
            {
              v142 = (*(_QWORD *)v307)-- == 1;
              if ( v142 )
              {
                v320 = sub_18000CC40(v307);
                (*(void (__fastcall **)(int *))(v320 + 48))(v307);
              }
            }
            if ( *v37 >= 0 )
            {
              v142 = (*(_QWORD *)v37)-- == 1;
              if ( v142 )
              {
                v321 = sub_18000CC40(v37);
                (*(void (__fastcall **)(int *))(v321 + 48))(v37);
              }
            }
            if ( v319 == nullptr )
              goto LABEL_549;
            if ( *v319 >= 0 )
            {
              v142 = (*(_QWORD *)v319)-- == 1;
              if ( v142 )
              {
                v323 = sub_18000CC40(v319);
                (*(void (__fastcall **)(int *))(v323 + 48))(v319);
              }
            }
            v10 = v867;
            if ( (unsigned __int8)sub_18037F3A0(a1, v867[2], qword_1803BE0D0, Py_FalseStruct) == 0 )
            {
              sub_18000CE30(a1, v863);
              v11 = v865;
              v254 = v313;
              v17 = v866;
              v16 = 136;
              goto LABEL_514;
            }
            v324 = v871;
            v325 = sub_18037F0C0(a1, *((_QWORD *)v871 + 2), qword_1803BDF10);
            v326 = (int *)v325;
            if ( v325 == 0 )
            {
              sub_18000CE30(a1, v863);
              v11 = v865;
              v254 = v313;
              v17 = v866;
              v16 = 137;
              goto LABEL_514;
            }
            v327 = v10[2];
            v16 = 137;
            *(_DWORD *)(v313 + 40) = 137;
            v328 = (int *)sub_18037E6C0(a1, v325, qword_1803BE0D8, v327);
            if ( *v326 >= 0 )
            {
              v142 = (*(_QWORD *)v326)-- == 1;
              if ( v142 )
              {
                v329 = sub_18000CC40(v326);
                (*(void (__fastcall **)(int *))(v329 + 48))(v326);
              }
            }
            if ( v328 == nullptr )
              goto LABEL_562;
            if ( *v328 >= 0 )
            {
              v142 = (*(_QWORD *)v328)-- == 1;
              if ( v142 )
              {
                v331 = sub_18000CC40(v328);
                (*(void (__fastcall **)(int *))(v331 + 48))(v328);
              }
            }
            v332 = (_DWORD *)qword_1803BE0E0;
            v333 = sub_180374A40(a1, 4);
            v334 = (int *)v333;
            if ( *v332 != -1 )
              ++*v332;
            sub_18000CC80(v333, 0, v332);
            v335 = sub_180375520(a1, v893[2], qword_1803BDFE8);
            if ( v335 != 0 )
            {
              sub_18000CC80(v334, 1, v335);
              v336 = qword_1803BE0E8;
              if ( *(_DWORD *)qword_1803BE0E8 != -1 )
                ++*(_DWORD *)qword_1803BE0E8;
              sub_18000CC80(v334, 2, v336);
              v337 = sub_18037F0C0(a1, v10[2], qword_1803BE070);
              v338 = (int *)v337;
              if ( v337 != 0 )
              {
                v339 = (int *)sub_180375370(a1, v337);
                if ( *v338 >= 0 )
                {
                  v142 = (*(_QWORD *)v338)-- == 1;
                  if ( v142 )
                  {
                    v340 = sub_18000CC40(v338);
                    (*(void (__fastcall **)(int *))(v340 + 48))(v338);
                  }
                }
                if ( v339 != nullptr )
                {
                  v341 = sub_180375520(a1, v339, qword_1803BDFE8);
                  if ( *v339 >= 0 )
                  {
                    v142 = (*(_QWORD *)v339)-- == 1;
                    if ( v142 )
                    {
                      v342 = sub_18000CC40(v339);
                      (*(void (__fastcall **)(int *))(v342 + 48))(v339);
                    }
                  }
                  if ( v341 != 0 )
                  {
                    sub_18000CC80(v334, 3, v341);
                    v344 = PyUnicode_Join(qword_1803BDFE8, v334);
                    if ( *v334 >= 0 )
                    {
                      v142 = (*(_QWORD *)v334)-- == 1;
                      if ( v142 )
                      {
                        v345 = sub_18000CC40(v334);
                        (*(void (__fastcall **)(int *))(v345 + 48))(v334);
                      }
                    }
                    if ( v344 == 0 )
                    {
                      sub_18000CE30(a1, v863);
                      v11 = v865;
                      v254 = v313;
                      v17 = v866;
                      v16 = 139;
                      goto LABEL_514;
                    }
                    v346 = v344;
                    v347 = v902;
                    sub_1800DEB50(v902, v346);
                    v348 = sub_1801F4060(a1);
                    if ( v348 == 0 )
                    {
                      sub_180373340(a1, v863, qword_1803BE0F0);
                      v11 = v865;
                      v254 = v313;
                      v17 = v866;
                      v16 = 140;
                      goto LABEL_514;
                    }
                    v16 = 140;
                    *(_DWORD *)(v313 + 40) = 140;
                    v349 = sub_18037E560(a1, v348, qword_1803BE0F0);
                    v896 = (int *)v349;
                    v330 = a1;
                    if ( v349 == 0 )
                      goto LABEL_563;
                    v350 = *((_QWORD *)v324 + 2);
                    v912 = (int *)v349;
                    v351 = (int *)sub_18037F0C0(a1, v350, *(_QWORD *)&qword_1803BE0F8);
                    if ( v351 == nullptr )
                    {
                      sub_18000CE30(a1, v863);
                      v11 = v865;
                      v254 = v313;
                      v17 = v866;
                      v16 = 141;
                      goto LABEL_514;
                    }
                    v352 = qword_1803BE100;
                    v16 = 141;
                    v353 = v893[2];
                    v354 = *((_QWORD *)v347 + 2);
                    *(_DWORD *)(v313 + 40) = 141;
                    v951[1] = v354;
                    v951[0] = v353;
                    v951[2] = v352;
                    v355 = (int *)sub_18037D280(a1, v351, qword_1803BE108, v951, qword_1803BE110);
                    if ( *v351 >= 0 )
                    {
                      v142 = (*(_QWORD *)v351)-- == 1;
                      if ( v142 )
                      {
                        v356 = sub_18000CC40(v351);
                        (*(void (__fastcall **)(int *))(v356 + 48))(v351);
                      }
                    }
                    if ( v355 == nullptr )
                      goto LABEL_562;
                    if ( *v355 >= 0 )
                    {
                      v142 = (*(_QWORD *)v355)-- == 1;
                      if ( v142 )
                      {
                        v357 = sub_18000CC40(v355);
                        (*(void (__fastcall **)(int *))(v357 + 48))(v355);
                      }
                    }
                    v870 = (_QWORD *)Py_NoneStruct;
                    if ( Py_NoneStruct != -1 )
                      ++Py_NoneStruct;
                    v358 = (int *)qword_1803BDFE8;
                    v875 = (int *)Py_NoneStruct;
                    if ( *(_DWORD *)qword_1803BDFE8 != -1 )
                      ++*(_DWORD *)qword_1803BDFE8;
                    v872 = v358;
                    v359 = sub_180372160(a1, 0);
                    v360 = v915;
                    sub_1800DEB50(v915, v359);
                    v361 = sub_18036FB80(a1, qword_1803BE118);
                    v957[0] = v360;
                    if ( *v360 != -1 )
                      ++*v360;
                    v957[1] = v902;
                    if ( *v902 != -1 )
                      ++*v902;
                    v362 = v871;
                    v957[2] = v871;
                    if ( *v871 != -1 )
                      ++*v871;
                    v957[3] = v893;
                    if ( *(_DWORD *)v893 != -1 )
                      ++*(_DWORD *)v893;
                    v363 = sub_180217110(a1, v361, v957);
                    v364 = *((_QWORD *)v362 + 2);
                    v16 = 159;
                    v365 = v10[2];
                    *(_DWORD *)(v313 + 40) = 159;
                    v903 = (int *)v363;
                    v869 = (int *)v363;
                    v366 = sub_18037E6C0(a1, v364, *(_QWORD *)&qword_1803BE130, v365);
                    v334 = (int *)v366;
                    if ( v366 == 0 )
                    {
LABEL_562:
                      v330 = a1;
LABEL_563:
                      sub_18000CE30(v330, v863);
LABEL_564:
                      v11 = v865;
                      v254 = v313;
                      v17 = v866;
                      goto LABEL_514;
                    }
                    v367 = sub_180001510(v366);
                    if ( v367 == -1 )
                    {
                      sub_18000CE30(a1, v863);
                      if ( *v334 < 0 )
                        goto LABEL_564;
                      v142 = (*(_QWORD *)v334)-- == 1;
                      if ( !v142 )
                        goto LABEL_564;
LABEL_584:
                      v343 = sub_18000CC40(v334);
                      (*(void (__fastcall **)(int *))(v343 + 48))(v334);
                      v11 = v865;
                      v254 = v313;
                      v17 = v866;
                      goto LABEL_514;
                    }
                    if ( *v334 >= 0 )
                    {
                      v142 = (*(_QWORD *)v334)-- == 1;
                      if ( v142 )
                      {
                        v368 = sub_18000CC40(v334);
                        (*(void (__fastcall **)(int *))(v368 + 48))(v334);
                      }
                    }
                    if ( v367 != 0 )
                    {
                      v369 = sub_1801F3F00(a1);
                      if ( v369 == 0 )
                      {
                        sub_180373340(a1, v863, qword_1803BE138);
LABEL_625:
                        v10 = v867;
                        v254 = v313;
                        v17 = v866;
                        v16 = 158;
                        v11 = v865;
                        goto LABEL_514;
                      }
                      v370 = v867;
                      v371 = (int *)sub_18037F0C0(a1, v867[2], qword_1803BDFA0);
                      if ( v371 == nullptr )
                      {
                        sub_18000CE30(a1, v863);
                        goto LABEL_625;
                      }
                      v372 = sub_18037F0C0(a1, v370[2], qword_1803BDFC0);
                      v373 = (int *)v372;
                      if ( v372 == 0 )
                      {
                        v374 = a1;
LABEL_630:
                        sub_18000CE30(v374, v863);
                        if ( *v371 >= 0 )
                        {
                          v142 = (*(_QWORD *)v371)-- == 1;
                          if ( v142 )
                          {
                            v375 = sub_18000CC40(v371);
                            (*(void (__fastcall **)(int *))(v375 + 48))(v371);
                          }
                        }
                        goto LABEL_625;
                      }
                      v376 = sub_180001510(v372);
                      if ( *v373 >= 0 )
                      {
                        v142 = (*(_QWORD *)v373)-- == 1;
                        if ( v142 )
                        {
                          v377 = sub_18000CC40(v373);
                          (*(void (__fastcall **)(int *))(v377 + 48))(v373);
                        }
                      }
                      v374 = a1;
                      if ( v376 == -1 )
                        goto LABEL_630;
                      v378 = Py_FalseStruct;
                      v142 = v376 == 0;
                      v16 = 158;
                      if ( !v142 )
                        v378 = Py_TrueStruct;
                      *(_DWORD *)(v313 + 40) = 158;
                      v939 = v378;
                      v940 = v371;
                      v379 = sub_18037D030(a1, v369, &v940, &v939, qword_1803BE140);
                      if ( *v371 >= 0 )
                      {
                        v142 = (*(_QWORD *)v371)-- == 1;
                        if ( v142 )
                        {
                          v380 = sub_18000CC40(v371);
                          (*(void (__fastcall **)(int *))(v380 + 48))(v371);
                        }
                      }
                      if ( v379 == 0 )
                      {
                        sub_18000CE30(a1, v863);
                        v10 = v867;
                        v254 = v313;
                        v17 = v866;
                        v11 = v865;
                        goto LABEL_514;
                      }
                      v898 = (int *)sub_180372160(a1, 1);
                      v381 = (__int64)v898;
                      sub_1800B46F0(v898, 0, v379);
                    }
                    else
                    {
                      v381 = sub_180372160(a1, 0);
                      v898 = (int *)v381;
                    }
                    v888 = (int *)v381;
                    if ( (unsigned int)sub_180001510(v381) != 1 )
                    {
                      v399 = 191;
                      v518 = qword_1803BE1F8;
                      *(_DWORD *)(v313 + 40) = 191;
                      v863[0] = sub_180376770(a1, PyExc_RuntimeError, v518);
                      sub_180373100(a1, v863);
                      v400 = v313;
                      goto LABEL_909;
                    }
                    v382 = v871;
                    v383 = sub_18037F0C0(a1, *((_QWORD *)v871 + 2), qword_1803BDF38);
                    v384 = (int *)v383;
                    v385 = a1;
                    if ( v383 == 0 )
                    {
LABEL_670:
                      sub_18000CE30(v385, v863);
                      v398 = v863[0];
                      v399 = 165;
                      sub_18000CE70(v863);
                      v863[0] = v398;
                      v400 = v864;
                      goto LABEL_909;
                    }
                    v386 = sub_18037F0C0(a1, v383, qword_1803BE148);
                    if ( *v384 >= 0 )
                    {
                      v142 = (*(_QWORD *)v384)-- == 1;
                      if ( v142 )
                      {
                        v387 = sub_18000CC40(v384);
                        (*(void (__fastcall **)(int *))(v387 + 48))(v384);
                      }
                    }
                    if ( v386 == 0 )
                    {
                      v385 = a1;
                      goto LABEL_670;
                    }
                    v388 = qword_1803BE070;
                    v389 = PyDict_NewPresized(5);
                    PyDict_SetItem(v389, v388, v882);
                    PyDict_SetItem(v389, qword_1803BE150, v898);
                    PyDict_SetItem(v389, qword_1803BE158, v908);
                    PyDict_SetItem(v389, qword_1803BE160, v916);
                    v390 = qword_1803BE168;
                    v952[0] = v382;
                    if ( *v382 != -1 )
                      ++*v382;
                    v952[1] = v867;
                    if ( *(_DWORD *)v867 != -1 )
                      ++*(_DWORD *)v867;
                    v952[2] = v893;
                    if ( *(_DWORD *)v893 != -1 )
                      ++*(_DWORD *)v893;
                    v391 = (int *)sub_180217170(a1, v952);
                    PyDict_SetItem(v389, v390, v391);
                    if ( *v391 >= 0 )
                    {
                      v142 = (*(_QWORD *)v391)-- == 1;
                      if ( v142 )
                      {
                        v392 = sub_18000CC40(v391);
                        (*(void (__fastcall **)(int *))(v392 + 48))(v391);
                      }
                    }
                    v393 = v909;
                    if ( *v909 != -1 )
                      ++*v909;
                    v953[2] = v393;
                    v953[0] = v386;
                    v953[1] = v389;
                    v394 = sub_180097770(a1, v953);
                    v395 = (int *)v394;
                    if ( v394 == 0 )
                    {
                      v385 = a1;
                      goto LABEL_670;
                    }
                    v396 = (int *)sub_180001A60(v394);
                    if ( *v395 >= 0 )
                    {
                      v142 = (*(_QWORD *)v395)-- == 1;
                      if ( v142 )
                      {
                        v397 = sub_18000CC40(v395);
                        (*(void (__fastcall **)(int *))(v397 + 48))(v395);
                      }
                    }
                    v385 = a1;
                    if ( v396 == nullptr )
                      goto LABEL_670;
                    v401 = sub_18002A560(a1, v863, v396, 0, 2);
                    v402 = (int *)v401;
                    if ( v401 == 0 )
                    {
                      v403 = v874;
LABEL_675:
                      v399 = 165;
                      v407 = v863[0];
                      sub_18000CE70(v863);
                      if ( *v396 >= 0 )
                      {
                        v142 = (*(_QWORD *)v396)-- == 1;
                        if ( v142 )
                        {
                          v408 = sub_18000CC40(v396);
                          (*(void (__fastcall **)(int *))(v408 + 48))(v396);
                        }
                      }
                      v863[0] = v407;
                      sub_18000CE70(v863);
                      v409 = v880;
                      if ( v403 != nullptr && *v403 >= 0 )
                      {
                        v142 = (*(_QWORD *)v403)-- == 1;
                        if ( v142 )
                        {
                          v410 = sub_18000CC40(v403);
                          (*(void (__fastcall **)(int *))(v410 + 48))(v403);
                        }
                      }
                      if ( v409 != nullptr && *v409 >= 0 )
                      {
                        v142 = (*(_QWORD *)v409)-- == 1;
                        if ( v142 )
                        {
                          v411 = sub_18000CC40(v409);
                          (*(void (__fastcall **)(int *))(v411 + 48))(v409);
                        }
                      }
LABEL_686:
                      v863[0] = v407;
                      v400 = v864;
                      goto LABEL_909;
                    }
                    v874 = (int *)v401;
                    v403 = (int *)v401;
                    v404 = sub_18002A560(a1, v863, v396, 1, 2);
                    v405 = (int *)v404;
                    if ( v404 == 0 )
                      goto LABEL_675;
                    v880 = (int *)v404;
                    v406 = (_QWORD *)v404;
                    if ( (unsigned __int8)sub_180065DB0(a1, v863, v396, 2) == 0 )
                      goto LABEL_675;
                    if ( *v396 >= 0 )
                    {
                      v142 = (*(_QWORD *)v396)-- == 1;
                      if ( v142 )
                      {
                        v412 = sub_18000CC40(v396);
                        (*(void (__fastcall **)(int *))(v412 + 48))(v396);
                      }
                    }
                    v413 = *v402;
                    v414 = *v402 + 1;
                    if ( *v402 != -1 )
                    {
                      *v402 = v414;
                      v413 = v414;
                    }
                    v885 = v402;
                    if ( v413 >= 0 )
                    {
                      v142 = (*(_QWORD *)v402)-- == 1;
                      if ( v142 )
                      {
                        v415 = sub_18000CC40(v402);
                        (*(void (__fastcall **)(int *))(v415 + 48))(v402);
                      }
                    }
                    v416 = *v405;
                    v417 = *v405 + 1;
                    if ( *v405 != -1 )
                    {
                      *v405 = v417;
                      v416 = v417;
                    }
                    v906 = v406;
                    if ( v416 >= 0 )
                    {
                      v142 = (*(_QWORD *)v405)-- == 1;
                      if ( v142 )
                      {
                        v418 = sub_18000CC40(v406);
                        (*(void (__fastcall **)(_QWORD *))(v418 + 48))(v406);
                      }
                    }
                    v419 = sub_18037F0C0(a1, v402, qword_1803BE068);
                    v420 = (int *)v419;
                    if ( v419 == 0 )
                      goto LABEL_701;
                    v421 = sub_180001510(v419);
                    if ( v421 == -1 )
                    {
                      sub_18000CE30(a1, v863);
                      if ( *v420 >= 0 )
                      {
                        v142 = (*(_QWORD *)v420)-- == 1;
                        if ( v142 )
                        {
                          v422 = sub_18000CC40(v420);
                          (*(void (__fastcall **)(int *))(v422 + 48))(v420);
                        }
                      }
                      v400 = v864;
                      v399 = 173;
                      goto LABEL_909;
                    }
                    if ( v421 != 1 )
                    {
                      if ( *v420 >= 0 )
                      {
                        v142 = (*(_QWORD *)v420)-- == 1;
                        if ( v142 )
                        {
                          v423 = sub_18000CC40(v420);
                          (*(void (__fastcall **)(int *))(v423 + 48))(v420);
                        }
                      }
                      v420 = (int *)qword_1803BDFE8;
                      if ( *(_DWORD *)qword_1803BDFE8 != -1 )
                        ++*(_DWORD *)qword_1803BDFE8;
                    }
                    v424 = (int *)sub_18037F0C0(a1, v420, qword_1803BDFF0);
                    if ( *v420 >= 0 )
                    {
                      v142 = (*(_QWORD *)v420)-- == 1;
                      if ( v142 )
                      {
                        v425 = sub_18000CC40(v420);
                        (*(void (__fastcall **)(int *))(v425 + 48))(v420);
                      }
                    }
                    if ( v424 == nullptr )
                    {
LABEL_701:
                      sub_18000CE30(a1, v863);
                      v400 = v864;
                      v399 = 173;
                      goto LABEL_909;
                    }
                    v400 = v864;
                    v399 = 173;
                    *(_DWORD *)(v864 + 40) = 173;
                    v426 = (_DWORD *)sub_180376210(a1, v424);
                    if ( *v424 >= 0 )
                    {
                      v142 = (*(_QWORD *)v424)-- == 1;
                      if ( v142 )
                      {
                        v427 = sub_18000CC40(v424);
                        (*(void (__fastcall **)(int *))(v427 + 48))(v424);
                      }
                    }
                    if ( v426 == nullptr )
                    {
                      sub_18000CE30(a1, v863);
                      goto LABEL_909;
                    }
                    v67 = *v358 < 0;
                    v872 = v426;
                    if ( !v67 )
                    {
                      v142 = (*(_QWORD *)v358)-- == 1;
                      if ( v142 )
                      {
                        v428 = sub_18000CC40(v358);
                        (*(void (__fastcall **)(int *))(v428 + 48))(v358);
                      }
                    }
                    v429 = sub_18036F270(a1, v874, qword_1803BE180, qword_1803BDFE8);
                    v430 = (int *)v429;
                    if ( v429 != 0 )
                    {
                      v431 = sub_180001510(v429);
                      if ( v431 == -1 )
                      {
                        sub_18000CE30(a1, v863);
                        if ( *v430 >= 0 )
                        {
                          v142 = (*(_QWORD *)v430)-- == 1;
                          if ( v142 )
                          {
                            v432 = sub_18000CC40(v430);
                            (*(void (__fastcall **)(int *))(v432 + 48))(v430);
                            v400 = v864;
                            v399 = 174;
                            goto LABEL_909;
                          }
                        }
LABEL_727:
                        v400 = v864;
                        v399 = 174;
                        goto LABEL_909;
                      }
                      if ( v431 != 1 )
                      {
                        if ( *v430 >= 0 )
                        {
                          v142 = (*(_QWORD *)v430)-- == 1;
                          if ( v142 )
                          {
                            v433 = sub_18000CC40(v430);
                            (*(void (__fastcall **)(int *))(v433 + 48))(v430);
                          }
                        }
                        v430 = (int *)qword_1803BDFE8;
                        if ( *(_DWORD *)qword_1803BDFE8 != -1 )
                          ++*(_DWORD *)qword_1803BDFE8;
                      }
                      v434 = sub_1803749A0(v430);
                      if ( *v430 >= 0 )
                      {
                        v142 = (*(_QWORD *)v430)-- == 1;
                        if ( v142 )
                        {
                          v435 = sub_18000CC40(v430);
                          (*(void (__fastcall **)(int *))(v435 + 48))(v430);
                        }
                      }
                      if ( v434 != 0 )
                      {
                        v877 = (_QWORD *)v434;
                        v436 = sub_180001510(v406);
                        if ( v436 == -1 )
                        {
                          sub_18000CE30(a1, v863);
                          v400 = v864;
                          v399 = 175;
                          goto LABEL_909;
                        }
                        if ( v436 == 0 )
                        {
LABEL_1143:
                          v495 = v873;
                          goto LABEL_1144;
                        }
                        v437 = sub_180001930(a1, v406);
                        v868 = (_QWORD *)v437;
                        v438 = (int *)v437;
                        if ( v437 == 0 )
                        {
                          sub_18000CE30(a1, v863);
                          v400 = v864;
                          v399 = 176;
                          goto LABEL_909;
                        }
                        v439 = (_DWORD *)sub_18000D430(v437);
                        if ( v439 == nullptr )
                        {
LABEL_771:
                          if ( (unsigned __int8)sub_1800012A0(a1) != 0 )
                          {
LABEL_856:
                            v493 = v873;
                            if ( v873 != nullptr && *(int *)v873 >= 0 && (v142 = *v873 == 1, --*v873, v142) )
                            {
                              v494 = sub_18000CC40(v493);
                              v495 = v873;
                              (*(void (__fastcall **)(_QWORD *))(v494 + 48))(v873);
                            }
                            else
                            {
                              v495 = v493;
                            }
                            if ( *v438 >= 0 )
                            {
                              v142 = (*(_QWORD *)v438)-- == 1;
                              if ( v142 )
                              {
                                v517 = sub_18000CC40(v438);
                                (*(void (__fastcall **)(int *))(v517 + 48))(v438);
                              }
                            }
LABEL_1144:
                            v874 = (int *)v886;
                            v870 = v906;
                            v879 = v885;
                            v880 = (int *)v877;
                            v887 = v872;
                            v894 = v495;
                            v645 = sub_18000D570(qword_1803BE230);
                            v646 = qword_1803AB258;
                            v647 = v645;
                            v648 = sub_1801F4060(a1);
                            if ( v648 == 0 )
                            {
                              sub_180373340(a1, v863, qword_1803BE0F0);
                              v11 = v865;
                              v16 = 213;
                              v17 = v866;
                              goto LABEL_512;
                            }
                            v313 = v864;
                            v16 = 213;
                            *(_DWORD *)(v864 + 40) = 213;
                            v649 = sub_18037E560(a1, v648, qword_1803BE0F0);
                            v650 = (int *)v649;
                            if ( v649 == 0 )
                              goto LABEL_549;
                            v651 = (int *)sub_1803820D0(v649, v896);
                            if ( *v650 >= 0 )
                            {
                              v142 = (*(_QWORD *)v650)-- == 1;
                              if ( v142 )
                              {
                                v652 = sub_18000CC40(v650);
                                (*(void (__fastcall **)(int *))(v652 + 48))(v650);
                              }
                            }
                            if ( v651 == nullptr )
                            {
LABEL_549:
                              v322 = a1;
LABEL_550:
                              sub_18000CE30(v322, v863);
                              v11 = v865;
                              v254 = v313;
                              v17 = v866;
                              v10 = v867;
                              goto LABEL_514;
                            }
                            v653 = (int *)PyNumber_Long(v651);
                            if ( *v651 >= 0 )
                            {
                              v142 = (*(_QWORD *)v651)-- == 1;
                              if ( v142 )
                              {
                                v654 = sub_18000CC40(v651);
                                (*(void (__fastcall **)(int *))(v654 + 48))(v651);
                              }
                            }
                            v322 = a1;
                            if ( v653 == nullptr )
                              goto LABEL_550;
                            *(_DWORD *)(v313 + 40) = 213;
                            v949[0] = v646;
                            v949[1] = v653;
                            v655 = sub_180377750(a1, v647, v949);
                            v67 = *v653 < 0;
                            v656 = (int *)v655;
                            v878 = (int *)v655;
                            if ( !v67 )
                            {
                              v142 = (*(_QWORD *)v653)-- == 1;
                              if ( v142 )
                              {
                                v657 = sub_18000CC40(v653);
                                (*(void (__fastcall **)(int *))(v657 + 48))(v653);
                              }
                            }
                            v322 = a1;
                            if ( v656 == nullptr )
                              goto LABEL_550;
                            v658 = qword_1803BE000;
                            v659 = (_DWORD *)qword_1803BE238;
                            v911 = v656;
                            v660 = sub_180374A40(a1, 2);
                            v661 = (int *)v660;
                            if ( *v659 != -1 )
                              ++*v659;
                            sub_18000CC80(v660, 0, v659);
                            v662 = v867[2];
                            if ( v662 == 0 )
                            {
                              sub_180370FE0(a1, v863, qword_1803BE1B8);
                              sub_18000D1D0(a1, v863[0]);
LABEL_1174:
                              v16 = 215;
                              if ( *v661 >= 0 )
                              {
                                v142 = (*(_QWORD *)v661)-- == 1;
                                if ( v142 )
                                {
                                  v670 = sub_18000CC40(v661);
                                  (*(void (__fastcall **)(int *))(v670 + 48))(v661);
                                  v11 = v865;
                                  v254 = v313;
                                  v17 = v866;
                                  v10 = v867;
                                  goto LABEL_514;
                                }
                              }
                              goto LABEL_1271;
                            }
                            v663 = sub_18037F0C0(a1, v662, qword_1803BE070);
                            v664 = (int *)v663;
                            v665 = a1;
                            if ( v663 == 0 )
                              goto LABEL_1173;
                            v666 = (int *)sub_180375370(a1, v663);
                            if ( *v664 >= 0 )
                            {
                              v142 = (*(_QWORD *)v664)-- == 1;
                              if ( v142 )
                              {
                                v667 = sub_18000CC40(v664);
                                (*(void (__fastcall **)(int *))(v667 + 48))(v664);
                              }
                            }
                            v665 = a1;
                            if ( v666 == nullptr )
                            {
LABEL_1173:
                              sub_18000CE30(v665, v863);
                              goto LABEL_1174;
                            }
                            v668 = sub_180375520(a1, v666, qword_1803BDFE8);
                            if ( *v666 >= 0 )
                            {
                              v142 = (*(_QWORD *)v666)-- == 1;
                              if ( v142 )
                              {
                                v669 = sub_18000CC40(v666);
                                (*(void (__fastcall **)(int *))(v669 + 48))(v666);
                              }
                            }
                            if ( v668 == 0 )
                            {
                              v665 = a1;
                              goto LABEL_1173;
                            }
                            sub_18000CC80(v661, 1, v668);
                            v671 = (int *)PyUnicode_Join(qword_1803BDFE8, v661);
                            if ( *v661 >= 0 )
                            {
                              v142 = (*(_QWORD *)v661)-- == 1;
                              if ( v142 )
                              {
                                v672 = sub_18000CC40(v661);
                                (*(void (__fastcall **)(int *))(v672 + 48))(v661);
                              }
                            }
                            if ( v671 == nullptr )
                            {
                              sub_18000CE30(a1, v863);
                              v11 = v865;
                              v254 = v313;
                              v17 = v866;
                              v16 = 215;
                              v10 = v867;
                              goto LABEL_514;
                            }
                            v673 = (int *)PyDict_NewPresized(5);
                            PyDict_SetItem(v673, v658, v671);
                            if ( *v671 >= 0 )
                            {
                              v142 = (*(_QWORD *)v671)-- == 1;
                              if ( v142 )
                              {
                                v674 = sub_18000CC40(v671);
                                (*(void (__fastcall **)(int *))(v674 + 48))(v671);
                              }
                            }
                            PyDict_SetItem(v673, qword_1803BE058, qword_1803BE240);
                            PyDict_SetItem(v673, qword_1803BE068, v872);
                            PyDict_SetItem(v673, qword_1803BE248, v656);
                            v675 = qword_1803BE0C0;
                            v676 = sub_1801F4170(a1);
                            if ( v676 == 0 )
                            {
                              sub_180373340(a1, v863, qword_1803BE0C8);
                              v16 = 219;
LABEL_1189:
                              if ( *v673 >= 0 )
                              {
                                v142 = (*(_QWORD *)v673)-- == 1;
                                if ( v142 )
                                {
                                  v680 = sub_18000CC40(v673);
                                  (*(void (__fastcall **)(int *))(v680 + 48))(v673);
                                  v11 = v865;
                                  v17 = v866;
                                  goto LABEL_512;
                                }
                              }
                              goto LABEL_1271;
                            }
                            v677 = v864;
                            v16 = 219;
                            *(_DWORD *)(v864 + 40) = 219;
                            v678 = sub_180376210(a1, v676);
                            v679 = (int *)v678;
                            if ( v678 == 0 )
                            {
                              sub_18000CE30(a1, v863);
                              goto LABEL_1189;
                            }
                            PyDict_SetItem(v673, v675, v678);
                            if ( *v679 >= 0 )
                            {
                              v142 = (*(_QWORD *)v679)-- == 1;
                              if ( v142 )
                              {
                                v681 = sub_18000CC40(v679);
                                (*(void (__fastcall **)(int *))(v681 + 48))(v679);
                              }
                            }
                            v682 = v877;
                            v907 = v673;
                            v683 = sub_180001510(v877);
                            if ( v683 == -1 )
                            {
                              sub_18000CE30(a1, v863);
                              v11 = v865;
                              v254 = v677;
                              v10 = v867;
                              v16 = 221;
                              v17 = v866;
                              goto LABEL_514;
                            }
                            if ( v683 != 0 )
                              PyDict_SetItem(v673, qword_1803BE180, v682);
                            v684 = v875;
                            if ( v875 == nullptr )
                            {
                              sub_180370FE0(a1, v863, qword_1803BE1E8);
                              sub_18000D1D0(a1, v863[0]);
                              v11 = v865;
                              v254 = v677;
                              v10 = v867;
                              v16 = 223;
                              v17 = v866;
                              goto LABEL_514;
                            }
                            v685 = sub_180001510(v875);
                            if ( v685 == -1 )
                            {
                              sub_18000CE30(a1, v863);
                              v11 = v865;
                              v254 = v677;
                              v10 = v867;
                              v16 = 223;
                              v17 = v866;
                              goto LABEL_514;
                            }
                            if ( v685 != 0 )
                            {
                              PyDict_SetItem(v673, qword_1803BE1E8, v684);
                              PyDict_SetItem(v673, qword_1803BE250, qword_1803BE190);
                              v686 = v867;
                              v687 = v867[2];
                              if ( v687 == 0 )
                              {
                                sub_180370FE0(a1, v863, qword_1803BE1B8);
                                sub_18000D1D0(a1, v863[0]);
                                v11 = v865;
                                v254 = v677;
                                v10 = v867;
                                v16 = 226;
                                v17 = v866;
                                goto LABEL_514;
                              }
                              if ( (unsigned __int8)sub_18037F3A0(a1, v687, qword_1803BE258, v684) == 0 )
                              {
                                sub_18000CE30(a1, v863);
                                v11 = v865;
                                v254 = v677;
                                v10 = v867;
                                v16 = 226;
                                v17 = v866;
                                goto LABEL_514;
                              }
                            }
                            else
                            {
                              v686 = v867;
                            }
                            v688 = v686[2];
                            if ( v688 == 0 )
                            {
                              sub_180370FE0(a1, v863, qword_1803BE1B8);
                              sub_18000D1D0(a1, v863[0]);
                              v11 = v865;
                              v254 = v677;
                              v10 = v867;
                              v16 = 228;
                              v17 = v866;
                              goto LABEL_514;
                            }
                            v689 = sub_18037F0C0(a1, v688, qword_1803BE070);
                            v690 = (int *)v689;
                            v691 = a1;
                            if ( v689 != 0 )
                            {
                              v692 = (int *)sub_18037F0C0(a1, v689, qword_1803BE0B8);
                              if ( *v690 >= 0 )
                              {
                                v142 = (*(_QWORD *)v690)-- == 1;
                                if ( v142 )
                                {
                                  v693 = sub_18000CC40(v690);
                                  (*(void (__fastcall **)(int *))(v693 + 48))(v690);
                                }
                              }
                              v691 = a1;
                              if ( v692 != nullptr )
                              {
                                v16 = 228;
                                *(_DWORD *)(v677 + 40) = 228;
                                v694 = (int *)sub_180376770(a1, v692, v673);
                                if ( *v692 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v692)-- == 1;
                                  if ( v142 )
                                  {
                                    v695 = sub_18000CC40(v692);
                                    (*(void (__fastcall **)(int *))(v695 + 48))(v692);
                                  }
                                }
                                if ( v694 == nullptr )
                                {
LABEL_1221:
                                  sub_18000CE30(a1, v863);
                                  v11 = v865;
                                  v254 = v677;
                                  v10 = v867;
                                  v17 = v866;
                                  goto LABEL_514;
                                }
                                if ( *v694 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v694)-- == 1;
                                  if ( v142 )
                                  {
                                    v696 = sub_18000CC40(v694);
                                    (*(void (__fastcall **)(int *))(v696 + 48))(v694);
                                  }
                                }
                                v697 = v686[2];
                                if ( v697 == 0 )
                                {
                                  sub_180370FE0(a1, v863, qword_1803BE1B8);
                                  sub_18000D1D0(a1, v863[0]);
                                  v11 = v865;
                                  v254 = v677;
                                  v10 = v867;
                                  v16 = 229;
                                  v17 = v866;
                                  goto LABEL_514;
                                }
                                v698 = sub_18037F3A0(a1, v697, qword_1803BE0D0, Py_FalseStruct);
                                v699 = a1;
                                if ( v698 == 0 )
                                {
                                  sub_18000CE30(a1, v863);
                                  v11 = v865;
                                  v254 = v677;
                                  v10 = v867;
                                  v16 = 229;
                                  v17 = v866;
                                  goto LABEL_514;
                                }
                                v700 = *((_QWORD *)v871 + 2);
                                if ( v700 != 0 )
                                {
                                  v702 = sub_18037F0C0(a1, v700, qword_1803BDF10);
                                  v703 = (int *)v702;
                                  if ( v702 == 0 )
                                    goto LABEL_1233;
                                  v704 = (int *)sub_18037F0C0(a1, v702, qword_1803BE0D8);
                                  if ( *v703 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v703)-- == 1;
                                    if ( v142 )
                                    {
                                      v705 = sub_18000CC40(v703);
                                      (*(void (__fastcall **)(int *))(v705 + 48))(v703);
                                    }
                                  }
                                  if ( v704 == nullptr )
                                  {
LABEL_1233:
                                    sub_18000CE30(a1, v863);
                                    v11 = v865;
                                    v254 = v677;
                                    v10 = v867;
                                    v16 = 230;
                                    v17 = v866;
                                    goto LABEL_514;
                                  }
                                  v706 = v686[2];
                                  if ( v706 != 0 )
                                  {
                                    v16 = 230;
                                    *(_DWORD *)(v677 + 40) = 230;
                                    v708 = (int *)sub_180376770(a1, v704, v706);
                                    if ( *v704 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v704)-- == 1;
                                      if ( v142 )
                                      {
                                        v709 = sub_18000CC40(v704);
                                        (*(void (__fastcall **)(int *))(v709 + 48))(v704);
                                      }
                                    }
                                    if ( v708 != nullptr )
                                    {
                                      if ( *v708 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v708)-- == 1;
                                        if ( v142 )
                                        {
                                          v710 = sub_18000CC40(v708);
                                          (*(void (__fastcall **)(int *))(v710 + 48))(v708);
                                        }
                                      }
                                      v711 = *((_QWORD *)v871 + 2);
                                      if ( v711 == 0 )
                                      {
                                        sub_180370FE0(a1, v863, qword_1803BE198);
                                        sub_18000D1D0(a1, v863[0]);
                                        v11 = v865;
                                        v254 = v677;
                                        v10 = v867;
                                        v16 = 232;
                                        v17 = v866;
                                        goto LABEL_514;
                                      }
                                      v712 = (int *)sub_18037F0C0(a1, v711, *(_QWORD *)&qword_1803BE0F8);
                                      if ( v712 == nullptr )
                                      {
                                        sub_18000CE30(a1, v863);
                                        v11 = v865;
                                        v16 = 232;
                                        v17 = v866;
                                        goto LABEL_512;
                                      }
                                      v713 = v893[2];
                                      v714 = *((_QWORD *)v902 + 2);
                                      v715 = (int *)sub_180370EC0(a1);
                                      v16 = 232;
                                      *(_DWORD *)(v864 + 40) = 232;
                                      v958[2] = v872;
                                      v958[1] = v714;
                                      v716 = v875;
                                      v958[4] = v878;
                                      v958[0] = v713;
                                      v958[3] = v875;
                                      v958[5] = v715;
                                      v717 = (int *)sub_18037D280(a1, v712, qword_1803BE260, v958, qword_1803BE268);
                                      if ( *v712 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v712)-- == 1;
                                        if ( v142 )
                                        {
                                          v718 = sub_18000CC40(v712);
                                          (*(void (__fastcall **)(int *))(v718 + 48))(v712);
                                        }
                                      }
                                      if ( *v715 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v715)-- == 1;
                                        if ( v142 )
                                        {
                                          v719 = sub_18000CC40(v715);
                                          (*(void (__fastcall **)(int *))(v719 + 48))(v715);
                                        }
                                      }
                                      if ( v717 == nullptr )
                                        goto LABEL_1261;
                                      if ( *v717 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v717)-- == 1;
                                        if ( v142 )
                                        {
                                          v720 = sub_18000CC40(v717);
                                          (*(void (__fastcall **)(int *))(v720 + 48))(v717);
                                        }
                                      }
                                      v721 = qword_1803BE270;
                                      v37 = (int *)PyDict_NewPresized(3);
                                      PyDict_SetItem(v37, v721, v872);
                                      PyDict_SetItem(v37, qword_1803BE1E8, v716);
                                      v722 = qword_1803BE1B8;
                                      v723 = v867[2];
                                      if ( v723 != 0 )
                                      {
                                        v16 = 241;
                                        *(_DWORD *)(v864 + 40) = 241;
                                        v724 = sub_18037E560(a1, v723, qword_1803BE278);
                                        v725 = (int *)v724;
                                        if ( v724 != 0 )
                                        {
                                          PyDict_SetItem(v37, v722, v724);
                                          if ( *v725 >= 0 )
                                          {
                                            v142 = (*(_QWORD *)v725)-- == 1;
                                            if ( v142 )
                                            {
                                              v727 = sub_18000CC40(v725);
                                              (*(void (__fastcall **)(int *))(v727 + 48))(v725);
                                            }
                                          }
                                          sub_180001D80(a1, v726);
                                          v728 = v905;
                                          v729 = v879;
                                          v730 = v870;
                                          v731 = v894;
                                          v67 = *v905 < 0;
                                          v732 = v874;
                                          v733 = v878;
                                          v876 = v937;
                                          v883 = v897;
                                          v865 = v900;
                                          v881 = v936;
                                          v904 = v935;
                                          v884 = v938;
                                          v901 = v925;
                                          v899 = v926;
                                          v913 = v889;
                                          v895 = v927;
                                          v890 = v928;
                                          v866 = v929;
                                          v891 = v930;
                                          v892 = v919;
                                          v912 = v896;
                                          v872 = v887;
                                          v877 = v880;
                                          v903 = v869;
                                          v898 = v888;
                                          v907 = v673;
                                          v885 = v879;
                                          v906 = v870;
                                          v873 = v894;
                                          v886 = v874;
                                          v911 = v878;
                                          if ( v67 )
                                          {
                                            v734 = (int *)v882;
                                          }
                                          else
                                          {
                                            v142 = (*(_QWORD *)v905)-- == 1;
                                            if ( v142 )
                                              (*(void (**)(void))(*((_QWORD *)v728 + 1) + 48LL))();
                                            v734 = (int *)v882;
                                            v876 = v937;
                                            v883 = v897;
                                            v865 = v900;
                                            v881 = v936;
                                            v904 = v935;
                                            v884 = v938;
                                            v901 = v925;
                                            v899 = v926;
                                            v913 = v889;
                                            v895 = v927;
                                            v890 = v928;
                                            v866 = v929;
                                            v891 = v930;
                                            v892 = v919;
                                            v912 = v896;
                                            v872 = v887;
                                            v877 = v880;
                                            v903 = v869;
                                            v898 = v888;
                                            v907 = v673;
                                            v885 = v729;
                                            v906 = v730;
                                            v873 = v731;
                                            v886 = v732;
                                            v911 = v733;
                                          }
                                          if ( *v734 >= 0 )
                                          {
                                            v142 = (*(_QWORD *)v734)-- == 1;
                                            if ( v142 )
                                              (*(void (**)(void))(*((_QWORD *)v734 + 1) + 48LL))();
                                          }
                                          v735 = v875;
                                          if ( *v875 >= 0 )
                                          {
                                            v142 = (*(_QWORD *)v875)-- == 1;
                                            if ( v142 )
                                              (*(void (**)(void))(*((_QWORD *)v735 + 1) + 48LL))();
                                          }
                                          v34 = v910;
                                          v33 = v871;
                                          v24 = v933;
                                          goto LABEL_1287;
                                        }
                                        sub_18000CE30(a1, v863);
                                      }
                                      else
                                      {
                                        sub_180370FE0(a1, v863, qword_1803BE1B8);
                                        sub_18000D1D0(a1, v863[0]);
                                        v16 = 241;
                                      }
                                      if ( *v37 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v37)-- == 1;
                                        if ( v142 )
                                        {
LABEL_538:
                                          v317 = sub_18000CC40(v37);
                                          (*(void (__fastcall **)(int *))(v317 + 48))(v37);
                                          v11 = v865;
                                          v17 = v866;
                                          goto LABEL_512;
                                        }
                                      }
LABEL_1271:
                                      v10 = v867;
                                      v11 = v865;
                                      v17 = v866;
                                      goto LABEL_513;
                                    }
                                    goto LABEL_1221;
                                  }
                                  if ( *v704 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v704)-- == 1;
                                    if ( v142 )
                                    {
                                      v707 = sub_18000CC40(v704);
                                      (*(void (__fastcall **)(int *))(v707 + 48))(v704);
                                    }
                                  }
                                  v701 = qword_1803BE1B8;
                                  v699 = a1;
                                }
                                else
                                {
                                  v701 = qword_1803BE198;
                                }
                                sub_180370FE0(v699, v863, v701);
                                sub_18000D1D0(a1, v863[0]);
                                v11 = v865;
                                v254 = v677;
                                v10 = v867;
                                v16 = 230;
                                v17 = v866;
                                goto LABEL_514;
                              }
                            }
                            sub_18000CE30(v691, v863);
                            v11 = v865;
                            v254 = v677;
                            v10 = v867;
                            v16 = 228;
                            v17 = v866;
                            goto LABEL_514;
                          }
LABEL_772:
                          sub_18000CE30(a1, v863);
                          v399 = 176;
                          goto LABEL_896;
                        }
                        v399 = 177;
                        while ( 1 )
                        {
                          v440 = (int *)v887;
                          v441 = (int *)v873;
                          v442 = v887;
                          v887 = v439;
                          v443 = (int *)v873;
                          if ( v442 != nullptr && *v440 >= 0 )
                          {
                            v142 = (*(_QWORD *)v440)-- == 1;
                            if ( v142 )
                            {
                              v444 = sub_18000CC40(v440);
                              (*(void (__fastcall **)(_QWORD *))(v444 + 48))(v442);
                              v441 = v443;
                            }
                          }
                          v445 = *v439;
                          v446 = v441;
                          v873 = v439;
                          v447 = v445 + 1;
                          if ( v447 != 0 )
                            *v439 = v447;
                          if ( v441 != nullptr && *v443 >= 0 )
                          {
                            v142 = (*(_QWORD *)v443)-- == 1;
                            if ( v142 )
                            {
                              v448 = sub_18000CC40(v441);
                              (*(void (__fastcall **)(int *))(v448 + 48))(v446);
                            }
                          }
                          v449 = (int *)sub_18037F0C0(a1, v439, qword_1803BDFD8);
                          if ( v449 == nullptr )
                            goto LABEL_903;
                          *(_DWORD *)(v864 + 40) = 177;
                          v450 = (int *)sub_1803770D0(a1, v449, qword_1803BE188);
                          if ( *v449 >= 0 )
                          {
                            v142 = (*(_QWORD *)v449)-- == 1;
                            if ( v142 )
                            {
                              v451 = sub_18000CC40(v449);
                              (*(void (__fastcall **)(int *))(v451 + 48))(v449);
                            }
                          }
                          if ( v450 == nullptr )
                            goto LABEL_903;
                          v452 = sub_1803836D0(v450, qword_1803BE190);
                          if ( *v450 >= 0 )
                          {
                            v142 = (*(_QWORD *)v450)-- == 1;
                            if ( v142 )
                            {
                              v453 = sub_18000CC40(v450);
                              (*(void (__fastcall **)(int *))(v453 + 48))(v450);
                            }
                          }
                          if ( v452 == -1 )
                          {
LABEL_903:
                            sub_18000CE30(a1, v863);
                            goto LABEL_896;
                          }
                          if ( v452 == 1 )
                            break;
                          if ( (unsigned __int8)sub_18000D5B0(a1) == 0 )
                            goto LABEL_772;
                          v439 = (_DWORD *)sub_18000D430(v438);
                          if ( v439 == nullptr )
                            goto LABEL_771;
                        }
                        v454 = *((_QWORD *)v871 + 2);
                        if ( v454 == 0 )
                        {
                          sub_180370FE0(a1, v863, qword_1803BE198);
                          sub_18000D1D0(a1, v863[0]);
                          v399 = 178;
                          goto LABEL_896;
                        }
                        v455 = (int *)sub_18037F0C0(a1, v454, *(_QWORD *)&qword_1803BE1A0);
                        if ( v455 == nullptr )
                        {
                          sub_18000CE30(a1, v863);
                          v399 = 178;
                          goto LABEL_896;
                        }
                        v456 = (int *)sub_18037F0C0(a1, v439, qword_1803BDFD8);
                        if ( v456 == nullptr )
                        {
                          sub_18000CE30(a1, v863);
                          if ( *v455 >= 0 )
                          {
                            v142 = (*(_QWORD *)v455)-- == 1;
                            if ( v142 )
                            {
                              v457 = sub_18000CC40(v455);
                              (*(void (__fastcall **)(int *))(v457 + 48))(v455);
                            }
                          }
                          v399 = 178;
                          goto LABEL_896;
                        }
                        v399 = 178;
                        *(_DWORD *)(v864 + 40) = 178;
                        v458 = (int *)sub_180378210(a1, v456, qword_1803BE1A8);
                        if ( *v456 >= 0 )
                        {
                          v142 = (*(_QWORD *)v456)-- == 1;
                          if ( v142 )
                          {
                            v459 = sub_18000CC40(v456);
                            (*(void (__fastcall **)(int *))(v459 + 48))(v456);
                          }
                        }
                        if ( v458 == nullptr )
                        {
                          sub_18000CE30(a1, v863);
                          if ( *v455 >= 0 )
                          {
                            v142 = (*(_QWORD *)v455)-- == 1;
                            if ( v142 )
                            {
                              v460 = sub_18000CC40(v455);
                              (*(void (__fastcall **)(int *))(v460 + 48))(v455);
                            }
                          }
                          goto LABEL_896;
                        }
                        *(_DWORD *)(v864 + 40) = 178;
                        v461 = (int *)sub_180376770(a1, v455, v458);
                        if ( *v455 >= 0 )
                        {
                          v142 = (*(_QWORD *)v455)-- == 1;
                          if ( v142 )
                          {
                            v462 = sub_18000CC40(v455);
                            (*(void (__fastcall **)(int *))(v462 + 48))(v455);
                          }
                        }
                        if ( *v458 >= 0 )
                        {
                          v142 = (*(_QWORD *)v458)-- == 1;
                          if ( v142 )
                          {
                            v463 = sub_18000CC40(v458);
                            (*(void (__fastcall **)(int *))(v463 + 48))(v458);
                          }
                        }
                        if ( v461 != nullptr )
                        {
                          v464 = v870;
                          v875 = v461;
                          if ( *(int *)v870 >= 0 )
                          {
                            v142 = (*v870)-- == 1;
                            if ( v142 )
                            {
                              v465 = sub_18000CC40(v464);
                              (*(void (__fastcall **)(_QWORD *))(v465 + 48))(v464);
                            }
                          }
                          v466 = sub_1801F4280(a1);
                          if ( v466 == 0 )
                          {
                            sub_180373340(a1, v863, qword_1803BE1B0);
                            v399 = 179;
LABEL_895:
                            v438 = (int *)v868;
LABEL_896:
                            v407 = v863[0];
                            sub_18000CE70(v863);
                            v514 = v873;
                            if ( v873 != nullptr && *(int *)v873 >= 0 )
                            {
                              v142 = (*v873)-- == 1;
                              if ( v142 )
                              {
                                v515 = sub_18000CC40(v514);
                                (*(void (__fastcall **)(_QWORD *))(v515 + 48))(v514);
                              }
                            }
                            if ( *v438 < 0 )
                              goto LABEL_686;
                            v142 = (*(_QWORD *)v438)-- == 1;
                            if ( !v142 )
                              goto LABEL_686;
                            v516 = sub_18000CC40(v438);
                            (*(void (__fastcall **)(int *))(v516 + 48))(v438);
                            v863[0] = v407;
                            v400 = v864;
LABEL_909:
                            v918 = v863[0];
                            sub_18000CE70(v863);
                            v934 = sub_18000CCC0(a1);
                            v519 = sub_18000CE80(&v918);
                            if ( v519 != 0 )
                              v520 = sub_18000CC90(v519, v400, v399);
                            else
                              v520 = sub_18038E2A0(v400, v399);
                            sub_180001300(&v918, v520);
                            sub_18000D050(a1, &v918);
                            v521 = sub_18000CC40(*(_QWORD *)(a1 + 248));
                            if ( (unsigned int)sub_18000CE90(a1, v521, PyExc_Exception) == 0 )
                            {
                              v16 = 163;
                              if ( (unsigned __int8)sub_180373290(a1, v863) != 0 )
                                v16 = 0;
                              v779 = sub_18000CE80(v863);
                              if ( v779 != 0 && *(_QWORD *)(v779 + 24) == v400 )
                                *(_DWORD *)(v400 + 40) = *(_DWORD *)(v779 + 36);
                              goto LABEL_1451;
                            }
                            v522 = *((_QWORD *)v871 + 2);
                            if ( v522 == 0 )
                            {
                              sub_180370FE0(a1, v863, qword_1803BE198);
                              sub_18000D1D0(a1, v863[0]);
                              v16 = 193;
LABEL_1451:
                              v62 = v863[0];
                              sub_18000CE70(v863);
                              sub_18000CCE0(a1, &v934);
                              v11 = v865;
                              goto LABEL_510;
                            }
                            v523 = sub_18037F0C0(a1, v522, qword_1803BDF38);
                            v524 = (int *)v523;
                            if ( v523 != 0 )
                            {
                              v525 = (int *)sub_18037F0C0(a1, v523, qword_1803BE200);
                              if ( *v524 >= 0 )
                              {
                                v142 = (*(_QWORD *)v524)-- == 1;
                                if ( v142 )
                                {
                                  v526 = sub_18000CC40(v524);
                                  (*(void (__fastcall **)(int *))(v526 + 48))(v524);
                                }
                              }
                              if ( v525 != nullptr )
                              {
                                v527 = qword_1803BE070;
                                v870 = (_QWORD *)qword_1803BE070;
                                v528 = (int *)sub_180001930(a1, v882);
                                if ( v528 == nullptr )
                                {
                                  sub_18000CE30(a1, v863);
                                  v529 = v863[0];
                                  sub_18000CE70(v863);
                                  v863[0] = v529;
                                  v16 = 194;
                                  goto LABEL_1451;
                                }
                                v530 = (int *)sub_180372160(a1, 0);
                                v531 = (int *)sub_180038D80(v528);
                                if ( v531 != nullptr )
                                {
                                  while ( 1 )
                                  {
                                    v532 = v878;
                                    v533 = v917;
                                    v534 = v878;
                                    v878 = v531;
                                    v535 = v917;
                                    if ( v534 != nullptr && *v532 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v532)-- == 1;
                                      if ( v142 )
                                      {
                                        v536 = sub_18000CC40(v532);
                                        (*(void (__fastcall **)(int *))(v536 + 48))(v534);
                                        v533 = v535;
                                      }
                                    }
                                    v537 = *v531;
                                    v538 = v533;
                                    v917 = v531;
                                    v539 = v537 + 1;
                                    if ( v539 != 0 )
                                      *v531 = v539;
                                    if ( v533 != nullptr && *v535 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v535)-- == 1;
                                      if ( v142 )
                                      {
                                        v540 = sub_18000CC40(v533);
                                        (*(void (__fastcall **)(int *))(v540 + 48))(v538);
                                      }
                                    }
                                    v541 = sub_180374A40(a1, 1);
                                    v542 = v541;
                                    if ( *v531 != -1 )
                                      ++*v531;
                                    sub_18000CC80(v541, 0, v531);
                                    v942 = v542;
                                    v543 = sub_180096610(a1, &v942);
                                    if ( v543 == 0
                                      || (unsigned __int8)sub_180372560(v530, v543) == 0
                                      || (unsigned __int8)sub_18000D5B0(a1) == 0 )
                                    {
                                      goto LABEL_1431;
                                    }
                                    v531 = (int *)sub_180038D80(v528);
                                    if ( v531 == nullptr )
                                    {
                                      v527 = (__int64)v870;
                                      break;
                                    }
                                  }
                                }
                                if ( (unsigned __int8)sub_1800012A0(a1) == 0 )
                                {
LABEL_1431:
                                  sub_18000CE30(a1, v863);
                                  v773 = v863[0];
                                  sub_18000CE70(v863);
                                  if ( *v528 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v528)-- == 1;
                                    if ( v142 )
                                    {
                                      v774 = sub_18000CC40(v528);
                                      (*(void (__fastcall **)(int *))(v774 + 48))(v528);
                                    }
                                  }
                                  if ( *v530 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v530)-- == 1;
                                    if ( v142 )
                                    {
                                      v775 = sub_18000CC40(v530);
                                      (*(void (__fastcall **)(int *))(v775 + 48))(v530);
                                    }
                                  }
                                  v776 = v917;
                                  if ( v917 != nullptr && *v917 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v917)-- == 1;
                                    if ( v142 )
                                    {
                                      v777 = sub_18000CC40(v776);
                                      (*(void (__fastcall **)(int *))(v777 + 48))(v776);
                                    }
                                  }
                                  v863[0] = v773;
                                  sub_18000CE70(v863);
                                  if ( v776 != nullptr && *v776 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v776)-- == 1;
                                    if ( v142 )
                                    {
                                      v778 = sub_18000CC40(v776);
                                      (*(void (__fastcall **)(int *))(v778 + 48))(v776);
                                    }
                                  }
                                  v863[0] = v773;
                                  v16 = 194;
                                  goto LABEL_1451;
                                }
                                if ( *v530 != -1 )
                                  ++*v530;
                                if ( *v528 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v528)-- == 1;
                                  if ( v142 )
                                  {
                                    v544 = sub_18000CC40(v528);
                                    (*(void (__fastcall **)(int *))(v544 + 48))(v528);
                                  }
                                }
                                if ( *v530 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v530)-- == 1;
                                  if ( v142 )
                                  {
                                    v545 = sub_18000CC40(v530);
                                    (*(void (__fastcall **)(int *))(v545 + 48))(v530);
                                  }
                                }
                                v546 = v917;
                                if ( v917 != nullptr && *v917 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v917)-- == 1;
                                  v547 = *(int **)v546;
                                  if ( v142 )
                                  {
                                    v548 = sub_18000CC40(v546);
                                    (*(void (__fastcall **)(int *))(v548 + 48))(v546);
                                    LODWORD(v547) = *v546;
                                  }
                                  if ( (int)v547 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v546)-- == 1;
                                    v878 = *(int **)v546;
                                    if ( v142 )
                                    {
                                      v549 = sub_18000CC40(v546);
                                      (*(void (__fastcall **)(int *))(v549 + 48))(v546);
                                    }
                                  }
                                }
                                v550 = (int *)PyDict_NewPresized(6);
                                PyDict_SetItem(v550, v527, v530);
                                if ( *v530 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v530)-- == 1;
                                  if ( v142 )
                                  {
                                    v551 = sub_18000CC40(v530);
                                    (*(void (__fastcall **)(int *))(v551 + 48))(v530);
                                  }
                                }
                                PyDict_SetItem(v550, qword_1803BE158, v908);
                                if ( v916 == nullptr )
                                {
                                  sub_180370FE0(a1, v863, qword_1803BE160);
                                  sub_18000D1D0(a1, v863[0]);
                                  v16 = 196;
                                  if ( *v525 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v525)-- == 1;
                                    if ( v142 )
                                    {
                                      v552 = sub_18000CC40(v525);
                                      (*(void (__fastcall **)(int *))(v552 + 48))(v525);
                                    }
                                  }
                                  goto LABEL_964;
                                }
                                PyDict_SetItem(v550, qword_1803BE160, v916);
                                v554 = v871;
                                v555 = qword_1803BE168;
                                v954[0] = v871;
                                if ( *v871 != -1 )
                                  ++*v871;
                                v954[1] = v867;
                                if ( *(_DWORD *)v867 != -1 )
                                  ++*(_DWORD *)v867;
                                v556 = v893;
                                v954[2] = v893;
                                if ( *(_DWORD *)v893 != -1 )
                                  ++*(_DWORD *)v893;
                                v557 = (int *)sub_1802171D0(a1, v954);
                                PyDict_SetItem(v550, v555, v557);
                                if ( *v557 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v557)-- == 1;
                                  if ( v142 )
                                  {
                                    v558 = sub_18000CC40(v557);
                                    (*(void (__fastcall **)(int *))(v558 + 48))(v557);
                                  }
                                }
                                v559 = qword_1803BE210;
                                v955[0] = v902;
                                if ( *v902 != -1 )
                                  ++*v902;
                                v955[1] = v554;
                                if ( *v554 != -1 )
                                  ++*v554;
                                v955[2] = v556;
                                if ( *v556 != -1 )
                                  ++*v556;
                                v560 = (int *)sub_180217230(a1, v955);
                                PyDict_SetItem(v550, v559, v560);
                                if ( *v560 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v560)-- == 1;
                                  if ( v142 )
                                  {
                                    v561 = sub_18000CC40(v560);
                                    (*(void (__fastcall **)(int *))(v561 + 48))(v560);
                                  }
                                }
                                PyDict_SetItem(v550, qword_1803BE218, v903);
                                v562 = v909;
                                if ( v909 == nullptr )
                                {
                                  if ( *v525 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v525)-- == 1;
                                    if ( v142 )
                                    {
                                      v563 = sub_18000CC40(v525);
                                      (*(void (__fastcall **)(int *))(v563 + 48))(v525);
                                    }
                                  }
                                  if ( *v550 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v550)-- == 1;
                                    if ( v142 )
                                    {
                                      v564 = sub_18000CC40(v550);
                                      (*(void (__fastcall **)(int *))(v564 + 48))(v550);
                                    }
                                  }
                                  sub_180370FE0(a1, v863, qword_1803BE220);
                                  sub_18000D1D0(a1, v863[0]);
                                  v16 = 202;
                                  goto LABEL_1451;
                                }
                                if ( *v909 != -1 )
                                  ++*v909;
                                v956[2] = v562;
                                v956[0] = v525;
                                v956[1] = v550;
                                v565 = sub_180097770(a1, v956);
                                v566 = v565;
                                if ( v565 != 0 )
                                {
                                  v567 = (int *)v885;
                                  v885 = (_QWORD *)v565;
                                  if ( v567 != nullptr && *v567 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v567)-- == 1;
                                    if ( v142 )
                                    {
                                      v568 = sub_18000CC40(v567);
                                      (*(void (__fastcall **)(int *))(v568 + 48))(v567);
                                    }
                                  }
                                  v569 = *((_QWORD *)v554 + 2);
                                  if ( v569 == 0 )
                                  {
                                    sub_180370FE0(a1, v863, qword_1803BE198);
                                    sub_18000D1D0(a1, v863[0]);
                                    v570 = v863[0];
                                    v16 = 204;
                                    sub_18000CE70(v863);
                                    v863[0] = v570;
                                    goto LABEL_1451;
                                  }
                                  v571 = (int *)sub_18037F0C0(a1, v569, *(_QWORD *)&qword_1803BE228);
                                  if ( v571 == nullptr )
                                  {
                                    sub_18000CE30(a1, v863);
                                    v572 = v863[0];
                                    v16 = 204;
                                    sub_18000CE70(v863);
                                    v863[0] = v572;
                                    goto LABEL_1451;
                                  }
                                  v573 = (int *)sub_18037F0C0(a1, v566, qword_1803BE068);
                                  if ( v573 == nullptr )
                                  {
                                    sub_18000CE30(a1, v863);
                                    if ( *v571 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v571)-- == 1;
                                      if ( v142 )
                                      {
                                        v574 = sub_18000CC40(v571);
                                        (*(void (__fastcall **)(int *))(v574 + 48))(v571);
                                      }
                                    }
                                    v575 = v863[0];
                                    v16 = 204;
                                    sub_18000CE70(v863);
                                    v863[0] = v575;
                                    goto LABEL_1451;
                                  }
                                  v16 = 204;
                                  *(_DWORD *)(v864 + 40) = 204;
                                  v576 = (int *)sub_180376770(a1, v571, v573);
                                  if ( *v571 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v571)-- == 1;
                                    if ( v142 )
                                    {
                                      v577 = sub_18000CC40(v571);
                                      (*(void (__fastcall **)(int *))(v577 + 48))(v571);
                                    }
                                  }
                                  if ( *v573 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v573)-- == 1;
                                    if ( v142 )
                                    {
                                      v578 = sub_18000CC40(v573);
                                      (*(void (__fastcall **)(int *))(v578 + 48))(v573);
                                    }
                                  }
                                  if ( v576 == nullptr )
                                  {
                                    v579 = a1;
LABEL_1022:
                                    sub_18000CE30(v579, v863);
                                    v582 = v863[0];
                                    sub_18000CE70(v863);
                                    v863[0] = v582;
                                    goto LABEL_1451;
                                  }
                                  v580 = (int *)sub_180001A60(v576);
                                  if ( *v576 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v576)-- == 1;
                                    if ( v142 )
                                    {
                                      v581 = sub_18000CC40(v576);
                                      (*(void (__fastcall **)(int *))(v581 + 48))(v576);
                                    }
                                  }
                                  v579 = a1;
                                  if ( v580 == nullptr )
                                    goto LABEL_1022;
                                  v583 = sub_18002A560(a1, v863, v580, 0, 2);
                                  v584 = (int *)v583;
                                  if ( v583 == 0 )
                                  {
                                    v585 = (int *)v879;
LABEL_1027:
                                    v588 = v863[0];
                                    sub_18000CE70(v863);
                                    if ( *v580 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v580)-- == 1;
                                      if ( v142 )
                                      {
                                        v589 = sub_18000CC40(v580);
                                        (*(void (__fastcall **)(int *))(v589 + 48))(v580);
                                      }
                                    }
                                    v863[0] = v588;
                                    sub_18000CE70(v863);
                                    v590 = (int *)v894;
                                    if ( v585 != nullptr && *v585 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v585)-- == 1;
                                      if ( v142 )
                                      {
                                        v591 = sub_18000CC40(v585);
                                        (*(void (__fastcall **)(int *))(v591 + 48))(v585);
                                      }
                                    }
                                    if ( v590 != nullptr && *v590 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v590)-- == 1;
                                      if ( v142 )
                                      {
                                        v592 = sub_18000CC40(v590);
                                        (*(void (__fastcall **)(int *))(v592 + 48))(v590);
                                      }
                                    }
                                    v863[0] = v588;
                                    goto LABEL_1451;
                                  }
                                  v585 = (int *)v583;
                                  v586 = sub_18002A560(a1, v863, v580, 1, 2);
                                  v587 = (int *)v586;
                                  if ( v586 == 0 )
                                    goto LABEL_1027;
                                  v894 = (_QWORD *)v586;
                                  if ( (unsigned __int8)sub_180065DB0(a1, v863, v580, 2) == 0 )
                                    goto LABEL_1027;
                                  if ( *v580 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v580)-- == 1;
                                    if ( v142 )
                                    {
                                      v593 = sub_18000CC40(v580);
                                      (*(void (__fastcall **)(int *))(v593 + 48))(v580);
                                    }
                                  }
                                  v594 = *v584;
                                  v595 = (int *)v872;
                                  v872 = v584;
                                  v596 = v594 + 1;
                                  if ( v596 != 0 )
                                    *v584 = v596;
                                  if ( *v595 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v595)-- == 1;
                                    if ( v142 )
                                    {
                                      v597 = sub_18000CC40(v595);
                                      (*(void (__fastcall **)(int *))(v597 + 48))(v595);
                                    }
                                  }
                                  if ( *v584 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v584)-- == 1;
                                    if ( v142 )
                                    {
                                      v598 = sub_18000CC40(v584);
                                      (*(void (__fastcall **)(int *))(v598 + 48))(v584);
                                    }
                                  }
                                  v599 = *v587;
                                  v600 = v875;
                                  v875 = v587;
                                  v601 = v599 + 1;
                                  if ( v601 != 0 )
                                    *v587 = v601;
                                  if ( v600 != nullptr && *v600 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v600)-- == 1;
                                    if ( v142 )
                                    {
                                      v602 = sub_18000CC40(v600);
                                      (*(void (__fastcall **)(int *))(v602 + 48))(v600);
                                    }
                                  }
                                  if ( *v587 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v587)-- == 1;
                                    if ( v142 )
                                    {
                                      v603 = sub_18000CC40(v587);
                                      (*(void (__fastcall **)(int *))(v603 + 48))(v587);
                                    }
                                  }
                                  v604 = sub_180001510(v587);
                                  if ( v604 == -1 )
                                  {
                                    sub_18000CE30(a1, v863);
                                    v16 = 205;
                                    goto LABEL_1451;
                                  }
                                  if ( v604 != 0 )
                                  {
                                    v605 = sub_1801F4280(a1);
                                    v606 = a1;
                                    if ( v605 == 0 )
                                    {
                                      sub_180373340(a1, v863, qword_1803BE1B0);
                                      v16 = 206;
                                      goto LABEL_1451;
                                    }
                                    v607 = v867;
                                    v608 = v867[2];
                                    if ( v608 == 0 )
                                      goto LABEL_1065;
                                    v609 = (int *)sub_18037F0C0(a1, v608, qword_1803BDFA0);
                                    v610 = v609;
                                    if ( v609 == nullptr )
                                    {
                                      sub_18000CE30(a1, v863);
                                      v16 = 206;
                                      goto LABEL_1451;
                                    }
                                    v611 = v607[2];
                                    if ( v611 == 0 )
                                    {
                                      if ( *v609 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v609)-- == 1;
                                        if ( v142 )
                                        {
                                          v612 = sub_18000CC40(v609);
                                          (*(void (__fastcall **)(int *))(v612 + 48))(v610);
                                        }
                                      }
                                      v606 = a1;
LABEL_1065:
                                      sub_180370FE0(v606, v863, qword_1803BE1B8);
                                      sub_18000D1D0(a1, v863[0]);
LABEL_1066:
                                      v16 = 206;
                                      goto LABEL_1451;
                                    }
                                    v613 = sub_18037F0C0(a1, v611, qword_1803BDFC0);
                                    v614 = (int *)v613;
                                    if ( v613 == 0 )
                                      goto LABEL_1075;
                                    v616 = sub_180001510(v613);
                                    if ( *v614 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v614)-- == 1;
                                      if ( v142 )
                                      {
                                        v617 = sub_18000CC40(v614);
                                        (*(void (__fastcall **)(int *))(v617 + 48))(v614);
                                      }
                                    }
                                    if ( v616 == -1 )
                                    {
LABEL_1075:
                                      sub_18000CE30(a1, v863);
                                      if ( *v610 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v610)-- == 1;
                                        if ( v142 )
                                        {
                                          v615 = sub_18000CC40(v610);
                                          (*(void (__fastcall **)(int *))(v615 + 48))(v610);
                                          v16 = 206;
                                          goto LABEL_1451;
                                        }
                                      }
                                      goto LABEL_1066;
                                    }
                                    v618 = Py_FalseStruct;
                                    v142 = v616 == 0;
                                    v16 = 206;
                                    if ( !v142 )
                                      v618 = Py_TrueStruct;
                                    *(_DWORD *)(v864 + 40) = 206;
                                    v943 = v618;
                                    v948[0] = v610;
                                    v948[1] = v587;
                                    v619 = sub_18037D650(a1, v605, v948, &v943, qword_1803BE140);
                                    if ( *v610 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v610)-- == 1;
                                      if ( v142 )
                                      {
                                        v620 = sub_18000CC40(v610);
                                        (*(void (__fastcall **)(int *))(v620 + 48))(v610);
                                      }
                                    }
                                    if ( v619 == 0 )
                                    {
LABEL_1137:
                                      sub_18000CE30(a1, v863);
                                      goto LABEL_1451;
                                    }
                                    v621 = (int *)v886;
                                    v886 = (_QWORD *)v619;
                                    if ( v621 != nullptr && *v621 >= 0 )
                                    {
                                      v142 = (*(_QWORD *)v621)-- == 1;
                                      if ( v142 )
                                      {
                                        v622 = sub_18000CC40(v621);
                                        (*(void (__fastcall **)(int *))(v622 + 48))(v621);
                                      }
                                    }
                                    v623 = sub_180001510(v619);
                                    if ( v623 == -1 )
                                    {
                                      sub_18000CE30(a1, v863);
                                      v16 = 207;
                                      goto LABEL_1451;
                                    }
                                    if ( v623 != 0 )
                                    {
                                      v875 = (int *)Py_NoneStruct;
                                      if ( Py_NoneStruct != -1 )
                                        ++Py_NoneStruct;
                                      if ( *v587 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v587)-- == 1;
                                        if ( v142 )
                                        {
                                          v624 = sub_18000CC40(v587);
                                          (*(void (__fastcall **)(int *))(v624 + 48))(v587);
                                        }
                                      }
                                    }
                                    else
                                    {
                                      v626 = sub_18037F0C0(a1, v587, qword_1803BDFD8);
                                      v627 = (int *)v626;
                                      if ( v626 == 0 )
                                      {
                                        sub_18000CE30(a1, v863);
                                        v16 = 209;
                                        goto LABEL_1451;
                                      }
                                      v628 = v864;
                                      v16 = 209;
                                      *(_DWORD *)(v864 + 40) = 209;
                                      v550 = (int *)sub_1803770D0(a1, v626, qword_1803BE1D0);
                                      if ( *v627 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v627)-- == 1;
                                        if ( v142 )
                                        {
                                          v629 = sub_18000CC40(v627);
                                          (*(void (__fastcall **)(int *))(v629 + 48))(v627);
                                        }
                                      }
                                      if ( v550 == nullptr )
                                        goto LABEL_1137;
                                      v630 = sub_1801F3630(a1);
                                      if ( v630 == 0 )
                                      {
                                        sub_180373340(a1, v863, qword_1803BE1D8);
LABEL_964:
                                        if ( *v550 >= 0 )
                                        {
                                          v142 = (*(_QWORD *)v550)-- == 1;
                                          if ( v142 )
                                          {
                                            v553 = sub_18000CC40(v550);
                                            (*(void (__fastcall **)(int *))(v553 + 48))(v550);
                                          }
                                        }
                                        goto LABEL_1451;
                                      }
                                      v631 = PySequence_Contains(v630, v550);
                                      if ( *v550 >= 0 )
                                      {
                                        v142 = (*(_QWORD *)v550)-- == 1;
                                        if ( v142 )
                                        {
                                          v632 = sub_18000CC40(v550);
                                          (*(void (__fastcall **)(int *))(v632 + 48))(v550);
                                        }
                                      }
                                      if ( v631 == -1 )
                                        goto LABEL_1137;
                                      if ( v631 == 1 )
                                      {
                                        v633 = *((_QWORD *)v871 + 2);
                                        if ( v633 == 0 )
                                        {
                                          sub_180370FE0(a1, v863, qword_1803BE198);
                                          sub_18000D1D0(a1, v863[0]);
                                          v16 = 210;
                                          goto LABEL_1451;
                                        }
                                        v634 = (int *)sub_18037F0C0(a1, v633, *(_QWORD *)&qword_1803BE1E0);
                                        if ( v634 == nullptr )
                                        {
                                          sub_18000CE30(a1, v863);
                                          v16 = 210;
                                          goto LABEL_1451;
                                        }
                                        v635 = sub_180001B80(a1, v587, qword_1803BE1E8);
                                        v636 = (int *)v635;
                                        if ( v635 == 0 )
                                        {
                                          sub_18000CE30(a1, v863);
                                          if ( *v634 >= 0 )
                                          {
                                            v142 = (*(_QWORD *)v634)-- == 1;
                                            if ( v142 )
                                            {
                                              v637 = sub_18000CC40(v634);
                                              (*(void (__fastcall **)(int *))(v637 + 48))(v634);
                                            }
                                          }
                                          v16 = 210;
                                          goto LABEL_1451;
                                        }
                                        v16 = 210;
                                        *(_DWORD *)(v628 + 40) = 210;
                                        v638 = (int *)sub_180376770(a1, v634, v635);
                                        if ( *v634 >= 0 )
                                        {
                                          v142 = (*(_QWORD *)v634)-- == 1;
                                          if ( v142 )
                                          {
                                            v639 = sub_18000CC40(v634);
                                            (*(void (__fastcall **)(int *))(v639 + 48))(v634);
                                          }
                                        }
                                        if ( *v636 >= 0 )
                                        {
                                          v142 = (*(_QWORD *)v636)-- == 1;
                                          if ( v142 )
                                          {
                                            v640 = sub_18000CC40(v636);
                                            (*(void (__fastcall **)(int *))(v640 + 48))(v636);
                                          }
                                        }
                                        if ( v638 == nullptr )
                                          goto LABEL_1137;
                                        v641 = sub_18000D470(a1, v587, qword_1803BE1F0, v638);
                                        if ( *v638 >= 0 )
                                        {
                                          v142 = (*(_QWORD *)v638)-- == 1;
                                          if ( v142 )
                                          {
                                            v642 = sub_18000CC40(v638);
                                            (*(void (__fastcall **)(int *))(v642 + 48))(v638);
                                          }
                                        }
                                        if ( v641 == 0 )
                                          goto LABEL_1137;
                                      }
                                    }
                                  }
                                  v625 = sub_180374990(a1, qword_1803BDFE8, v915[2]);
                                  if ( v625 == 0 )
                                  {
                                    sub_18000CE30(a1, v863);
                                    v16 = 211;
                                    goto LABEL_1451;
                                  }
                                  v643 = (int *)v877;
                                  v877 = (_QWORD *)v625;
                                  if ( v643 != nullptr && *v643 >= 0 )
                                  {
                                    v142 = (*(_QWORD *)v643)-- == 1;
                                    if ( v142 )
                                    {
                                      v644 = sub_18000CC40(v643);
                                      (*(void (__fastcall **)(int *))(v644 + 48))(v643);
                                    }
                                  }
                                  sub_18000CCE0(a1, &v934);
                                  goto LABEL_1143;
                                }
                              }
                            }
                            sub_18000CE30(a1, v863);
                            v16 = 193;
                            goto LABEL_1451;
                          }
                          v467 = v867;
                          v468 = v867[2];
                          if ( v468 == 0 )
                          {
                            sub_180370FE0(a1, v863, qword_1803BE1B8);
                            sub_18000D1D0(a1, v863[0]);
                            v399 = 180;
                            goto LABEL_895;
                          }
                          v469 = (int *)sub_18037F0C0(a1, v468, qword_1803BDFA0);
                          v470 = v469;
                          if ( v469 == nullptr )
                          {
                            sub_18000CE30(a1, v863);
                            v399 = 180;
                            goto LABEL_895;
                          }
                          v471 = v467[2];
                          if ( v471 == 0 )
                          {
                            if ( *v469 >= 0 )
                            {
                              v142 = (*(_QWORD *)v469)-- == 1;
                              if ( v142 )
                              {
                                v472 = sub_18000CC40(v469);
                                (*(void (__fastcall **)(int *))(v472 + 48))(v470);
                              }
                            }
                            sub_180370FE0(a1, v863, qword_1803BE1B8);
                            sub_18000D1D0(a1, v863[0]);
                            v399 = 182;
                            goto LABEL_895;
                          }
                          v473 = sub_18037F0C0(a1, v471, qword_1803BDFC0);
                          v474 = (int *)v473;
                          if ( v473 == 0 )
                            goto LABEL_811;
                          v476 = sub_180001510(v473);
                          if ( *v474 >= 0 )
                          {
                            v142 = (*(_QWORD *)v474)-- == 1;
                            if ( v142 )
                            {
                              v477 = sub_18000CC40(v474);
                              (*(void (__fastcall **)(int *))(v477 + 48))(v474);
                            }
                          }
                          if ( v476 == -1 )
                          {
LABEL_811:
                            sub_18000CE30(a1, v863);
                            if ( *v470 >= 0 )
                            {
                              v142 = (*(_QWORD *)v470)-- == 1;
                              if ( v142 )
                              {
                                v475 = sub_18000CC40(v470);
                                (*(void (__fastcall **)(int *))(v475 + 48))(v470);
                              }
                            }
                            v399 = 182;
                            goto LABEL_895;
                          }
                          v478 = Py_FalseStruct;
                          v399 = 179;
                          if ( v476 != 0 )
                            v478 = Py_TrueStruct;
                          *(_DWORD *)(v864 + 40) = 179;
                          v941 = v478;
                          v947[0] = v470;
                          v947[1] = v461;
                          v479 = sub_18037D650(a1, v466, v947, &v941, qword_1803BE140);
                          if ( *v470 >= 0 )
                          {
                            v142 = (*(_QWORD *)v470)-- == 1;
                            if ( v142 )
                            {
                              v480 = sub_18000CC40(v470);
                              (*(void (__fastcall **)(int *))(v480 + 48))(v470);
                            }
                          }
                          if ( v479 != 0 )
                          {
                            v886 = (_QWORD *)v479;
                            v481 = sub_180001510(v479);
                            if ( v481 == -1 )
                            {
                              sub_18000CE30(a1, v863);
                              v399 = 184;
                              goto LABEL_895;
                            }
                            if ( v481 != 0 )
                            {
                              v875 = (int *)Py_NoneStruct;
                              if ( Py_NoneStruct != -1 )
                                ++Py_NoneStruct;
                              if ( *v461 >= 0 )
                              {
                                v142 = (*(_QWORD *)v461)-- == 1;
                                if ( v142 )
                                {
                                  v482 = sub_18000CC40(v461);
                                  (*(void (__fastcall **)(int *))(v482 + 48))(v461);
                                }
                              }
                              v483 = sub_180001510(v426);
                              if ( v483 == -1 )
                                goto LABEL_834;
                              if ( v483 == 1 )
                              {
                                if ( *v426 != -1 )
                                  ++*v426;
                                v484 = (__int64)v426;
LABEL_852:
                                v67 = (int)*v426 < 0;
                                v872 = (_QWORD *)v484;
                                if ( !v67 )
                                {
                                  v142 = (*(_QWORD *)v426)-- == 1;
                                  if ( v142 )
                                  {
                                    v492 = sub_18000CC40(v426);
                                    (*(void (__fastcall **)(_DWORD *))(v492 + 48))(v426);
                                  }
                                }
                                goto LABEL_855;
                              }
                              v485 = (_DWORD *)qword_1803BE1C0;
                              v486 = sub_180374A40(a1, 3);
                              v487 = (int *)v486;
                              if ( *v485 != -1 )
                                ++*v485;
                              sub_18000CC80(v486, 0, v485);
                              v488 = sub_180375520(a1, v479, qword_1803BDFE8);
                              if ( v488 != 0 )
                              {
                                sub_18000CC80(v487, 1, v488);
                                v490 = qword_1803BE1C8;
                                if ( *(_DWORD *)qword_1803BE1C8 != -1 )
                                  ++*(_DWORD *)qword_1803BE1C8;
                                sub_18000CC80(v487, 2, v490);
                                v484 = PyUnicode_Join(qword_1803BDFE8, v487);
                                if ( *v487 >= 0 )
                                {
                                  v142 = (*(_QWORD *)v487)-- == 1;
                                  if ( v142 )
                                  {
                                    v491 = sub_18000CC40(v487);
                                    (*(void (__fastcall **)(int *))(v491 + 48))(v487);
                                  }
                                }
                                if ( v484 == 0 )
                                {
LABEL_834:
                                  sub_18000CE30(a1, v863);
                                  v399 = 186;
                                  goto LABEL_895;
                                }
                                goto LABEL_852;
                              }
                              sub_18000CE30(a1, v863);
                              v399 = 186;
LABEL_843:
                              if ( *v487 >= 0 )
                              {
                                v142 = (*(_QWORD *)v487)-- == 1;
                                if ( v142 )
                                {
                                  v489 = sub_18000CC40(v487);
                                  (*(void (__fastcall **)(int *))(v489 + 48))(v487);
                                }
                              }
                              goto LABEL_895;
                            }
                            v496 = sub_18037F0C0(a1, v461, qword_1803BDFD8);
                            v497 = (int *)v496;
                            if ( v496 == 0 )
                            {
                              sub_18000CE30(a1, v863);
                              v399 = 187;
                              goto LABEL_895;
                            }
                            v498 = v864;
                            v399 = 187;
                            *(_DWORD *)(v864 + 40) = 187;
                            v487 = (int *)sub_1803770D0(a1, v496, qword_1803BE1D0);
                            if ( *v497 >= 0 )
                            {
                              v142 = (*(_QWORD *)v497)-- == 1;
                              if ( v142 )
                              {
                                v499 = sub_18000CC40(v497);
                                (*(void (__fastcall **)(int *))(v499 + 48))(v497);
                              }
                            }
                            v500 = a1;
                            if ( v487 == nullptr )
                              goto LABEL_894;
                            v501 = sub_1801F3630(a1);
                            if ( v501 == 0 )
                            {
                              sub_180373340(a1, v863, qword_1803BE1D8);
                              goto LABEL_843;
                            }
                            v502 = PySequence_Contains(v501, v487);
                            if ( *v487 >= 0 )
                            {
                              v142 = (*(_QWORD *)v487)-- == 1;
                              if ( v142 )
                              {
                                v503 = sub_18000CC40(v487);
                                (*(void (__fastcall **)(int *))(v503 + 48))(v487);
                              }
                            }
                            if ( v502 == -1 )
                              goto LABEL_893;
                            if ( v502 != 1 )
                            {
LABEL_855:
                              v438 = (int *)v868;
                              goto LABEL_856;
                            }
                            v504 = *((_QWORD *)v871 + 2);
                            if ( v504 == 0 )
                            {
                              sub_180370FE0(a1, v863, qword_1803BE198);
                              sub_18000D1D0(a1, v863[0]);
                              v399 = 188;
                              goto LABEL_895;
                            }
                            v505 = (int *)sub_18037F0C0(a1, v504, *(_QWORD *)&qword_1803BE1E0);
                            if ( v505 == nullptr )
                            {
                              sub_18000CE30(a1, v863);
                              v399 = 188;
                              goto LABEL_895;
                            }
                            v506 = sub_180001B80(a1, v461, qword_1803BE1E8);
                            v507 = (int *)v506;
                            if ( v506 == 0 )
                            {
                              sub_18000CE30(a1, v863);
                              if ( *v505 >= 0 )
                              {
                                v142 = (*(_QWORD *)v505)-- == 1;
                                if ( v142 )
                                {
                                  v508 = sub_18000CC40(v505);
                                  (*(void (__fastcall **)(int *))(v508 + 48))(v505);
                                }
                              }
                              v399 = 188;
                              goto LABEL_895;
                            }
                            v399 = 188;
                            *(_DWORD *)(v498 + 40) = 188;
                            v509 = (int *)sub_180376770(a1, v505, v506);
                            if ( *v505 >= 0 )
                            {
                              v142 = (*(_QWORD *)v505)-- == 1;
                              if ( v142 )
                              {
                                v510 = sub_18000CC40(v505);
                                (*(void (__fastcall **)(int *))(v510 + 48))(v505);
                              }
                            }
                            if ( *v507 >= 0 )
                            {
                              v142 = (*(_QWORD *)v507)-- == 1;
                              if ( v142 )
                              {
                                v511 = sub_18000CC40(v507);
                                (*(void (__fastcall **)(int *))(v511 + 48))(v507);
                              }
                            }
                            v500 = a1;
                            if ( v509 == nullptr )
                            {
LABEL_894:
                              sub_18000CE30(v500, v863);
                              goto LABEL_895;
                            }
                            v512 = sub_18000D470(a1, v461, qword_1803BE1F0, v509);
                            if ( *v509 >= 0 )
                            {
                              v142 = (*(_QWORD *)v509)-- == 1;
                              if ( v142 )
                              {
                                v513 = sub_18000CC40(v509);
                                (*(void (__fastcall **)(int *))(v513 + 48))(v509);
                              }
                            }
                            if ( v512 != 0 )
                              goto LABEL_855;
                          }
                        }
LABEL_893:
                        v500 = a1;
                        goto LABEL_894;
                      }
                    }
                    sub_18000CE30(a1, v863);
                    goto LABEL_727;
                  }
                }
              }
            }
            v16 = 139;
            sub_18000CE30(a1, v863);
            if ( *v334 < 0 )
              goto LABEL_356;
            v142 = (*(_QWORD *)v334)-- == 1;
            if ( !v142 )
              goto LABEL_356;
            goto LABEL_584;
          }
        }
      }
    }
    sub_18000CE30(a1, v863);
    if ( *v153 >= 0 )
    {
      v142 = (*(_QWORD *)v153)-- == 1;
      if ( !v142 )
        goto LABEL_274;
      v159 = sub_18000CC40(v153);
      v160 = v153;
LABEL_273:
      (*(void (__fastcall **)(int *))(v159 + 48))(v160);
    }
LABEL_274:
    v11 = v865;
    v16 = 106;
    v17 = v866;
    goto LABEL_512;
  }
  v117 = PySet_New(v71);
  v118 = v117;
  if ( v117 == 0 )
  {
    sub_18000CE30(a1, v863);
    v17 = v866;
    v16 = 97;
    goto LABEL_512;
  }
  v904 = (int *)v117;
  v119 = (int *)sub_180001930(a1, v70);
  if ( v119 == nullptr )
  {
    sub_18000CE30(a1, v863);
    v62 = v863[0];
    sub_18000CE70(v863);
    v16 = 98;
    goto LABEL_510;
  }
  v120 = (int *)sub_180372160(a1, 0);
  v121 = (int *)sub_18000D430(v119);
  if ( v121 != nullptr )
  {
    do
    {
      v122 = v897;
      v123 = (int *)v868;
      v124 = v897;
      v897 = v121;
      v125 = (int *)v868;
      if ( v124 != nullptr && *v122 >= 0 )
      {
        v142 = (*(_QWORD *)v122)-- == 1;
        if ( v142 )
        {
          v126 = sub_18000CC40(v122);
          (*(void (__fastcall **)(int *))(v126 + 48))(v124);
          v123 = v125;
        }
      }
      v127 = *v121;
      v128 = v123;
      v868 = v121;
      v129 = v127 + 1;
      if ( v129 != 0 )
        *v121 = v129;
      if ( v123 != nullptr && *v125 >= 0 )
      {
        v142 = (*(_QWORD *)v125)-- == 1;
        if ( v142 )
        {
          v130 = sub_18000CC40(v123);
          (*(void (__fastcall **)(int *))(v130 + 48))(v128);
        }
      }
      v131 = sub_180001B80(a1, v121, qword_1803BE000);
      v132 = (int *)v131;
      if ( v131 == 0 )
        goto LABEL_235;
      v133 = PySet_Contains(v118, v131);
      if ( *v132 >= 0 )
      {
        v142 = (*(_QWORD *)v132)-- == 1;
        if ( v142 )
        {
          v134 = sub_18000CC40(v132);
          (*(void (__fastcall **)(int *))(v134 + 48))(v132);
        }
      }
      if ( v133 == -1
        || v133 == 1 && (unsigned __int8)sub_1803725B0(v120, v121) == 0
        || (unsigned __int8)sub_18000D5B0(a1) == 0 )
      {
        goto LABEL_235;
      }
      v121 = (int *)sub_18000D430(v119);
    }
    while ( v121 != nullptr );
    v70 = v883;
  }
  if ( (unsigned __int8)sub_1800012A0(a1) != 0 )
  {
    if ( *v120 != -1 )
      ++*v120;
    if ( *v119 >= 0 )
    {
      v142 = (*(_QWORD *)v119)-- == 1;
      if ( v142 )
      {
        v135 = sub_18000CC40(v119);
        (*(void (__fastcall **)(int *))(v135 + 48))(v119);
      }
    }
    if ( *v120 >= 0 )
    {
      v142 = (*(_QWORD *)v120)-- == 1;
      if ( v142 )
      {
        v136 = sub_18000CC40(v120);
        (*(void (__fastcall **)(int *))(v136 + 48))(v120);
      }
    }
    v137 = (int **)v868;
    if ( v868 != nullptr && *(int *)v868 >= 0 )
    {
      v142 = (*v868)-- == 1;
      v138 = *v137;
      if ( v142 )
      {
        v139 = sub_18000CC40(v137);
        (*(void (__fastcall **)(int **))(v139 + 48))(v137);
        LODWORD(v138) = *(_DWORD *)v137;
      }
      if ( (int)v138 >= 0 )
      {
        v142 = *v137 == (int *)1;
        *v137 = (int *)((char *)*v137 - 1);
        v897 = *v137;
        if ( v142 )
        {
          v140 = sub_18000CC40(v137);
          (*(void (__fastcall **)(int **))(v140 + 48))(v137);
        }
      }
    }
    v141 = v70;
    v70 = v120;
    v883 = v120;
    if ( *v888 >= 0 )
    {
      v142 = (*(_QWORD *)v888)-- == 1;
LABEL_230:
      if ( v142 )
      {
        v143 = sub_18000CC40(v141);
        (*(void (__fastcall **)(int *))(v143 + 48))(v141);
      }
    }
    goto LABEL_232;
  }
LABEL_235:
  sub_18000CE30(a1, v863);
  v62 = v863[0];
  sub_18000CE70(v863);
  if ( *v119 >= 0 )
  {
    v142 = (*(_QWORD *)v119)-- == 1;
    if ( v142 )
    {
      v145 = sub_18000CC40(v119);
      (*(void (__fastcall **)(int *))(v145 + 48))(v119);
    }
  }
  if ( *v120 >= 0 )
  {
    v142 = (*(_QWORD *)v120)-- == 1;
    if ( v142 )
    {
      v146 = sub_18000CC40(v120);
      (*(void (__fastcall **)(int *))(v146 + 48))(v120);
    }
  }
  v147 = (int *)v868;
  if ( v868 != nullptr && *(int *)v868 >= 0 )
  {
    v142 = (*v868)-- == 1;
    if ( v142 )
    {
      v148 = sub_18000CC40(v147);
      (*(void (__fastcall **)(int *))(v148 + 48))(v147);
    }
  }
  v863[0] = v62;
  sub_18000CE70(v863);
  if ( v147 != nullptr && *v147 >= 0 )
  {
    v142 = (*(_QWORD *)v147)-- == 1;
    if ( v142 )
    {
      v149 = sub_18000CC40(v147);
      (*(void (__fastcall **)(int *))(v149 + 48))(v147);
    }
  }
  v11 = v865;
  v16 = 98;
LABEL_510:
  v17 = v866;
LABEL_511:
  v863[0] = v62;
LABEL_512:
  v10 = v867;
LABEL_513:
  v254 = v864;
LABEL_514:
  v299 = *(_QWORD *)(v863[0] + 40LL);
  if ( v299 == 0 )
  {
    v300 = sub_18038E2A0(v254, v16);
    goto LABEL_1456;
  }
  if ( *(_QWORD *)(v299 + 24) != v254 )
  {
    v300 = sub_18038E2A0(v254, v16);
    *(_QWORD *)(v300 + 16) = v299;
    if ( *(_DWORD *)v299 != -1 )
      ++*(_DWORD *)v299;
LABEL_1456:
    sub_180001300(v863, v300);
  }
  v862 = v885;
  v861 = v898;
  v860 = v903;
  v859 = v877;
  v858 = v872;
  v857 = v912;
  v856 = v902;
  v855 = v892;
  v854 = v891;
  v780 = (int *)v881;
  v853 = v17;
  v852 = v890;
  v851 = v895;
  v850 = v913;
  v849 = v899;
  v848 = v901;
  v847 = v884;
  v846 = v904;
  v781 = v875;
  v845 = v881;
  v844 = v883;
  v843 = v876;
  v842 = v909;
  v841 = v908;
  v840 = v907;
  v839 = v915;
  v838 = v875;
  v837 = v882;
  v836 = v905;
  v835 = v923;
  v834 = v922;
  v833 = v914;
  v832 = v916;
  v831 = v921;
  v830 = v920;
  v827 = v910;
  sub_18036D0A0(v254, "ccoooooooooococoooooooooooooooocoooooooooo", v871);
  v785 = qword_1803BDD90;
  if ( v254 == qword_1803BDD90 )
  {
    if ( *(int *)qword_1803BDD90 >= 0 )
    {
      v142 = (*(_QWORD *)qword_1803BDD90)-- == 1;
      if ( v142 )
        (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *, int *, int *, int *, int *, int *, int *, _QWORD *, int *, _QWORD *, int *, _QWORD *, int *, int *, int *, int *, int *, _QWORD *, int *, int *, int *, int *, int *, _QWORD *, int *, int *, _QWORD *, _QWORD *, int *, int *, _QWORD *, _QWORD *, int *, int *, _QWORD *))(*(_QWORD *)(v785 + 8) + 48LL))(
          v785,
          *(_QWORD *)(*(_QWORD *)(v785 + 8) + 48LL),
          v783,
          v784,
          v827,
          v830,
          v831,
          v832,
          v833,
          v834,
          v835,
          v836,
          v837,
          v838,
          v839,
          v840,
          v10,
          v841,
          v842,
          v843,
          v844,
          v11,
          v845,
          v846,
          v847,
          v848,
          v849,
          v850,
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
          v861,
          v862);
    }
    qword_1803BDD90 = 0;
  }
  sub_180001D80(a1, v782);
  v786 = v905;
  v787 = v863[0];
  v863[0] = 0;
  if ( v905 != nullptr && *v905 >= 0 )
  {
    v142 = (*(_QWORD *)v905)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v786 + 1) + 48LL))();
  }
  v788 = v882;
  if ( v882 != nullptr && *(int *)v882 >= 0 )
  {
    v142 = (*v882)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v788[1] + 48LL))();
  }
  if ( v781 != nullptr && *v781 >= 0 )
  {
    v142 = (*(_QWORD *)v781)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v781 + 1) + 48LL))(v781);
  }
  v789 = v915;
  if ( *(int *)v915 >= 0 )
  {
    v142 = (*v915)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v789[1] + 48LL))();
  }
  v790 = v907;
  if ( v907 != nullptr && *v907 >= 0 )
  {
    v142 = (*(_QWORD *)v907)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v790 + 1) + 48LL))();
  }
  if ( *(int *)v10 >= 0 )
  {
    v142 = (*v10)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(_QWORD *))(v10[1] + 48LL))(v10);
  }
  v791 = v908;
  if ( v908 != nullptr && *v908 >= 0 )
  {
    v142 = (*(_QWORD *)v908)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v791 + 1) + 48LL))();
  }
  v792 = v909;
  if ( v909 != nullptr && *v909 >= 0 )
  {
    v142 = (*(_QWORD *)v909)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v792 + 1) + 48LL))();
  }
  v793 = v876;
  if ( v876 != nullptr && *v876 >= 0 )
  {
    v142 = (*(_QWORD *)v876)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v793 + 1) + 48LL))();
  }
  v794 = v883;
  if ( v883 != nullptr && *v883 >= 0 )
  {
    v142 = (*(_QWORD *)v883)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v794 + 1) + 48LL))();
  }
  if ( v11 != nullptr && *v11 >= 0 )
  {
    v142 = (*(_QWORD *)v11)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v11 + 1) + 48LL))(v11);
  }
  if ( v780 != nullptr && *v780 >= 0 )
  {
    v142 = (*(_QWORD *)v780)-- == 1;
    if ( v142 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v780 + 1) + 48LL))(v780);
  }
  v795 = v904;
  if ( v904 != nullptr && *v904 >= 0 )
  {
    v142 = (*(_QWORD *)v904)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v795 + 1) + 48LL))();
  }
  v796 = v884;
  if ( v884 != nullptr && *v884 >= 0 )
  {
    v142 = (*(_QWORD *)v884)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v796 + 1) + 48LL))();
  }
  v797 = v901;
  if ( v901 != nullptr && *v901 >= 0 )
  {
    v142 = (*(_QWORD *)v901)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v797 + 1) + 48LL))();
  }
  v798 = v899;
  if ( v899 != nullptr && *v899 >= 0 )
  {
    v142 = (*(_QWORD *)v899)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v798 + 1) + 48LL))();
  }
  v799 = v913;
  if ( v913 != nullptr && *v913 >= 0 )
  {
    v142 = (*(_QWORD *)v913)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v799 + 1) + 48LL))();
  }
  v800 = v895;
  if ( v895 != nullptr && *(int *)v895 >= 0 )
  {
    v142 = (*v895)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v800[1] + 48LL))();
  }
  v801 = v890;
  if ( v890 != nullptr && *v890 >= 0 )
  {
    v142 = (*(_QWORD *)v890)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v801 + 1) + 48LL))();
  }
  v802 = v866;
  if ( v866 != nullptr && *v866 >= 0 )
  {
    v142 = (*(_QWORD *)v866)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v802 + 1) + 48LL))();
  }
  v803 = v891;
  if ( v891 != nullptr && *(int *)v891 >= 0 )
  {
    v142 = (*v891)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v803[1] + 48LL))();
  }
  v804 = v892;
  if ( v892 != nullptr && *(int *)v892 >= 0 )
  {
    v142 = (*v892)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v804[1] + 48LL))();
  }
  v805 = v902;
  if ( *v902 >= 0 )
  {
    v142 = (*(_QWORD *)v902)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v805 + 1) + 48LL))();
  }
  v806 = v912;
  if ( v912 != nullptr && *v912 >= 0 )
  {
    v142 = (*(_QWORD *)v912)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v806 + 1) + 48LL))();
  }
  v807 = v872;
  if ( v872 != nullptr && *(int *)v872 >= 0 )
  {
    v142 = (*v872)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v807[1] + 48LL))();
  }
  v808 = v877;
  if ( v877 != nullptr && *(int *)v877 >= 0 )
  {
    v142 = (*v877)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v808[1] + 48LL))();
  }
  v809 = v903;
  if ( v903 != nullptr && *v903 >= 0 )
  {
    v142 = (*(_QWORD *)v903)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v809 + 1) + 48LL))();
  }
  v810 = v898;
  if ( v898 != nullptr && *v898 >= 0 )
  {
    v142 = (*(_QWORD *)v898)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v810 + 1) + 48LL))();
  }
  v811 = v885;
  if ( v885 != nullptr && *(int *)v885 >= 0 )
  {
    v142 = (*v885)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v811[1] + 48LL))();
  }
  v812 = v906;
  if ( v906 != nullptr && *(int *)v906 >= 0 )
  {
    v142 = (*v906)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v812[1] + 48LL))();
  }
  v813 = v873;
  if ( v873 != nullptr && *(int *)v873 >= 0 )
  {
    v142 = (*v873)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v813[1] + 48LL))();
  }
  v814 = v886;
  if ( v886 != nullptr && *(int *)v886 >= 0 )
  {
    v142 = (*v886)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v814[1] + 48LL))();
  }
  v815 = v911;
  if ( v911 != nullptr && *v911 >= 0 )
  {
    v142 = (*(_QWORD *)v911)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v815 + 1) + 48LL))();
  }
  v863[0] = v787;
  v816 = v871;
  if ( *v871 >= 0 )
  {
    v142 = (*(_QWORD *)v871)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v816 + 1) + 48LL))();
  }
  v817 = v893;
  if ( *(int *)v893 >= 0 )
  {
    v142 = (*v893)-- == 1;
    if ( v142 )
      (*(void (**)(void))(v817[1] + 48LL))();
  }
  v818 = v910;
  if ( *v910 >= 0 )
  {
    v142 = (*(_QWORD *)v910)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v818 + 1) + 48LL))();
  }
  v819 = v920;
  if ( *v920 >= 0 )
  {
    v142 = (*(_QWORD *)v920)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v819 + 1) + 48LL))();
  }
  v820 = v921;
  if ( *v921 >= 0 )
  {
    v142 = (*(_QWORD *)v921)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v820 + 1) + 48LL))();
  }
  v821 = v916;
  if ( *v916 >= 0 )
  {
    v142 = (*(_QWORD *)v916)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v821 + 1) + 48LL))();
  }
  v822 = v914;
  if ( *v914 >= 0 )
  {
    v142 = (*(_QWORD *)v914)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v822 + 1) + 48LL))();
  }
  v823 = v922;
  if ( *v922 >= 0 )
  {
    v142 = (*(_QWORD *)v922)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v823 + 1) + 48LL))();
  }
  v824 = v923;
  if ( *v923 >= 0 )
  {
    v142 = (*(_QWORD *)v923)-- == 1;
    if ( v142 )
      (*(void (**)(void))(*((_QWORD *)v824 + 1) + 48LL))();
  }
  sub_18000CE40(a1, v863);
  return nullptr;
}
