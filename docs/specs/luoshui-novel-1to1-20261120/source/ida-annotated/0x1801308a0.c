// Analyst annotations only; original export: source/ida/0x1801308a0.c
// qword comments derive module constant slot + original file range; not runtime values.
__int64 __fastcall sub_1801308A0(__int64 a1, __int64 a2, __int64 a3)
{
  __int64 v4; // rcx
  __int64 *v6; // rax
  __int64 v7; // rbx
  __int64 Length; // rax
  __int64 v9; // rax
  __int64 v10; // rax
  _QWORD *v11; // rax
  __int64 Dict; // rax
  __int64 v13; // rax
  __int64 v14; // rax
  int *AttrString; // rbx
  __int64 v16; // rsi
  bool v17; // zf
  __int64 v18; // r15
  __int64 v19; // rbx
  __int64 v20; // rax
  __int64 v21; // rdx
  unsigned int v22; // esi
  __int64 v23; // rax
  __int64 v24; // rax
  __int64 v25; // rax
  __int64 v26; // rdx
  __int64 v27; // rax
  __int64 v28; // rax
  __int64 v29; // rdx
  __int64 v30; // rax
  __int64 v31; // rax
  __int64 v32; // rdx
  __int64 v33; // r8
  __int64 v34; // rax
  __int64 v35; // rax
  int *v36; // rbx
  __int64 v37; // rax
  __int64 v38; // rsi
  int *v39; // rax
  int *v40; // r14
  __int64 v41; // rax
  __int64 v42; // rbx
  __int64 v43; // rax
  __int64 v44; // rax
  __int64 v45; // rax
  __int64 v46; // rax
  int v47; // edx
  int v48; // r8d
  __int64 v49; // rax
  int *v50; // rbx
  __int64 v51; // rax
  __int64 v52; // rsi
  int v53; // edx
  int v54; // r8d
  __int64 v55; // rax
  int *v56; // rbx
  __int64 v57; // rax
  __int64 v58; // rsi
  __int64 v59; // rax
  int v60; // edx
  int v61; // r8d
  __int64 v62; // rax
  int *v63; // rbx
  __int64 v64; // rax
  __int64 v65; // rsi
  __int64 v66; // rax
  int v67; // edx
  int v68; // r8d
  __int64 v69; // rax
  int *v70; // rbx
  __int64 v71; // rax
  __int64 v72; // rsi
  __int64 v73; // rax
  int v74; // edx
  int v75; // r8d
  __int64 v76; // rax
  int *v77; // rbx
  __int64 v78; // rax
  __int64 v79; // rsi
  __int64 v80; // rax
  int v81; // edx
  int v82; // r8d
  __int64 v83; // rax
  int *v84; // rbx
  __int64 v85; // rax
  __int64 v86; // rsi
  __int64 v87; // rax
  int v88; // edx
  int v89; // r8d
  __int64 v90; // rax
  int *v91; // rbx
  __int64 v92; // rax
  __int64 v93; // rsi
  __int64 v94; // rax
  int v95; // edx
  int v96; // r8d
  __int64 v97; // rax
  int *v98; // rbx
  __int64 v99; // rax
  __int64 v100; // rsi
  __int64 v101; // rax
  int v102; // edx
  int v103; // r8d
  __int64 v104; // rax
  int *v105; // rbx
  __int64 v106; // rax
  __int64 v107; // rsi
  __int64 v108; // rax
  int v109; // edx
  int v110; // r8d
  __int64 v111; // rax
  int *v112; // rbx
  __int64 v113; // rax
  __int64 v114; // rsi
  __int64 v115; // rax
  int v116; // edx
  int v117; // r8d
  __int64 v118; // rax
  int *v119; // rbx
  __int64 v120; // rax
  __int64 v121; // rsi
  __int64 v122; // rax
  int v123; // edx
  int v124; // r8d
  __int64 v125; // rax
  int *v126; // rbx
  __int64 v127; // rax
  __int64 v128; // rsi
  __int64 v129; // rax
  int v130; // edx
  int v131; // r8d
  __int64 v132; // rax
  int *v133; // rbx
  __int64 v134; // rax
  __int64 v135; // rsi
  __int64 v136; // rax
  int v137; // edx
  int v138; // r8d
  __int64 v139; // rax
  int *v140; // r14
  __int64 v141; // rax
  __int64 v142; // rax
  __int64 v143; // rax
  __int64 v144; // rax
  __int64 v145; // rax
  int v146; // edx
  int v147; // r8d
  __int64 v148; // rax
  __int64 v149; // rax
  __int64 v150; // rax
  __int64 v151; // rax
  __int64 v152; // rax
  __int64 v153; // rax
  int v154; // edx
  int v155; // r8d
  __int64 v156; // rax
  int *v157; // rbx
  __int64 v158; // rax
  __int64 v159; // rsi
  __int64 v160; // rax
  int v161; // edx
  int v162; // r8d
  __int64 v163; // rax
  int *v164; // rbx
  __int64 v165; // rax
  __int64 v166; // rsi
  __int64 v167; // rax
  int v168; // edx
  int v169; // r8d
  __int64 v170; // rax
  int *v171; // rbx
  __int64 v172; // rax
  __int64 v173; // rsi
  __int64 v174; // rax
  int v175; // edx
  int v176; // r8d
  __int64 v177; // rax
  __int64 v178; // rax
  __int64 v179; // rax
  __int64 v180; // rax
  __int64 v181; // rax
  int v182; // edx
  int v183; // r8d
  __int64 v184; // rax
  int *v185; // rbx
  __int64 v186; // rax
  __int64 v187; // rsi
  __int64 v188; // rax
  int v189; // edx
  int v190; // r8d
  __int64 v191; // rax
  __int64 v192; // rax
  __int64 v193; // rax
  __int64 v194; // rax
  __int64 v195; // rax
  int v196; // edx
  int v197; // r8d
  __int64 v198; // rax
  int *v199; // rbx
  __int64 v200; // rax
  __int64 v201; // rsi
  __int64 v202; // rax
  int v203; // edx
  int v204; // r8d
  __int64 v205; // rax
  int *v206; // rbx
  __int64 v207; // rax
  __int64 v208; // rsi
  __int64 v209; // rax
  int v210; // edx
  int v211; // r8d
  __int64 v212; // rax
  int *v213; // rbx
  __int64 v214; // rax
  __int64 v215; // rsi
  __int64 v216; // rax
  int v217; // edx
  int v218; // r8d
  __int64 v219; // rax
  int *v220; // rbx
  __int64 v221; // rax
  __int64 v222; // rsi
  __int64 v223; // rax
  int v224; // edx
  int v225; // r8d
  __int64 v226; // rax
  __int64 v227; // rax
  __int64 v228; // rax
  __int64 v229; // rax
  int v230; // edx
  int v231; // r8d
  __int64 v232; // rax
  int *v233; // rbx
  __int64 v234; // rax
  __int64 v235; // rsi
  __int64 v236; // rax
  int v237; // edx
  int v238; // r8d
  __int64 v239; // rax
  __int64 v240; // rax
  __int64 v241; // rax
  __int64 v242; // rax
  int v243; // edx
  int v244; // r8d
  __int64 v245; // rax
  int *v246; // rbx
  __int64 v247; // rax
  __int64 v248; // rsi
  __int64 v249; // rax
  int v250; // edx
  int v251; // r8d
  __int64 v252; // rax
  int *v253; // rbx
  __int64 v254; // rax
  __int64 v255; // rsi
  __int64 v256; // rax
  int v257; // edx
  int v258; // r8d
  __int64 v259; // rax
  int *v260; // rbx
  __int64 v261; // rax
  __int64 v262; // rsi
  __int64 v263; // rax
  int v264; // edx
  int v265; // r8d
  __int64 v266; // rax
  int *v267; // rbx
  __int64 v268; // rax
  __int64 v269; // rsi
  __int64 v270; // rax
  int v271; // edx
  int v272; // r8d
  __int64 v273; // rax
  int *v274; // rbx
  __int64 v275; // rax
  __int64 v276; // rsi
  __int64 v277; // rax
  int v278; // edx
  int v279; // r8d
  __int64 v280; // rax
  int *v281; // rbx
  __int64 v282; // rax
  __int64 v283; // rsi
  __int64 v284; // rax
  int v285; // edx
  int v286; // r8d
  __int64 v287; // rax
  int *v288; // rbx
  __int64 v289; // rax
  __int64 v290; // rsi
  __int64 v291; // rax
  int v292; // edx
  int v293; // r8d
  __int64 v294; // rax
  int *v295; // rbx
  __int64 v296; // rax
  __int64 v297; // rsi
  __int64 v298; // rax
  int v299; // edx
  int v300; // r8d
  __int64 v301; // rax
  __int64 v302; // rax
  __int64 v303; // rax
  __int64 v304; // rax
  int v305; // edx
  int v306; // r8d
  __int64 v307; // rax
  int *v308; // rbx
  __int64 v309; // rax
  __int64 v310; // rsi
  __int64 v311; // rax
  int v312; // edx
  int v313; // r8d
  __int64 v314; // rax
  int *v315; // rbx
  __int64 v316; // rax
  __int64 v317; // rsi
  __int64 v318; // rax
  int *v319; // r13
  _DWORD *v320; // rax
  _DWORD *v321; // r14
  __int64 v322; // rax
  int *v323; // rbx
  __int64 v324; // rax
  int *v325; // rbx
  __int64 v326; // rax
  int *v327; // rbx
  __int64 v328; // rax
  int *v329; // rbx
  __int64 v330; // rax
  _DWORD *v331; // rbx
  __int64 v332; // rax
  int *v333; // rbx
  __int64 v334; // rax
  __int64 v335; // rax
  int *v336; // rbx
  __int64 v337; // rax
  __int64 v338; // rax
  int *v339; // rbx
  __int64 v340; // rax
  __int64 v341; // rax
  int *v342; // rbx
  __int64 v343; // rax
  _DWORD *v344; // rbx
  __int64 v345; // rax
  int *v346; // rbx
  __int64 v347; // rax
  _DWORD *v348; // rbx
  __int64 v349; // rax
  int *v350; // rbx
  __int64 v351; // rax
  __int64 v352; // rax
  int *v353; // rbx
  __int64 v354; // rax
  _DWORD *v355; // rbx
  __int64 v356; // rax
  int *v357; // rbx
  __int64 v358; // rax
  _DWORD *v359; // rbx
  __int64 v360; // rax
  int *v361; // rbx
  __int64 v362; // rax
  _DWORD *v363; // rbx
  __int64 v364; // rsi
  __int64 v365; // rax
  int *v366; // rbx
  __int64 v367; // rax
  __int64 v368; // rax
  int *v369; // rbx
  __int64 v370; // rax
  __int64 v371; // rax
  int *v372; // rbx
  __int64 v373; // rax
  __int64 v374; // rax
  int *v375; // rbx
  __int64 v376; // rax
  __int64 v377; // rax
  int *v378; // rbx
  __int64 v379; // rax
  __int64 v380; // rax
  int *v381; // rbx
  __int64 v382; // rax
  __int64 v383; // rbx
  __int64 v384; // rax
  int *v385; // rbx
  __int64 v386; // rax
  _DWORD *v387; // rbx
  __int64 v388; // rax
  int *v389; // rbx
  __int64 v390; // rax
  _DWORD *v391; // rbx
  __int64 v392; // rax
  int *v393; // rbx
  __int64 v394; // rax
  __int64 v395; // rax
  int *v396; // rbx
  __int64 v397; // rax
  __int64 v398; // rax
  int *v399; // rbx
  __int64 v400; // rax
  __int64 v401; // rax
  int *v402; // rbx
  __int64 v403; // rax
  __int64 v404; // rax
  int *v405; // rbx
  __int64 v406; // rax
  __int64 v407; // rbx
  __int64 v408; // rax
  int *v409; // rbx
  __int64 v410; // rax
  __int64 v411; // rax
  int *v412; // rbx
  __int64 v413; // rax
  __int64 v414; // rax
  int *v415; // rbx
  __int64 v416; // rax
  _DWORD *v417; // rbx
  __int64 v418; // rax
  int *v419; // rbx
  __int64 v420; // rax
  __int64 v421; // rax
  int *v422; // rbx
  __int64 v423; // rax
  __int64 v424; // rax
  int *v425; // rbx
  __int64 v426; // rax
  __int64 v427; // rax
  int *v428; // rbx
  __int64 v429; // rax
  __int64 v430; // rax
  int *v431; // rbx
  __int64 v432; // rax
  _DWORD *v433; // rbx
  __int64 v434; // rax
  int *v435; // rbx
  __int64 v436; // rax
  __int64 v437; // rax
  int *v438; // rbx
  __int64 v439; // rax
  __int64 v440; // rax
  int *v441; // rsi
  int *v442; // rbx
  __int64 v443; // rax
  __int64 v444; // rax
  _DWORD *v445; // rbx
  __int64 v446; // rax
  int *v447; // rbx
  __int64 v448; // rax
  _DWORD *v449; // rbx
  __int64 v450; // rax
  int *v451; // rbx
  __int64 v452; // rax
  _DWORD *v453; // rbx
  __int64 v454; // rax
  int *v455; // rbx
  __int64 v456; // rax
  _DWORD *v457; // rbx
  __int64 v458; // rax
  int *v459; // rbx
  __int64 v460; // rax
  _DWORD *v461; // rbx
  __int64 v462; // rax
  int *v463; // rbx
  __int64 v464; // rax
  _DWORD *v465; // rbx
  __int64 v466; // rax
  int *v467; // rbx
  __int64 v468; // rax
  __int64 v469; // rax
  int *v470; // rbx
  __int64 v471; // rax
  __int64 v472; // rax
  int *v473; // rbx
  __int64 v474; // rax
  __int64 v475; // rax
  int *v476; // rbx
  __int64 v477; // rax
  _DWORD *v478; // rbx
  __int64 v479; // rax
  int *v480; // rbx
  __int64 v481; // rax
  _DWORD *v482; // rbx
  __int64 v483; // rax
  int *v484; // rbx
  __int64 v485; // rax
  _DWORD *v486; // rbx
  __int64 v487; // rax
  int *v488; // rbx
  __int64 v489; // rax
  _DWORD *v490; // rbx
  __int64 v491; // rax
  int *v492; // rbx
  __int64 v493; // rax
  __int64 v494; // rbx
  __int64 v495; // rax
  int *v496; // rbx
  __int64 v497; // rax
  __int64 v498; // rax
  int *v499; // rbx
  __int64 v500; // rax
  __int64 v501; // rax
  int *v502; // rbx
  __int64 v503; // rax
  __int64 v504; // rax
  int *v505; // rbx
  __int64 v506; // rax
  __int64 v507; // rax
  int *v508; // rbx
  __int64 v509; // rax
  _DWORD *v510; // rbx
  __int64 v511; // rax
  int *v512; // r15
  __int64 v513; // r8
  _DWORD *v514; // rbx
  __int64 v515; // rax
  __int64 v516; // rax
  __int64 v517; // rax
  __int64 v518; // rax
  __int64 v519; // rbx
  __int64 v520; // rax
  __int64 result; // rax
  __int64 v522; // rax
  __int64 v523; // rax
  __int64 v524; // rdx
  __int64 v525; // rax
  __int64 v526; // rax
  __int64 v527; // [rsp+20h] [rbp-39h]
  __int64 v528; // [rsp+20h] [rbp-39h]
  __int64 v529; // [rsp+20h] [rbp-39h]
  __int64 v530; // [rsp+20h] [rbp-39h]
  __int64 v531; // [rsp+20h] [rbp-39h]
  __int64 v532; // [rsp+20h] [rbp-39h]
  __int64 v533; // [rsp+20h] [rbp-39h]
  __int64 v534; // [rsp+20h] [rbp-39h]
  __int64 v535; // [rsp+20h] [rbp-39h]
  __int64 v536; // [rsp+20h] [rbp-39h]
  __int64 v537; // [rsp+20h] [rbp-39h]
  __int64 v538; // [rsp+20h] [rbp-39h]
  __int64 v539; // [rsp+20h] [rbp-39h]
  __int64 v540; // [rsp+20h] [rbp-39h]
  __int64 v541; // [rsp+20h] [rbp-39h]
  __int64 v542; // [rsp+20h] [rbp-39h]
  __int64 v543; // [rsp+20h] [rbp-39h]
  __int64 v544; // [rsp+20h] [rbp-39h]
  __int64 v545; // [rsp+20h] [rbp-39h]
  __int64 v546; // [rsp+20h] [rbp-39h]
  __int64 v547; // [rsp+20h] [rbp-39h]
  __int64 v548; // [rsp+20h] [rbp-39h]
  __int64 v549; // [rsp+20h] [rbp-39h]
  __int64 v550; // [rsp+20h] [rbp-39h]
  __int64 v551; // [rsp+20h] [rbp-39h]
  __int64 v552; // [rsp+20h] [rbp-39h]
  __int64 v553; // [rsp+20h] [rbp-39h]
  __int64 v554; // [rsp+20h] [rbp-39h]
  __int64 v555; // [rsp+20h] [rbp-39h]
  __int64 v556; // [rsp+20h] [rbp-39h]
  __int64 v557; // [rsp+20h] [rbp-39h]
  __int64 v558; // [rsp+20h] [rbp-39h]
  __int64 v559; // [rsp+20h] [rbp-39h]
  __int64 v560; // [rsp+20h] [rbp-39h]
  __int64 v561; // [rsp+20h] [rbp-39h]
  __int64 v562; // [rsp+20h] [rbp-39h]
  __int64 v563; // [rsp+20h] [rbp-39h]
  __int64 v564; // [rsp+20h] [rbp-39h]
  __int64 v565; // [rsp+20h] [rbp-39h]
  __int64 v566; // [rsp+28h] [rbp-31h]
  __int64 v567; // [rsp+28h] [rbp-31h]
  __int64 v568; // [rsp+28h] [rbp-31h]
  __int64 v569; // [rsp+28h] [rbp-31h]
  __int64 v570; // [rsp+28h] [rbp-31h]
  __int64 v571; // [rsp+28h] [rbp-31h]
  __int64 v572; // [rsp+28h] [rbp-31h]
  __int64 v573; // [rsp+28h] [rbp-31h]
  __int64 v574; // [rsp+28h] [rbp-31h]
  __int64 v575; // [rsp+28h] [rbp-31h]
  __int64 v576; // [rsp+28h] [rbp-31h]
  __int64 v577; // [rsp+28h] [rbp-31h]
  __int64 v578; // [rsp+28h] [rbp-31h]
  __int64 v579; // [rsp+28h] [rbp-31h]
  __int64 v580; // [rsp+28h] [rbp-31h]
  __int64 v581; // [rsp+28h] [rbp-31h]
  __int64 v582; // [rsp+28h] [rbp-31h]
  __int64 v583; // [rsp+28h] [rbp-31h]
  __int64 v584; // [rsp+28h] [rbp-31h]
  __int64 v585; // [rsp+28h] [rbp-31h]
  __int64 v586; // [rsp+28h] [rbp-31h]
  __int64 v587; // [rsp+28h] [rbp-31h]
  __int64 v588; // [rsp+28h] [rbp-31h]
  __int64 v589; // [rsp+28h] [rbp-31h]
  __int64 v590; // [rsp+28h] [rbp-31h]
  __int64 v591; // [rsp+28h] [rbp-31h]
  __int64 v592; // [rsp+28h] [rbp-31h]
  __int64 v593; // [rsp+28h] [rbp-31h]
  __int64 v594; // [rsp+28h] [rbp-31h]
  __int64 v595; // [rsp+28h] [rbp-31h]
  __int64 v596; // [rsp+28h] [rbp-31h]
  __int64 v597; // [rsp+28h] [rbp-31h]
  __int64 v598; // [rsp+28h] [rbp-31h]
  __int64 v599; // [rsp+28h] [rbp-31h]
  __int64 v600; // [rsp+28h] [rbp-31h]
  __int64 v601; // [rsp+28h] [rbp-31h]
  __int64 v602; // [rsp+28h] [rbp-31h]
  __int64 v603; // [rsp+28h] [rbp-31h]
  __int64 v604; // [rsp+28h] [rbp-31h]
  _QWORD v605[2]; // [rsp+30h] [rbp-29h] BYREF
  __int64 v606; // [rsp+40h] [rbp-19h]
  __int64 v607; // [rsp+48h] [rbp-11h] BYREF
  __int64 v608; // [rsp+50h] [rbp-9h]
  __int64 v609; // [rsp+58h] [rbp-1h]
  __int64 v610; // [rsp+60h] [rbp+7h]
  __int64 v611; // [rsp+68h] [rbp+Fh]

  v4 = *(_QWORD *)(a2 + 16);
  qword_1803CA780 = v4;
  qword_1803CA778 = a2;
  if ( byte_1803AE79F == 0 )
  {
    if ( byte_1803AE79E == 0 )
    {
// Original IDA line 618 | qword_1803B3F10 -> novels.pipeline.pipeline_runner#0 bytes [3963839, 3963846]: "state"
      sub_180388A30(a1, &qword_1803B3F10, "novels.pipeline.pipeline_runner");
      byte_1803AE79E = 1;
    }
    sub_1800E1350();
    v4 = qword_1803CA780;
    byte_1803AE79F = 1;
  }
  sub_1800DE740(v4, qword_1803AB330, qword_1803AB560);
  v6 = (__int64 *)sub_1800015A0(qword_1803CA780, qword_1803AB2A8);
  v7 = 0;
  if ( v6 != nullptr )
    v7 = *v6;
  Length = PyUnicode_GetLength(v7);
  v9 = PyUnicode_Find(v7, qword_1803AB4A8, 0, Length, -1);
  if ( v9 != -1 )
  {
    v10 = PyUnicode_Substring(v7, 0, v9);
    sub_1800DE7D0(qword_1803CA780, qword_1803AB2B0, v10);
  }
  v11 = (_QWORD *)sub_1800015A0(qword_1803CA780, qword_1803AB300);
  if ( v11 == nullptr || *v11 == 0 )
  {
    Dict = PyModule_GetDict(qword_1803AB780);
    sub_1800DE740(qword_1803CA780, qword_1803AB300, Dict);
  }
  v13 = sub_18038D790(a3);
  sub_1800DE740(qword_1803CA780, qword_1803AB328, v13);
  v14 = qword_1803C9FD0;
  if ( qword_1803C9FD0 == 0 )
  {
    v14 = PyImport_ImportModule("importlib._bootstrap");
    qword_1803C9FD0 = v14;
  }
  AttrString = (int *)PyObject_GetAttrString(v14, "_spec_from_module");
  v16 = sub_180376770(a1, AttrString, qword_1803CA778);
  if ( *AttrString >= 0 )
  {
    v17 = (*(_QWORD *)AttrString)-- == 1;
    if ( v17 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)AttrString + 1) + 48LL))(AttrString);
  }
  if ( v16 == 0 )
  {
    PyErr_PrintEx(0);
    abort();
  }
  sub_18037F3A0(a1, v16, qword_1803AB4D0, Py_TrueStruct);
  sub_1800DE7D0(qword_1803CA780, qword_1803AB4C8, v16);
  v605[0] = 0;
// Original IDA line 667 | qword_1803B54E8 -> novels.pipeline.pipeline_runner#699 bytes [3978213, 3978222]: "__doc__"
  sub_1800DE740(qword_1803CA780, qword_1803B54E8, Py_NoneStruct);
// Original IDA line 668 | qword_1803B54F0 -> novels.pipeline.pipeline_runner#700 bytes [3978222, 3978232]: "__file__"
  sub_1800DE740(qword_1803CA780, qword_1803B54F0, qword_1803B3F00);
  v606 = sub_18036CCF0(qword_1803B3BA8, qword_1803CA778);
  v18 = v606;
  sub_180001D00(a1, v606);
  v19 = qword_1803B3F00;
  v20 = sub_1800DFBA0(a1);
// Original IDA line 674 | qword_1803B54F8 -> novels.pipeline.pipeline_runner#701 bytes [3978232, 3978240]: "origin"
  if ( (unsigned __int8)sub_18037F3A0(a1, v20, qword_1803B54F8, v19) == 0
// Original IDA line 675 | qword_1803B5500 -> novels.pipeline.pipeline_runner#702 bytes [3978240, 3978254]: "has_location"
    || (v23 = sub_1800DFBA0(a1), (unsigned __int8)sub_18037F3A0(a1, v23, qword_1803B5500, Py_TrueStruct) == 0) )
  {
    v22 = 1;
    v605[0] = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_715;
  }
// Original IDA line 682 | qword_1803B5508 -> novels.pipeline.pipeline_runner#703 bytes [3978254, 3978266]: "__cached__"
  sub_1800DE740(qword_1803CA780, qword_1803B5508, Py_NoneStruct);
  v24 = sub_180372070();
// Original IDA line 684 | qword_1803B5510 -> novels.pipeline.pipeline_runner#704 bytes [3978266, 3978279]: "annotations"
  v25 = sub_18037F0C0(a1, v24, qword_1803B5510);
// Original IDA line 685 | qword_1803B5510 -> novels.pipeline.pipeline_runner#704 bytes [3978266, 3978279]: "annotations"
  sub_1800DE7D0(qword_1803CA780, qword_1803B5510, v25);
  v26 = qword_1803AB250;
// Original IDA line 687 | qword_1803B5518 -> novels.pipeline.pipeline_runner#705 bytes [3978279, 3978285]: "json"
  v607 = qword_1803B5518;
  v608 = qword_1803CA780;
  v27 = qword_1803AB578;
  *(_DWORD *)(v606 + 40) = 3;
  v609 = Py_NoneStruct;
  v610 = Py_NoneStruct;
  v611 = v26;
  if ( v27 == 0 )
  {
    v27 = sub_18035F440("__import__");
    qword_1803AB578 = v27;
  }
  v28 = sub_18037A640(a1, v27, &v607);
  if ( v28 == 0 )
  {
    v22 = 3;
    v605[0] = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_715;
  }
// Original IDA line 707 | qword_1803B5518 -> novels.pipeline.pipeline_runner#705 bytes [3978279, 3978285]: "json"
  sub_1800DE7D0(qword_1803CA780, qword_1803B5518, v28);
  v29 = qword_1803AB250;
// Original IDA line 709 | qword_1803B4EA0 -> novels.pipeline.pipeline_runner#498 bytes [3974200, 3974204]: "re"
  v607 = qword_1803B4EA0;
  v608 = qword_1803CA780;
  v30 = qword_1803AB578;
  *(_DWORD *)(v18 + 40) = 4;
  v609 = Py_NoneStruct;
  v610 = Py_NoneStruct;
  v611 = v29;
  if ( v30 == 0 )
  {
    v30 = sub_18035F440("__import__");
    qword_1803AB578 = v30;
  }
  v31 = sub_18037A640(a1, v30, &v607);
  if ( v31 == 0 )
  {
    v22 = 4;
    v605[0] = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_715;
  }
// Original IDA line 729 | qword_1803B4EA0 -> novels.pipeline.pipeline_runner#498 bytes [3974200, 3974204]: "re"
  sub_1800DE7D0(qword_1803CA780, qword_1803B4EA0, v31);
// Original IDA line 730 | qword_1803B5528 -> novels.pipeline.pipeline_runner#707 bytes [3978294, 3978302]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978296, 3978302], "value": "Path"}]}
  v32 = qword_1803B5528;
  v33 = qword_1803AB250;
// Original IDA line 732 | qword_1803B5520 -> novels.pipeline.pipeline_runner#706 bytes [3978285, 3978294]: "pathlib"
  v607 = qword_1803B5520;
  v608 = qword_1803CA780;
  v34 = qword_1803AB578;
  *(_DWORD *)(v18 + 40) = 5;
  v609 = Py_NoneStruct;
  v610 = v32;
  v611 = v33;
  if ( v34 == 0 )
  {
    v34 = sub_18035F440("__import__");
    qword_1803AB578 = v34;
  }
  v35 = sub_18037A640(a1, v34, &v607);
  v36 = (int *)v35;
  if ( v35 != 0 )
  {
    if ( (unsigned int)sub_18000CC50(v35, PyModule_Type) != 0 )
// Original IDA line 749 | qword_1803B5530 -> novels.pipeline.pipeline_runner#708 bytes [3978302, 3978308]: "Path"
      v37 = sub_180371D80(a1, (_DWORD)v36, qword_1803CA780, qword_1803B5530, qword_1803AB250);
    else
// Original IDA line 751 | qword_1803B5530 -> novels.pipeline.pipeline_runner#708 bytes [3978302, 3978308]: "Path"
      v37 = sub_180371890(a1, v36, qword_1803B5530);
    v38 = v37;
    if ( *v36 >= 0 )
    {
      v17 = (*(_QWORD *)v36)-- == 1;
      if ( v17 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
    }
    if ( v38 != 0 )
    {
// Original IDA line 761 | qword_1803B5530 -> novels.pipeline.pipeline_runner#708 bytes [3978302, 3978308]: "Path"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5530, v38);
      v39 = (int *)sub_180372110();
      v40 = v39;
      if ( *v39 != -1 )
        ++*v39;
      if ( (unsigned int)sub_18000CC50(v39, PyModule_Type) != 0 )
// Original IDA line 767 | qword_1803B5538 -> novels.pipeline.pipeline_runner#709 bytes [3978308, 3978313]: "Any"
        v41 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5538, qword_1803AB250);
      else
// Original IDA line 769 | qword_1803B5538 -> novels.pipeline.pipeline_runner#709 bytes [3978308, 3978313]: "Any"
        v41 = sub_180371890(a1, v40, qword_1803B5538);
      if ( v41 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
LABEL_68:
        v22 = 6;
        v605[0] = 0;
        if ( *v40 >= 0 )
        {
          v17 = (*(_QWORD *)v40)-- == 1;
          if ( v17 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v40 + 1) + 48LL))(v40);
        }
LABEL_714:
        v605[0] = v42;
LABEL_715:
        v519 = *(_QWORD *)(v605[0] + 40LL);
        if ( v519 != 0 )
        {
          if ( *(_QWORD *)(v519 + 24) == v18 )
          {
LABEL_721:
            sub_180001D80(a1, v21);
            sub_18000CE40(a1, v605);
            return 0;
          }
          v520 = sub_18038E2A0(v18, v22);
          *(_QWORD *)(v520 + 16) = v519;
          if ( *(_DWORD *)v519 != -1 )
            ++*(_DWORD *)v519;
        }
        else
        {
          v520 = sub_18038E2A0(v18, v22);
        }
        sub_180001300(v605, v520);
        goto LABEL_721;
      }
// Original IDA line 808 | qword_1803B5538 -> novels.pipeline.pipeline_runner#709 bytes [3978308, 3978313]: "Any"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5538, v41);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
// Original IDA line 810 | qword_1803B5540 -> novels.pipeline.pipeline_runner#710 bytes [3978313, 3978323]: "Callable"
        v43 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5540, qword_1803AB250);
      else
// Original IDA line 812 | qword_1803B5540 -> novels.pipeline.pipeline_runner#710 bytes [3978313, 3978323]: "Callable"
        v43 = sub_180371890(a1, v40, qword_1803B5540);
      if ( v43 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
// Original IDA line 819 | qword_1803B5540 -> novels.pipeline.pipeline_runner#710 bytes [3978313, 3978323]: "Callable"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5540, v43);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
// Original IDA line 821 | qword_1803B5548 -> novels.pipeline.pipeline_runner#711 bytes [3978323, 3978329]: "Dict"
        v44 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5548, qword_1803AB250);
      else
// Original IDA line 823 | qword_1803B5548 -> novels.pipeline.pipeline_runner#711 bytes [3978323, 3978329]: "Dict"
        v44 = sub_180371890(a1, v40, qword_1803B5548);
      if ( v44 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
// Original IDA line 830 | qword_1803B5548 -> novels.pipeline.pipeline_runner#711 bytes [3978323, 3978329]: "Dict"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5548, v44);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
// Original IDA line 832 | qword_1803B5550 -> novels.pipeline.pipeline_runner#712 bytes [3978329, 3978335]: "List"
        v45 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5550, qword_1803AB250);
      else
// Original IDA line 834 | qword_1803B5550 -> novels.pipeline.pipeline_runner#712 bytes [3978329, 3978335]: "List"
        v45 = sub_180371890(a1, v40, qword_1803B5550);
      if ( v45 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
// Original IDA line 841 | qword_1803B5550 -> novels.pipeline.pipeline_runner#712 bytes [3978329, 3978335]: "List"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5550, v45);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
// Original IDA line 843 | qword_1803B5558 -> novels.pipeline.pipeline_runner#713 bytes [3978335, 3978345]: "Optional"
        v46 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5558, qword_1803AB250);
      else
// Original IDA line 845 | qword_1803B5558 -> novels.pipeline.pipeline_runner#713 bytes [3978335, 3978345]: "Optional"
        v46 = sub_180371890(a1, v40, qword_1803B5558);
      if ( v46 == 0 )
      {
        sub_18000CE30(a1, v605);
        v42 = v605[0];
        goto LABEL_68;
      }
// Original IDA line 852 | qword_1803B5558 -> novels.pipeline.pipeline_runner#713 bytes [3978335, 3978345]: "Optional"
      sub_1800DE7D0(qword_1803CA780, qword_1803B5558, v46);
      if ( *v40 >= 0 )
      {
        v17 = (*(_QWORD *)v40)-- == 1;
        if ( v17 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v40 + 1) + 48LL))(v40);
      }
// Original IDA line 859 | qword_1803B5560 -> novels.pipeline.pipeline_runner#714 bytes [3978345, 3978369]: "agents.architect_agent"
      v47 = qword_1803B5560;
      v48 = qword_1803CA780;
// Original IDA line 861 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
      v566 = qword_1803B4258;
// Original IDA line 862 | qword_1803B5568 -> novels.pipeline.pipeline_runner#715 bytes [3978369, 3978387]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978371, 3978387], "value": "ArchitectAgent"}]}
      v527 = qword_1803B5568;
      *(_DWORD *)(v18 + 40) = 8;
      v49 = sub_1803717E0(a1, v47, v48, Py_NoneStruct, v527, v566);
      v50 = (int *)v49;
      if ( v49 != 0 )
      {
        if ( (unsigned int)sub_18000CC50(v49, PyModule_Type) != 0 )
// Original IDA line 869 | qword_1803B40D8 -> novels.pipeline.pipeline_runner#57 bytes [3964865, 3964881]: "ArchitectAgent"
          v51 = sub_180371D80(a1, (_DWORD)v50, qword_1803CA780, qword_1803B40D8, qword_1803AB250);
        else
// Original IDA line 871 | qword_1803B40D8 -> novels.pipeline.pipeline_runner#57 bytes [3964865, 3964881]: "ArchitectAgent"
          v51 = sub_180371890(a1, v50, qword_1803B40D8);
        v52 = v51;
        if ( *v50 >= 0 )
        {
          v17 = (*(_QWORD *)v50)-- == 1;
          if ( v17 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v50 + 1) + 48LL))(v50);
        }
        if ( v52 != 0 )
        {
// Original IDA line 881 | qword_1803B40D8 -> novels.pipeline.pipeline_runner#57 bytes [3964865, 3964881]: "ArchitectAgent"
          sub_1800DE7D0(qword_1803CA780, qword_1803B40D8, v52);
// Original IDA line 882 | qword_1803B5570 -> novels.pipeline.pipeline_runner#716 bytes [3978387, 3978421]: "agents.foundation_reviewer_agent"
          v53 = qword_1803B5570;
          v54 = qword_1803CA780;
// Original IDA line 884 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
          v567 = qword_1803B4258;
// Original IDA line 885 | qword_1803B5578 -> novels.pipeline.pipeline_runner#717 bytes [3978421, 3978448]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978423, 3978448], "value": "FoundationReviewerAgent"}]}
          v528 = qword_1803B5578;
          *(_DWORD *)(v18 + 40) = 9;
          v55 = sub_1803717E0(a1, v53, v54, Py_NoneStruct, v528, v567);
          v56 = (int *)v55;
          if ( v55 != 0 )
          {
            if ( (unsigned int)sub_18000CC50(v55, PyModule_Type) != 0 )
// Original IDA line 892 | qword_1803B40E8 -> novels.pipeline.pipeline_runner#59 bytes [3964901, 3964926]: "FoundationReviewerAgent"
              v57 = sub_180371D80(a1, (_DWORD)v56, qword_1803CA780, qword_1803B40E8, qword_1803AB250);
            else
// Original IDA line 894 | qword_1803B40E8 -> novels.pipeline.pipeline_runner#59 bytes [3964901, 3964926]: "FoundationReviewerAgent"
              v57 = sub_180371890(a1, v56, qword_1803B40E8);
            v58 = v57;
            if ( *v56 >= 0 )
            {
              v17 = (*(_QWORD *)v56)-- == 1;
              if ( v17 )
              {
                v59 = sub_18000CC40(v56);
                (*(void (__fastcall **)(int *))(v59 + 48))(v56);
              }
            }
            if ( v58 != 0 )
            {
// Original IDA line 907 | qword_1803B40E8 -> novels.pipeline.pipeline_runner#59 bytes [3964901, 3964926]: "FoundationReviewerAgent"
              sub_1800DE7D0(qword_1803CA780, qword_1803B40E8, v58);
// Original IDA line 908 | qword_1803B5580 -> novels.pipeline.pipeline_runner#718 bytes [3978448, 3978470]: "agents.auditor_agent"
              v60 = qword_1803B5580;
              v61 = qword_1803CA780;
// Original IDA line 910 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
              v568 = qword_1803B4258;
// Original IDA line 911 | qword_1803B5588 -> novels.pipeline.pipeline_runner#719 bytes [3978470, 3978486]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978472, 3978486], "value": "AuditorAgent"}]}
              v529 = qword_1803B5588;
              *(_DWORD *)(v18 + 40) = 10;
              v62 = sub_1803717E0(a1, v60, v61, Py_NoneStruct, v529, v568);
              v63 = (int *)v62;
              if ( v62 != 0 )
              {
                if ( (unsigned int)sub_18000CC50(v62, PyModule_Type) != 0 )
// Original IDA line 918 | qword_1803B43F8 -> novels.pipeline.pipeline_runner#157 bytes [3966777, 3966791]: "AuditorAgent"
                  v64 = sub_180371D80(a1, (_DWORD)v63, qword_1803CA780, qword_1803B43F8, qword_1803AB250);
                else
// Original IDA line 920 | qword_1803B43F8 -> novels.pipeline.pipeline_runner#157 bytes [3966777, 3966791]: "AuditorAgent"
                  v64 = sub_180371890(a1, v63, qword_1803B43F8);
                v65 = v64;
                if ( *v63 >= 0 )
                {
                  v17 = (*(_QWORD *)v63)-- == 1;
                  if ( v17 )
                  {
                    v66 = sub_18000CC40(v63);
                    (*(void (__fastcall **)(int *))(v66 + 48))(v63);
                  }
                }
                if ( v65 != 0 )
                {
// Original IDA line 933 | qword_1803B43F8 -> novels.pipeline.pipeline_runner#157 bytes [3966777, 3966791]: "AuditorAgent"
                  sub_1800DE7D0(qword_1803CA780, qword_1803B43F8, v65);
// Original IDA line 934 | qword_1803B5590 -> novels.pipeline.pipeline_runner#720 bytes [3978486, 3978505]: "agents.base_agent"
                  v67 = qword_1803B5590;
                  v68 = qword_1803CA780;
// Original IDA line 936 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                  v569 = qword_1803B4258;
// Original IDA line 937 | qword_1803B5598 -> novels.pipeline.pipeline_runner#721 bytes [3978505, 3978521]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978507, 3978521], "value": "AgentContext"}]}
                  v530 = qword_1803B5598;
                  *(_DWORD *)(v18 + 40) = 11;
                  v69 = sub_1803717E0(a1, v67, v68, Py_NoneStruct, v530, v569);
                  v70 = (int *)v69;
                  if ( v69 != 0 )
                  {
                    if ( (unsigned int)sub_18000CC50(v69, PyModule_Type) != 0 )
// Original IDA line 944 | qword_1803B3F78 -> novels.pipeline.pipeline_runner#13 bytes [3964000, 3964014]: "AgentContext"
                      v71 = sub_180371D80(a1, (_DWORD)v70, qword_1803CA780, qword_1803B3F78, qword_1803AB250);
                    else
// Original IDA line 946 | qword_1803B3F78 -> novels.pipeline.pipeline_runner#13 bytes [3964000, 3964014]: "AgentContext"
                      v71 = sub_180371890(a1, v70, qword_1803B3F78);
                    v72 = v71;
                    if ( *v70 >= 0 )
                    {
                      v17 = (*(_QWORD *)v70)-- == 1;
                      if ( v17 )
                      {
                        v73 = sub_18000CC40(v70);
                        (*(void (__fastcall **)(int *))(v73 + 48))(v70);
                      }
                    }
                    if ( v72 != 0 )
                    {
// Original IDA line 959 | qword_1803B3F78 -> novels.pipeline.pipeline_runner#13 bytes [3964000, 3964014]: "AgentContext"
                      sub_1800DE7D0(qword_1803CA780, qword_1803B3F78, v72);
// Original IDA line 960 | qword_1803B55A0 -> novels.pipeline.pipeline_runner#722 bytes [3978521, 3978544]: "agents.composer_agent"
                      v74 = qword_1803B55A0;
                      v75 = qword_1803CA780;
// Original IDA line 962 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                      v570 = qword_1803B4258;
// Original IDA line 963 | qword_1803B55A8 -> novels.pipeline.pipeline_runner#723 bytes [3978544, 3978561]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978546, 3978561], "value": "ComposerAgent"}]}
                      v531 = qword_1803B55A8;
                      *(_DWORD *)(v18 + 40) = 12;
                      v76 = sub_1803717E0(a1, v74, v75, Py_NoneStruct, v531, v570);
                      v77 = (int *)v76;
                      if ( v76 != 0 )
                      {
                        if ( (unsigned int)sub_18000CC50(v76, PyModule_Type) != 0 )
// Original IDA line 970 | qword_1803B42E8 -> novels.pipeline.pipeline_runner#123 bytes [3966210, 3966225]: "ComposerAgent"
                          v78 = sub_180371D80(a1, (_DWORD)v77, qword_1803CA780, qword_1803B42E8, qword_1803AB250);
                        else
// Original IDA line 972 | qword_1803B42E8 -> novels.pipeline.pipeline_runner#123 bytes [3966210, 3966225]: "ComposerAgent"
                          v78 = sub_180371890(a1, v77, qword_1803B42E8);
                        v79 = v78;
                        if ( *v77 >= 0 )
                        {
                          v17 = (*(_QWORD *)v77)-- == 1;
                          if ( v17 )
                          {
                            v80 = sub_18000CC40(v77);
                            (*(void (__fastcall **)(int *))(v80 + 48))(v77);
                          }
                        }
                        if ( v79 != 0 )
                        {
// Original IDA line 985 | qword_1803B42E8 -> novels.pipeline.pipeline_runner#123 bytes [3966210, 3966225]: "ComposerAgent"
                          sub_1800DE7D0(qword_1803CA780, qword_1803B42E8, v79);
// Original IDA line 986 | qword_1803B55B0 -> novels.pipeline.pipeline_runner#724 bytes [3978561, 3978588]: "agents.consolidator_agent"
                          v81 = qword_1803B55B0;
                          v82 = qword_1803CA780;
// Original IDA line 988 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                          v571 = qword_1803B4258;
// Original IDA line 989 | qword_1803B55B8 -> novels.pipeline.pipeline_runner#725 bytes [3978588, 3978609]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978590, 3978609], "value": "ConsolidatorAgent"}]}
                          v532 = qword_1803B55B8;
                          *(_DWORD *)(v18 + 40) = 13;
                          v83 = sub_1803717E0(a1, v81, v82, Py_NoneStruct, v532, v571);
                          v84 = (int *)v83;
                          if ( v83 != 0 )
                          {
                            if ( (unsigned int)sub_18000CC50(v83, PyModule_Type) != 0 )
// Original IDA line 996 | qword_1803B4D80 -> novels.pipeline.pipeline_runner#462 bytes [3973719, 3973738]: "ConsolidatorAgent"
                              v85 = sub_180371D80(a1, (_DWORD)v84, qword_1803CA780, qword_1803B4D80, qword_1803AB250);
                            else
// Original IDA line 998 | qword_1803B4D80 -> novels.pipeline.pipeline_runner#462 bytes [3973719, 3973738]: "ConsolidatorAgent"
                              v85 = sub_180371890(a1, v84, qword_1803B4D80);
                            v86 = v85;
                            if ( *v84 >= 0 )
                            {
                              v17 = (*(_QWORD *)v84)-- == 1;
                              if ( v17 )
                              {
                                v87 = sub_18000CC40(v84);
                                (*(void (__fastcall **)(int *))(v87 + 48))(v84);
                              }
                            }
                            if ( v86 != 0 )
                            {
// Original IDA line 1011 | qword_1803B4D80 -> novels.pipeline.pipeline_runner#462 bytes [3973719, 3973738]: "ConsolidatorAgent"
                              sub_1800DE7D0(qword_1803CA780, qword_1803B4D80, v86);
// Original IDA line 1012 | qword_1803B55C0 -> novels.pipeline.pipeline_runner#726 bytes [3978609, 3978631]: "agents.planner_agent"
                              v88 = qword_1803B55C0;
                              v89 = qword_1803CA780;
// Original IDA line 1014 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                              v572 = qword_1803B4258;
// Original IDA line 1015 | qword_1803B55C8 -> novels.pipeline.pipeline_runner#727 bytes [3978631, 3978647]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978633, 3978647], "value": "PlannerAgent"}]}
                              v533 = qword_1803B55C8;
                              *(_DWORD *)(v18 + 40) = 14;
                              v90 = sub_1803717E0(a1, v88, v89, Py_NoneStruct, v533, v572);
                              v91 = (int *)v90;
                              if ( v90 != 0 )
                              {
                                if ( (unsigned int)sub_18000CC50(v90, PyModule_Type) != 0 )
                                  v92 = sub_180371D80(
                                          a1,
                                          (_DWORD)v91,
                                          qword_1803CA780,
// Original IDA line 1026 | qword_1803B42B8 -> novels.pipeline.pipeline_runner#117 bytes [3966098, 3966112]: "PlannerAgent"
                                          qword_1803B42B8,
                                          qword_1803AB250);
                                else
// Original IDA line 1029 | qword_1803B42B8 -> novels.pipeline.pipeline_runner#117 bytes [3966098, 3966112]: "PlannerAgent"
                                  v92 = sub_180371890(a1, v91, qword_1803B42B8);
                                v93 = v92;
                                if ( *v91 >= 0 )
                                {
                                  v17 = (*(_QWORD *)v91)-- == 1;
                                  if ( v17 )
                                  {
                                    v94 = sub_18000CC40(v91);
                                    (*(void (__fastcall **)(int *))(v94 + 48))(v91);
                                  }
                                }
                                if ( v93 != 0 )
                                {
// Original IDA line 1042 | qword_1803B42B8 -> novels.pipeline.pipeline_runner#117 bytes [3966098, 3966112]: "PlannerAgent"
                                  sub_1800DE7D0(qword_1803CA780, qword_1803B42B8, v93);
// Original IDA line 1043 | qword_1803B55D0 -> novels.pipeline.pipeline_runner#728 bytes [3978647, 3978669]: "agents.reviser_agent"
                                  v95 = qword_1803B55D0;
                                  v96 = qword_1803CA780;
// Original IDA line 1045 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                  v573 = qword_1803B4258;
// Original IDA line 1046 | qword_1803B55D8 -> novels.pipeline.pipeline_runner#729 bytes [3978669, 3978685]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978671, 3978685], "value": "ReviserAgent"}]}
                                  v534 = qword_1803B55D8;
                                  *(_DWORD *)(v18 + 40) = 15;
                                  v97 = sub_1803717E0(a1, v95, v96, Py_NoneStruct, v534, v573);
                                  v98 = (int *)v97;
                                  if ( v97 != 0 )
                                  {
                                    if ( (unsigned int)sub_18000CC50(v97, PyModule_Type) != 0 )
                                      v99 = sub_180371D80(
                                              a1,
                                              (_DWORD)v98,
                                              qword_1803CA780,
// Original IDA line 1057 | qword_1803B4408 -> novels.pipeline.pipeline_runner#159 bytes [3966809, 3966823]: "ReviserAgent"
                                              qword_1803B4408,
                                              qword_1803AB250);
                                    else
// Original IDA line 1060 | qword_1803B4408 -> novels.pipeline.pipeline_runner#159 bytes [3966809, 3966823]: "ReviserAgent"
                                      v99 = sub_180371890(a1, v98, qword_1803B4408);
                                    v100 = v99;
                                    if ( *v98 >= 0 )
                                    {
                                      v17 = (*(_QWORD *)v98)-- == 1;
                                      if ( v17 )
                                      {
                                        v101 = sub_18000CC40(v98);
                                        (*(void (__fastcall **)(int *))(v101 + 48))(v98);
                                      }
                                    }
                                    if ( v100 != 0 )
                                    {
// Original IDA line 1073 | qword_1803B4408 -> novels.pipeline.pipeline_runner#159 bytes [3966809, 3966823]: "ReviserAgent"
                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4408, v100);
// Original IDA line 1074 | qword_1803B55E0 -> novels.pipeline.pipeline_runner#730 bytes [3978685, 3978707]: "agents.settler_agent"
                                      v102 = qword_1803B55E0;
                                      v103 = qword_1803CA780;
// Original IDA line 1076 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                      v574 = qword_1803B4258;
// Original IDA line 1077 | qword_1803B55E8 -> novels.pipeline.pipeline_runner#731 bytes [3978707, 3978723]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978709, 3978723], "value": "SettlerAgent"}]}
                                      v535 = qword_1803B55E8;
                                      *(_DWORD *)(v18 + 40) = 16;
                                      v104 = sub_1803717E0(a1, v102, v103, Py_NoneStruct, v535, v574);
                                      v105 = (int *)v104;
                                      if ( v104 != 0 )
                                      {
                                        if ( (unsigned int)sub_18000CC50(v104, PyModule_Type) != 0 )
                                          v106 = sub_180371D80(
                                                   a1,
                                                   (_DWORD)v105,
                                                   qword_1803CA780,
// Original IDA line 1088 | qword_1803B4B88 -> novels.pipeline.pipeline_runner#399 bytes [3972325, 3972339]: "SettlerAgent"
                                                   qword_1803B4B88,
                                                   qword_1803AB250);
                                        else
// Original IDA line 1091 | qword_1803B4B88 -> novels.pipeline.pipeline_runner#399 bytes [3972325, 3972339]: "SettlerAgent"
                                          v106 = sub_180371890(a1, v105, qword_1803B4B88);
                                        v107 = v106;
                                        if ( *v105 >= 0 )
                                        {
                                          v17 = (*(_QWORD *)v105)-- == 1;
                                          if ( v17 )
                                          {
                                            v108 = sub_18000CC40(v105);
                                            (*(void (__fastcall **)(int *))(v108 + 48))(v105);
                                          }
                                        }
                                        if ( v107 != 0 )
                                        {
// Original IDA line 1104 | qword_1803B4B88 -> novels.pipeline.pipeline_runner#399 bytes [3972325, 3972339]: "SettlerAgent"
                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4B88, v107);
// Original IDA line 1105 | qword_1803B55F0 -> novels.pipeline.pipeline_runner#732 bytes [3978723, 3978753]: "agents.state_validator_agent"
                                          v109 = qword_1803B55F0;
                                          v110 = qword_1803CA780;
// Original IDA line 1107 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                          v575 = qword_1803B4258;
// Original IDA line 1108 | qword_1803B55F8 -> novels.pipeline.pipeline_runner#733 bytes [3978753, 3978776]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978755, 3978776], "value": "StateValidatorAgent"}]}
                                          v536 = qword_1803B55F8;
                                          *(_DWORD *)(v18 + 40) = 17;
                                          v111 = sub_1803717E0(a1, v109, v110, Py_NoneStruct, v536, v575);
                                          v112 = (int *)v111;
                                          if ( v111 != 0 )
                                          {
                                            if ( (unsigned int)sub_18000CC50(v111, PyModule_Type) != 0 )
                                              v113 = sub_180371D80(
                                                       a1,
                                                       (_DWORD)v112,
                                                       qword_1803CA780,
// Original IDA line 1119 | qword_1803B4BA0 -> novels.pipeline.pipeline_runner#402 bytes [3972365, 3972386]: "StateValidatorAgent"
                                                       qword_1803B4BA0,
                                                       qword_1803AB250);
                                            else
// Original IDA line 1122 | qword_1803B4BA0 -> novels.pipeline.pipeline_runner#402 bytes [3972365, 3972386]: "StateValidatorAgent"
                                              v113 = sub_180371890(a1, v112, qword_1803B4BA0);
                                            v114 = v113;
                                            if ( *v112 >= 0 )
                                            {
                                              v17 = (*(_QWORD *)v112)-- == 1;
                                              if ( v17 )
                                              {
                                                v115 = sub_18000CC40(v112);
                                                (*(void (__fastcall **)(int *))(v115 + 48))(v112);
                                              }
                                            }
                                            if ( v114 != 0 )
                                            {
// Original IDA line 1135 | qword_1803B4BA0 -> novels.pipeline.pipeline_runner#402 bytes [3972365, 3972386]: "StateValidatorAgent"
                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4BA0, v114);
// Original IDA line 1136 | qword_1803B5600 -> novels.pipeline.pipeline_runner#734 bytes [3978776, 3978797]: "agents.writer_agent"
                                              v116 = qword_1803B5600;
                                              v117 = qword_1803CA780;
// Original IDA line 1138 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                              v576 = qword_1803B4258;
// Original IDA line 1139 | qword_1803B5608 -> novels.pipeline.pipeline_runner#735 bytes [3978797, 3978812]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978799, 3978812], "value": "WriterAgent"}]}
                                              v537 = qword_1803B5608;
                                              *(_DWORD *)(v18 + 40) = 18;
                                              v118 = sub_1803717E0(a1, v116, v117, Py_NoneStruct, v537, v576);
                                              v119 = (int *)v118;
                                              if ( v118 != 0 )
                                              {
                                                if ( (unsigned int)sub_18000CC50(v118, PyModule_Type) != 0 )
                                                  v120 = sub_180371D80(
                                                           a1,
                                                           (_DWORD)v119,
                                                           qword_1803CA780,
// Original IDA line 1150 | qword_1803B4318 -> novels.pipeline.pipeline_runner#129 bytes [3966361, 3966374]: "WriterAgent"
                                                           qword_1803B4318,
                                                           qword_1803AB250);
                                                else
// Original IDA line 1153 | qword_1803B4318 -> novels.pipeline.pipeline_runner#129 bytes [3966361, 3966374]: "WriterAgent"
                                                  v120 = sub_180371890(a1, v119, qword_1803B4318);
                                                v121 = v120;
                                                if ( *v119 >= 0 )
                                                {
                                                  v17 = (*(_QWORD *)v119)-- == 1;
                                                  if ( v17 )
                                                  {
                                                    v122 = sub_18000CC40(v119);
                                                    (*(void (__fastcall **)(int *))(v122 + 48))(v119);
                                                  }
                                                }
                                                if ( v121 != 0 )
                                                {
// Original IDA line 1166 | qword_1803B4318 -> novels.pipeline.pipeline_runner#129 bytes [3966361, 3966374]: "WriterAgent"
                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4318, v121);
// Original IDA line 1167 | qword_1803B5610 -> novels.pipeline.pipeline_runner#736 bytes [3978812, 3978832]: "models.write_input"
                                                  v123 = qword_1803B5610;
                                                  v124 = qword_1803CA780;
// Original IDA line 1169 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                  v577 = qword_1803B4258;
// Original IDA line 1170 | qword_1803B5618 -> novels.pipeline.pipeline_runner#737 bytes [3978832, 3978853]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978834, 3978853], "value": "WriteChapterInput"}]}
                                                  v538 = qword_1803B5618;
                                                  *(_DWORD *)(v18 + 40) = 19;
                                                  v125 = sub_1803717E0(a1, v123, v124, Py_NoneStruct, v538, v577);
                                                  v126 = (int *)v125;
                                                  if ( v125 != 0 )
                                                  {
                                                    if ( (unsigned int)sub_18000CC50(v125, PyModule_Type) != 0 )
                                                      v127 = sub_180371D80(
                                                               a1,
                                                               (_DWORD)v126,
                                                               qword_1803CA780,
// Original IDA line 1181 | qword_1803B4308 -> novels.pipeline.pipeline_runner#127 bytes [3966279, 3966298]: "WriteChapterInput"
                                                               qword_1803B4308,
                                                               qword_1803AB250);
                                                    else
// Original IDA line 1184 | qword_1803B4308 -> novels.pipeline.pipeline_runner#127 bytes [3966279, 3966298]: "WriteChapterInput"
                                                      v127 = sub_180371890(a1, v126, qword_1803B4308);
                                                    v128 = v127;
                                                    if ( *v126 >= 0 )
                                                    {
                                                      v17 = (*(_QWORD *)v126)-- == 1;
                                                      if ( v17 )
                                                      {
                                                        v129 = sub_18000CC40(v126);
                                                        (*(void (__fastcall **)(int *))(v129 + 48))(v126);
                                                      }
                                                    }
                                                    if ( v128 != 0 )
                                                    {
// Original IDA line 1197 | qword_1803B4308 -> novels.pipeline.pipeline_runner#127 bytes [3966279, 3966298]: "WriteChapterInput"
                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4308, v128);
// Original IDA line 1198 | qword_1803B5620 -> novels.pipeline.pipeline_runner#738 bytes [3978853, 3978881]: "services.foundation_review"
                                                      v130 = qword_1803B5620;
                                                      v131 = qword_1803CA780;
// Original IDA line 1200 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                      v578 = qword_1803B4258;
// Original IDA line 1201 | qword_1803B5628 -> novels.pipeline.pipeline_runner#739 bytes [3978881, 3978915]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978883, 3978915], "value": "generate_and_review_foundation"}]}
                                                      v539 = qword_1803B5628;
                                                      *(_DWORD *)(v18 + 40) = 20;
                                                      v132 = sub_1803717E0(a1, v130, v131, Py_NoneStruct, v539, v578);
                                                      v133 = (int *)v132;
                                                      if ( v132 != 0 )
                                                      {
                                                        if ( (unsigned int)sub_18000CC50(v132, PyModule_Type) != 0 )
                                                          v134 = sub_180371D80(
                                                                   a1,
                                                                   (_DWORD)v133,
                                                                   qword_1803CA780,
// Original IDA line 1212 | qword_1803B4160 -> novels.pipeline.pipeline_runner#74 bytes [3965231, 3965263]: "generate_and_review_foundation"
                                                                   qword_1803B4160,
                                                                   qword_1803AB250);
                                                        else
// Original IDA line 1215 | qword_1803B4160 -> novels.pipeline.pipeline_runner#74 bytes [3965231, 3965263]: "generate_and_review_foundation"
                                                          v134 = sub_180371890(a1, v133, qword_1803B4160);
                                                        v135 = v134;
                                                        if ( *v133 >= 0 )
                                                        {
                                                          v17 = (*(_QWORD *)v133)-- == 1;
                                                          if ( v17 )
                                                          {
                                                            v136 = sub_18000CC40(v133);
                                                            (*(void (__fastcall **)(int *))(v136 + 48))(v133);
                                                          }
                                                        }
                                                        if ( v135 != 0 )
                                                        {
// Original IDA line 1228 | qword_1803B4160 -> novels.pipeline.pipeline_runner#74 bytes [3965231, 3965263]: "generate_and_review_foundation"
                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4160, v135);
// Original IDA line 1229 | qword_1803B5630 -> novels.pipeline.pipeline_runner#740 bytes [3978915, 3978938]: "services.memory_index"
                                                          v137 = qword_1803B5630;
                                                          v138 = qword_1803CA780;
// Original IDA line 1231 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                          v579 = qword_1803B4258;
// Original IDA line 1232 | qword_1803B5638 -> novels.pipeline.pipeline_runner#741 bytes [3978938, 3979001]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3978940, 3978960], "value": "clear_memory_index"}, {"tag": "a", "fileRange": [3978960, 3978979], "value": "load_memory_index"}, {"tag": "a", "fileRange": [3978979, 3979001], "value": "u... (truncated literal)
                                                          v540 = qword_1803B5638;
                                                          *(_DWORD *)(v18 + 40) = 21;
                                                          v139 = sub_1803717E0(
                                                                   a1,
                                                                   v137,
                                                                   v138,
                                                                   Py_NoneStruct,
                                                                   v540,
                                                                   v579);
                                                          v140 = (int *)v139;
                                                          if ( v139 == 0 )
                                                          {
                                                            sub_18000CE30(a1, v605);
                                                            v22 = 21;
                                                            goto LABEL_715;
                                                          }
                                                          if ( (unsigned int)sub_18000CC50(v139, PyModule_Type) != 0 )
                                                            v141 = sub_180371D80(
                                                                     a1,
                                                                     (_DWORD)v140,
                                                                     qword_1803CA780,
// Original IDA line 1253 | qword_1803B4D10 -> novels.pipeline.pipeline_runner#448 bytes [3973436, 3973456]: "clear_memory_index"
                                                                     qword_1803B4D10,
                                                                     qword_1803AB250);
                                                          else
// Original IDA line 1256 | qword_1803B4D10 -> novels.pipeline.pipeline_runner#448 bytes [3973436, 3973456]: "clear_memory_index"
                                                            v141 = sub_180371890(a1, v140, qword_1803B4D10);
                                                          if ( v141 == 0
// Original IDA line 1258 | qword_1803B4D10 -> novels.pipeline.pipeline_runner#448 bytes [3973436, 3973456]: "clear_memory_index"
                                                            || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4D10, v141),
                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1260 | qword_1803B4A90 -> novels.pipeline.pipeline_runner#368 bytes [3971530, 3971549]: "load_memory_index"
                                                              ? (v142 = sub_180371890(a1, v140, qword_1803B4A90))
                                                              : (v142 = sub_180371D80(
                                                                          a1,
                                                                          (_DWORD)v140,
                                                                          qword_1803CA780,
// Original IDA line 1265 | qword_1803B4A90 -> novels.pipeline.pipeline_runner#368 bytes [3971530, 3971549]: "load_memory_index"
                                                                          qword_1803B4A90,
                                                                          qword_1803AB250)),
                                                                v142 == 0
// Original IDA line 1268 | qword_1803B4A90 -> novels.pipeline.pipeline_runner#368 bytes [3971530, 3971549]: "load_memory_index"
                                                             || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4A90, v142),
                                                                  (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1270 | qword_1803B4A30 -> novels.pipeline.pipeline_runner#356 bytes [3971252, 3971274]: "upsert_chapter_entry"
                                                               ? (v143 = sub_180371890(a1, v140, qword_1803B4A30))
                                                               : (v143 = sub_180371D80(
                                                                           a1,
                                                                           (_DWORD)v140,
                                                                           qword_1803CA780,
// Original IDA line 1275 | qword_1803B4A30 -> novels.pipeline.pipeline_runner#356 bytes [3971252, 3971274]: "upsert_chapter_entry"
                                                                           qword_1803B4A30,
                                                                           qword_1803AB250)),
                                                                 v143 == 0)) )
                                                          {
                                                            v22 = 21;
LABEL_206:
                                                            sub_18000CE30(a1, v605);
LABEL_207:
                                                            v42 = v605[0];
                                                            sub_18000CE70(v605);
                                                            if ( *v140 >= 0 )
                                                            {
                                                              v17 = (*(_QWORD *)v140)-- == 1;
                                                              if ( v17 )
                                                              {
                                                                v144 = sub_18000CC40(v140);
                                                                (*(void (__fastcall **)(int *))(v144 + 48))(v140);
                                                              }
                                                            }
                                                            goto LABEL_714;
                                                          }
// Original IDA line 1296 | qword_1803B4A30 -> novels.pipeline.pipeline_runner#356 bytes [3971252, 3971274]: "upsert_chapter_entry"
                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4A30, v143);
                                                          if ( *v140 >= 0 )
                                                          {
                                                            v17 = (*(_QWORD *)v140)-- == 1;
                                                            if ( v17 )
                                                            {
                                                              v145 = sub_18000CC40(v140);
                                                              (*(void (__fastcall **)(int *))(v145 + 48))(v140);
                                                            }
                                                          }
// Original IDA line 1306 | qword_1803B5640 -> novels.pipeline.pipeline_runner#742 bytes [3979001, 3979026]: "services.chapter_status"
                                                          v146 = qword_1803B5640;
                                                          v147 = qword_1803CA780;
// Original IDA line 1308 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                          v580 = qword_1803B4258;
// Original IDA line 1309 | qword_1803B5648 -> novels.pipeline.pipeline_runner#743 bytes [3979026, 3979130]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979028, 3979049], "value": "build_audit_summary"}, {"tag": "a", "fileRange": [3979049, 3979078], "value": "build_chapter_write_summary"}, {"tag": "a", "fileRange": [3979078, 3979100], ... (truncated literal)
                                                          v541 = qword_1803B5648;
                                                          *(_DWORD *)(v18 + 40) = 22;
                                                          v148 = sub_1803717E0(
                                                                   a1,
                                                                   v146,
                                                                   v147,
                                                                   Py_NoneStruct,
                                                                   v541,
                                                                   v580);
                                                          v140 = (int *)v148;
                                                          if ( v148 == 0 )
                                                          {
                                                            sub_18000CE30(a1, v605);
                                                            v22 = 22;
                                                            goto LABEL_715;
                                                          }
                                                          if ( (unsigned int)sub_18000CC50(v148, PyModule_Type) != 0 )
                                                            v149 = sub_180371D80(
                                                                     a1,
                                                                     (_DWORD)v140,
                                                                     qword_1803CA780,
// Original IDA line 1330 | qword_1803B4AD0 -> novels.pipeline.pipeline_runner#376 bytes [3971673, 3971694]: "build_audit_summary"
                                                                     qword_1803B4AD0,
                                                                     qword_1803AB250);
                                                          else
// Original IDA line 1333 | qword_1803B4AD0 -> novels.pipeline.pipeline_runner#376 bytes [3971673, 3971694]: "build_audit_summary"
                                                            v149 = sub_180371890(a1, v140, qword_1803B4AD0);
                                                          if ( v149 == 0
// Original IDA line 1335 | qword_1803B4AD0 -> novels.pipeline.pipeline_runner#376 bytes [3971673, 3971694]: "build_audit_summary"
                                                            || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4AD0, v149),
                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1337 | qword_1803B4650 -> novels.pipeline.pipeline_runner#232 bytes [3968471, 3968500]: "build_chapter_write_summary"
                                                              ? (v150 = sub_180371890(a1, v140, qword_1803B4650))
                                                              : (v150 = sub_180371D80(
                                                                          a1,
                                                                          (_DWORD)v140,
                                                                          qword_1803CA780,
// Original IDA line 1342 | qword_1803B4650 -> novels.pipeline.pipeline_runner#232 bytes [3968471, 3968500]: "build_chapter_write_summary"
                                                                          qword_1803B4650,
                                                                          qword_1803AB250)),
                                                                v150 == 0
// Original IDA line 1345 | qword_1803B4650 -> novels.pipeline.pipeline_runner#232 bytes [3968471, 3968500]: "build_chapter_write_summary"
                                                             || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4650, v150),
                                                                  (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1347 | qword_1803B4648 -> novels.pipeline.pipeline_runner#231 bytes [3968449, 3968471]: "chapter_needs_review"
                                                               ? (v151 = sub_180371890(a1, v140, qword_1803B4648))
                                                               : (v151 = sub_180371D80(
                                                                           a1,
                                                                           (_DWORD)v140,
                                                                           qword_1803CA780,
// Original IDA line 1352 | qword_1803B4648 -> novels.pipeline.pipeline_runner#231 bytes [3968449, 3968471]: "chapter_needs_review"
                                                                           qword_1803B4648,
                                                                           qword_1803AB250)),
                                                                 v151 == 0
// Original IDA line 1355 | qword_1803B4648 -> novels.pipeline.pipeline_runner#231 bytes [3968449, 3968471]: "chapter_needs_review"
                                                              || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4648, v151),
                                                                   (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1357 | qword_1803B4588 -> novels.pipeline.pipeline_runner#207 bytes [3967924, 3967954]: "resolve_write_chapter_status"
                                                                ? (v152 = sub_180371890(a1, v140, qword_1803B4588))
                                                                : (v152 = sub_180371D80(
                                                                            a1,
                                                                            (_DWORD)v140,
                                                                            qword_1803CA780,
// Original IDA line 1362 | qword_1803B4588 -> novels.pipeline.pipeline_runner#207 bytes [3967924, 3967954]: "resolve_write_chapter_status"
                                                                            qword_1803B4588,
                                                                            qword_1803AB250)),
                                                                  v152 == 0))) )
                                                          {
                                                            v22 = 22;
                                                            goto LABEL_206;
                                                          }
// Original IDA line 1369 | qword_1803B4588 -> novels.pipeline.pipeline_runner#207 bytes [3967924, 3967954]: "resolve_write_chapter_status"
                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4588, v152);
                                                          if ( *v140 >= 0 )
                                                          {
                                                            v17 = (*(_QWORD *)v140)-- == 1;
                                                            if ( v17 )
                                                            {
                                                              v153 = sub_18000CC40(v140);
                                                              (*(void (__fastcall **)(int *))(v153 + 48))(v140);
                                                            }
                                                          }
// Original IDA line 1379 | qword_1803B5650 -> novels.pipeline.pipeline_runner#744 bytes [3979130, 3979160]: "services.runtime_state_store"
                                                          v154 = qword_1803B5650;
                                                          v155 = qword_1803CA780;
// Original IDA line 1381 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                          v581 = qword_1803B4258;
// Original IDA line 1382 | qword_1803B5658 -> novels.pipeline.pipeline_runner#745 bytes [3979160, 3979204]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979162, 3979204], "value": "bootstrap_structured_state_from_markdown"}]}
                                                          v542 = qword_1803B5658;
                                                          *(_DWORD *)(v18 + 40) = 28;
                                                          v156 = sub_1803717E0(
                                                                   a1,
                                                                   v154,
                                                                   v155,
                                                                   Py_NoneStruct,
                                                                   v542,
                                                                   v581);
                                                          v157 = (int *)v156;
                                                          if ( v156 != 0 )
                                                          {
                                                            if ( (unsigned int)sub_18000CC50(v156, PyModule_Type) != 0 )
                                                              v158 = sub_180371D80(
                                                                       a1,
                                                                       (_DWORD)v157,
                                                                       qword_1803CA780,
// Original IDA line 1399 | qword_1803B4198 -> novels.pipeline.pipeline_runner#81 bytes [3965393, 3965435]: "bootstrap_structured_state_from_markdown"
                                                                       qword_1803B4198,
                                                                       qword_1803AB250);
                                                            else
// Original IDA line 1402 | qword_1803B4198 -> novels.pipeline.pipeline_runner#81 bytes [3965393, 3965435]: "bootstrap_structured_state_from_markdown"
                                                              v158 = sub_180371890(a1, v157, qword_1803B4198);
                                                            v159 = v158;
                                                            if ( *v157 >= 0 )
                                                            {
                                                              v17 = (*(_QWORD *)v157)-- == 1;
                                                              if ( v17 )
                                                              {
                                                                v160 = sub_18000CC40(v157);
                                                                (*(void (__fastcall **)(int *))(v160 + 48))(v157);
                                                              }
                                                            }
                                                            if ( v159 != 0 )
                                                            {
// Original IDA line 1415 | qword_1803B4198 -> novels.pipeline.pipeline_runner#81 bytes [3965393, 3965435]: "bootstrap_structured_state_from_markdown"
                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4198, v159);
// Original IDA line 1416 | qword_1803B5660 -> novels.pipeline.pipeline_runner#746 bytes [3979204, 3979227]: "services.memory_store"
                                                              v161 = qword_1803B5660;
                                                              v162 = qword_1803CA780;
// Original IDA line 1418 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                              v582 = qword_1803B4258;
// Original IDA line 1419 | qword_1803B5668 -> novels.pipeline.pipeline_runner#747 bytes [3979227, 3979251]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979229, 3979251], "value": "delete_chapters_from"}]}
                                                              v543 = qword_1803B5668;
                                                              *(_DWORD *)(v18 + 40) = 29;
                                                              v163 = sub_1803717E0(
                                                                       a1,
                                                                       v161,
                                                                       v162,
                                                                       Py_NoneStruct,
                                                                       v543,
                                                                       v582);
                                                              v164 = (int *)v163;
                                                              if ( v163 != 0 )
                                                              {
                                                                if ( (unsigned int)sub_18000CC50(v163, PyModule_Type) != 0 )
                                                                  v165 = sub_180371D80(
                                                                           a1,
                                                                           (_DWORD)v164,
                                                                           qword_1803CA780,
// Original IDA line 1436 | qword_1803B4D70 -> novels.pipeline.pipeline_runner#460 bytes [3973688, 3973710]: "delete_chapters_from"
                                                                           qword_1803B4D70,
                                                                           qword_1803AB250);
                                                                else
// Original IDA line 1439 | qword_1803B4D70 -> novels.pipeline.pipeline_runner#460 bytes [3973688, 3973710]: "delete_chapters_from"
                                                                  v165 = sub_180371890(a1, v164, qword_1803B4D70);
                                                                v166 = v165;
                                                                if ( *v164 >= 0 )
                                                                {
                                                                  v17 = (*(_QWORD *)v164)-- == 1;
                                                                  if ( v17 )
                                                                  {
                                                                    v167 = sub_18000CC40(v164);
                                                                    (*(void (__fastcall **)(int *))(v167 + 48))(v164);
                                                                  }
                                                                }
                                                                if ( v166 != 0 )
                                                                {
// Original IDA line 1452 | qword_1803B4D70 -> novels.pipeline.pipeline_runner#460 bytes [3973688, 3973710]: "delete_chapters_from"
                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4D70, v166);
// Original IDA line 1453 | qword_1803B5670 -> novels.pipeline.pipeline_runner#748 bytes [3979251, 3979275]: "services.runtime_state"
                                                                  v168 = qword_1803B5670;
                                                                  v169 = qword_1803CA780;
// Original IDA line 1455 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                  v583 = qword_1803B4258;
// Original IDA line 1456 | qword_1803B5678 -> novels.pipeline.pipeline_runner#749 bytes [3979275, 3979306]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979277, 3979306], "value": "write_chapter_runtime_state"}]}
                                                                  v544 = qword_1803B5678;
                                                                  *(_DWORD *)(v18 + 40) = 30;
                                                                  v170 = sub_1803717E0(
                                                                           a1,
                                                                           v168,
                                                                           v169,
                                                                           Py_NoneStruct,
                                                                           v544,
                                                                           v583);
                                                                  v171 = (int *)v170;
                                                                  if ( v170 != 0 )
                                                                  {
                                                                    if ( (unsigned int)sub_18000CC50(
                                                                                         v170,
                                                                                         PyModule_Type) != 0 )
                                                                      v172 = sub_180371D80(
                                                                               a1,
                                                                               (_DWORD)v171,
                                                                               qword_1803CA780,
// Original IDA line 1475 | qword_1803B4C08 -> novels.pipeline.pipeline_runner#415 bytes [3972680, 3972709]: "write_chapter_runtime_state"
                                                                               qword_1803B4C08,
                                                                               qword_1803AB250);
                                                                    else
// Original IDA line 1478 | qword_1803B4C08 -> novels.pipeline.pipeline_runner#415 bytes [3972680, 3972709]: "write_chapter_runtime_state"
                                                                      v172 = sub_180371890(a1, v171, qword_1803B4C08);
                                                                    v173 = v172;
                                                                    if ( *v171 >= 0 )
                                                                    {
                                                                      v17 = (*(_QWORD *)v171)-- == 1;
                                                                      if ( v17 )
                                                                      {
                                                                        v174 = sub_18000CC40(v171);
                                                                        (*(void (__fastcall **)(int *))(v174 + 48))(v171);
                                                                      }
                                                                    }
                                                                    if ( v173 != 0 )
                                                                    {
                                                                      sub_1800DE7D0(
                                                                        qword_1803CA780,
// Original IDA line 1493 | qword_1803B4C08 -> novels.pipeline.pipeline_runner#415 bytes [3972680, 3972709]: "write_chapter_runtime_state"
                                                                        qword_1803B4C08,
                                                                        v173);
// Original IDA line 1495 | qword_1803B5680 -> novels.pipeline.pipeline_runner#750 bytes [3979306, 3979328]: "services.truth_files"
                                                                      v175 = qword_1803B5680;
                                                                      v176 = qword_1803CA780;
// Original IDA line 1497 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                      v584 = qword_1803B4258;
// Original IDA line 1498 | qword_1803B5688 -> novels.pipeline.pipeline_runner#751 bytes [3979328, 3979404]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979330, 3979364], "value": "default_chapter_summaries_header"}, {"tag": "a", "fileRange": [3979364, 3979385], "value": "read_truth_snapshot"}, {"tag": "a", "fileRange": [3979385, 39794... (truncated literal)
                                                                      v545 = qword_1803B5688;
                                                                      *(_DWORD *)(v18 + 40) = 31;
                                                                      v177 = sub_1803717E0(
                                                                               a1,
                                                                               v175,
                                                                               v176,
                                                                               Py_NoneStruct,
                                                                               v545,
                                                                               v584);
                                                                      v140 = (int *)v177;
                                                                      if ( v177 == 0 )
                                                                      {
                                                                        sub_18000CE30(a1, v605);
                                                                        v22 = 31;
                                                                        goto LABEL_715;
                                                                      }
                                                                      if ( (unsigned int)sub_18000CC50(
                                                                                           v177,
                                                                                           PyModule_Type) != 0 )
                                                                        v178 = sub_180371D80(
                                                                                 a1,
                                                                                 (_DWORD)v140,
                                                                                 qword_1803CA780,
// Original IDA line 1521 | qword_1803B4CF8 -> novels.pipeline.pipeline_runner#445 bytes [3973366, 3973400]: "default_chapter_summaries_header"
                                                                                 qword_1803B4CF8,
                                                                                 qword_1803AB250);
                                                                      else
// Original IDA line 1524 | qword_1803B4CF8 -> novels.pipeline.pipeline_runner#445 bytes [3973366, 3973400]: "default_chapter_summaries_header"
                                                                        v178 = sub_180371890(a1, v140, qword_1803B4CF8);
                                                                      if ( v178 == 0
                                                                        || ((sub_1800DE7D0(
                                                                               qword_1803CA780,
// Original IDA line 1528 | qword_1803B4CF8 -> novels.pipeline.pipeline_runner#445 bytes [3973366, 3973400]: "default_chapter_summaries_header"
                                                                               qword_1803B4CF8,
                                                                               v178),
                                                                             (unsigned int)sub_18000CC50(
                                                                                             v140,
                                                                                             PyModule_Type) == 0)
                                                                          ? (v179 = sub_180371890(
                                                                                      a1,
                                                                                      v140,
// Original IDA line 1536 | qword_1803B4B80 -> novels.pipeline.pipeline_runner#398 bytes [3972304, 3972325]: "read_truth_snapshot"
                                                                                      qword_1803B4B80))
                                                                          : (v179 = sub_180371D80(
                                                                                      a1,
                                                                                      (_DWORD)v140,
                                                                                      qword_1803CA780,
// Original IDA line 1541 | qword_1803B4B80 -> novels.pipeline.pipeline_runner#398 bytes [3972304, 3972325]: "read_truth_snapshot"
                                                                                      qword_1803B4B80,
                                                                                      qword_1803AB250)),
                                                                            v179 == 0
                                                                         || ((sub_1800DE7D0(
                                                                                qword_1803CA780,
// Original IDA line 1546 | qword_1803B4B80 -> novels.pipeline.pipeline_runner#398 bytes [3972304, 3972325]: "read_truth_snapshot"
                                                                                qword_1803B4B80,
                                                                                v179),
                                                                              (unsigned int)sub_18000CC50(
                                                                                              v140,
                                                                                              PyModule_Type) == 0)
                                                                           ? (v180 = sub_180371890(
                                                                                       a1,
                                                                                       v140,
// Original IDA line 1554 | qword_1803B4BF0 -> novels.pipeline.pipeline_runner#412 bytes [3972591, 3972610]: "write_truth_files"
                                                                                       qword_1803B4BF0))
                                                                           : (v180 = sub_180371D80(
                                                                                       a1,
                                                                                       (_DWORD)v140,
                                                                                       qword_1803CA780,
// Original IDA line 1559 | qword_1803B4BF0 -> novels.pipeline.pipeline_runner#412 bytes [3972591, 3972610]: "write_truth_files"
                                                                                       qword_1803B4BF0,
                                                                                       qword_1803AB250)),
                                                                             v180 == 0)) )
                                                                      {
                                                                        v22 = 31;
                                                                        goto LABEL_206;
                                                                      }
                                                                      sub_1800DE7D0(
                                                                        qword_1803CA780,
// Original IDA line 1568 | qword_1803B4BF0 -> novels.pipeline.pipeline_runner#412 bytes [3972591, 3972610]: "write_truth_files"
                                                                        qword_1803B4BF0,
                                                                        v180);
                                                                      if ( *v140 >= 0 )
                                                                      {
                                                                        v17 = (*(_QWORD *)v140)-- == 1;
                                                                        if ( v17 )
                                                                        {
                                                                          v181 = sub_18000CC40(v140);
                                                                          (*(void (__fastcall **)(int *))(v181 + 48))(v140);
                                                                        }
                                                                      }
// Original IDA line 1579 | qword_1803B5690 -> novels.pipeline.pipeline_runner#752 bytes [3979404, 3979427]: "services.volume_utils"
                                                                      v182 = qword_1803B5690;
                                                                      v183 = qword_1803CA780;
// Original IDA line 1581 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                      v585 = qword_1803B4258;
// Original IDA line 1582 | qword_1803B5698 -> novels.pipeline.pipeline_runner#753 bytes [3979427, 3979464]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979429, 3979464], "value": "truncate_summaries_before_chapter"}]}
                                                                      v546 = qword_1803B5698;
                                                                      *(_DWORD *)(v18 + 40) = 32;
                                                                      v184 = sub_1803717E0(
                                                                               a1,
                                                                               v182,
                                                                               v183,
                                                                               Py_NoneStruct,
                                                                               v546,
                                                                               v585);
                                                                      v185 = (int *)v184;
                                                                      if ( v184 != 0 )
                                                                      {
                                                                        if ( (unsigned int)sub_18000CC50(
                                                                                             v184,
                                                                                             PyModule_Type) != 0 )
                                                                          v186 = sub_180371D80(
                                                                                   a1,
                                                                                   (_DWORD)v185,
                                                                                   qword_1803CA780,
// Original IDA line 1601 | qword_1803B4D50 -> novels.pipeline.pipeline_runner#456 bytes [3973612, 3973647]: "truncate_summaries_before_chapter"
                                                                                   qword_1803B4D50,
                                                                                   qword_1803AB250);
                                                                        else
                                                                          v186 = sub_180371890(
                                                                                   a1,
                                                                                   v185,
// Original IDA line 1607 | qword_1803B4D50 -> novels.pipeline.pipeline_runner#456 bytes [3973612, 3973647]: "truncate_summaries_before_chapter"
                                                                                   qword_1803B4D50);
                                                                        v187 = v186;
                                                                        if ( *v185 >= 0 )
                                                                        {
                                                                          v17 = (*(_QWORD *)v185)-- == 1;
                                                                          if ( v17 )
                                                                          {
                                                                            v188 = sub_18000CC40(v185);
                                                                            (*(void (__fastcall **)(int *))(v188 + 48))(v185);
                                                                          }
                                                                        }
                                                                        if ( v187 != 0 )
                                                                        {
                                                                          sub_1800DE7D0(
                                                                            qword_1803CA780,
// Original IDA line 1622 | qword_1803B4D50 -> novels.pipeline.pipeline_runner#456 bytes [3973612, 3973647]: "truncate_summaries_before_chapter"
                                                                            qword_1803B4D50,
                                                                            v187);
// Original IDA line 1624 | qword_1803B56A0 -> novels.pipeline.pipeline_runner#754 bytes [3979464, 3979480]: "models.schemas"
                                                                          v189 = qword_1803B56A0;
                                                                          v190 = qword_1803CA780;
// Original IDA line 1626 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                          v586 = qword_1803B4258;
// Original IDA line 1627 | qword_1803B56A8 -> novels.pipeline.pipeline_runner#755 bytes [3979480, 3979520]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979482, 3979494], "value": "BookConfig"}, {"tag": "a", "fileRange": [3979494, 3979507], "value": "ChapterMeta"}, {"tag": "a", "fileRange": [3979507, 3979520], "value": "utc_now_iso"}]}
                                                                          v547 = qword_1803B56A8;
                                                                          *(_DWORD *)(v18 + 40) = 33;
                                                                          v191 = sub_1803717E0(
                                                                                   a1,
                                                                                   v189,
                                                                                   v190,
                                                                                   Py_NoneStruct,
                                                                                   v547,
                                                                                   v586);
                                                                          v140 = (int *)v191;
                                                                          if ( v191 == 0 )
                                                                          {
                                                                            sub_18000CE30(a1, v605);
                                                                            v22 = 33;
                                                                            goto LABEL_715;
                                                                          }
                                                                          if ( (unsigned int)sub_18000CC50(
                                                                                               v191,
                                                                                               PyModule_Type) != 0 )
                                                                            v192 = sub_180371D80(
                                                                                     a1,
                                                                                     (_DWORD)v140,
                                                                                     qword_1803CA780,
// Original IDA line 1650 | qword_1803B4010 -> novels.pipeline.pipeline_runner#32 bytes [3964338, 3964350]: "BookConfig"
                                                                                     qword_1803B4010,
                                                                                     qword_1803AB250);
                                                                          else
                                                                            v192 = sub_180371890(
                                                                                     a1,
                                                                                     v140,
// Original IDA line 1656 | qword_1803B4010 -> novels.pipeline.pipeline_runner#32 bytes [3964338, 3964350]: "BookConfig"
                                                                                     qword_1803B4010);
                                                                          if ( v192 == 0
                                                                            || ((sub_1800DE7D0(
                                                                                   qword_1803CA780,
// Original IDA line 1660 | qword_1803B4010 -> novels.pipeline.pipeline_runner#32 bytes [3964338, 3964350]: "BookConfig"
                                                                                   qword_1803B4010,
                                                                                   v192),
                                                                                 (unsigned int)sub_18000CC50(
                                                                                                 v140,
                                                                                                 PyModule_Type) == 0)
                                                                              ? (v193 = sub_180371890(
                                                                                          a1,
                                                                                          v140,
// Original IDA line 1668 | qword_1803B45C0 -> novels.pipeline.pipeline_runner#214 bytes [3968186, 3968199]: "ChapterMeta"
                                                                                          qword_1803B45C0))
                                                                              : (v193 = sub_180371D80(
                                                                                          a1,
                                                                                          (_DWORD)v140,
                                                                                          qword_1803CA780,
// Original IDA line 1673 | qword_1803B45C0 -> novels.pipeline.pipeline_runner#214 bytes [3968186, 3968199]: "ChapterMeta"
                                                                                          qword_1803B45C0,
                                                                                          qword_1803AB250)),
                                                                                v193 == 0
                                                                             || ((sub_1800DE7D0(
                                                                                    qword_1803CA780,
// Original IDA line 1678 | qword_1803B45C0 -> novels.pipeline.pipeline_runner#214 bytes [3968186, 3968199]: "ChapterMeta"
                                                                                    qword_1803B45C0,
                                                                                    v193),
                                                                                  (unsigned int)sub_18000CC50(
                                                                                                  v140,
                                                                                                  PyModule_Type) == 0)
                                                                               ? (v194 = sub_180371890(
                                                                                           a1,
                                                                                           v140,
// Original IDA line 1686 | qword_1803B45D0 -> novels.pipeline.pipeline_runner#216 bytes [3968245, 3968258]: "utc_now_iso"
                                                                                           qword_1803B45D0))
                                                                               : (v194 = sub_180371D80(
                                                                                           a1,
                                                                                           (_DWORD)v140,
                                                                                           qword_1803CA780,
// Original IDA line 1691 | qword_1803B45D0 -> novels.pipeline.pipeline_runner#216 bytes [3968245, 3968258]: "utc_now_iso"
                                                                                           qword_1803B45D0,
                                                                                           qword_1803AB250)),
                                                                                 v194 == 0)) )
                                                                          {
                                                                            v22 = 33;
                                                                            goto LABEL_206;
                                                                          }
                                                                          sub_1800DE7D0(
                                                                            qword_1803CA780,
// Original IDA line 1700 | qword_1803B45D0 -> novels.pipeline.pipeline_runner#216 bytes [3968245, 3968258]: "utc_now_iso"
                                                                            qword_1803B45D0,
                                                                            v194);
                                                                          if ( *v140 >= 0 )
                                                                          {
                                                                            v17 = (*(_QWORD *)v140)-- == 1;
                                                                            if ( v17 )
                                                                            {
                                                                              v195 = sub_18000CC40(v140);
                                                                              (*(void (__fastcall **)(int *))(v195 + 48))(v140);
                                                                            }
                                                                          }
// Original IDA line 1711 | qword_1803B56B0 -> novels.pipeline.pipeline_runner#756 bytes [3979520, 3979545]: "storage.file_repository"
                                                                          v196 = qword_1803B56B0;
                                                                          v197 = qword_1803CA780;
// Original IDA line 1713 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                          v587 = qword_1803B4258;
// Original IDA line 1714 | qword_1803B56B8 -> novels.pipeline.pipeline_runner#757 bytes [3979545, 3979563]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979547, 3979563], "value": "FileRepository"}]}
                                                                          v548 = qword_1803B56B8;
                                                                          *(_DWORD *)(v18 + 40) = 34;
                                                                          v198 = sub_1803717E0(
                                                                                   a1,
                                                                                   v196,
                                                                                   v197,
                                                                                   Py_NoneStruct,
                                                                                   v548,
                                                                                   v587);
                                                                          v199 = (int *)v198;
                                                                          if ( v198 != 0 )
                                                                          {
                                                                            if ( (unsigned int)sub_18000CC50(
                                                                                                 v198,
                                                                                                 PyModule_Type) != 0 )
                                                                              v200 = sub_180371D80(
                                                                                       a1,
                                                                                       (_DWORD)v199,
                                                                                       qword_1803CA780,
// Original IDA line 1733 | qword_1803B41B8 -> novels.pipeline.pipeline_runner#85 bytes [3965479, 3965495]: "FileRepository"
                                                                                       qword_1803B41B8,
                                                                                       qword_1803AB250);
                                                                            else
                                                                              v200 = sub_180371890(
                                                                                       a1,
                                                                                       v199,
// Original IDA line 1739 | qword_1803B41B8 -> novels.pipeline.pipeline_runner#85 bytes [3965479, 3965495]: "FileRepository"
                                                                                       qword_1803B41B8);
                                                                            v201 = v200;
                                                                            if ( *v199 >= 0 )
                                                                            {
                                                                              v17 = (*(_QWORD *)v199)-- == 1;
                                                                              if ( v17 )
                                                                              {
                                                                                v202 = sub_18000CC40(v199);
                                                                                (*(void (__fastcall **)(int *))(v202 + 48))(v199);
                                                                              }
                                                                            }
                                                                            if ( v201 != 0 )
                                                                            {
                                                                              sub_1800DE7D0(
                                                                                qword_1803CA780,
// Original IDA line 1754 | qword_1803B41B8 -> novels.pipeline.pipeline_runner#85 bytes [3965479, 3965495]: "FileRepository"
                                                                                qword_1803B41B8,
                                                                                v201);
// Original IDA line 1756 | qword_1803B56C0 -> novels.pipeline.pipeline_runner#758 bytes [3979563, 3979586]: "storage.state_manager"
                                                                              v203 = qword_1803B56C0;
                                                                              v204 = qword_1803CA780;
// Original IDA line 1758 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                              v588 = qword_1803B4258;
// Original IDA line 1759 | qword_1803B56C8 -> novels.pipeline.pipeline_runner#759 bytes [3979586, 3979608]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979588, 3979608], "value": "NovelsStateManager"}]}
                                                                              v549 = qword_1803B56C8;
                                                                              *(_DWORD *)(v18 + 40) = 35;
                                                                              v205 = sub_1803717E0(
                                                                                       a1,
                                                                                       v203,
                                                                                       v204,
                                                                                       Py_NoneStruct,
                                                                                       v549,
                                                                                       v588);
                                                                              v206 = (int *)v205;
                                                                              if ( v205 != 0 )
                                                                              {
                                                                                if ( (unsigned int)sub_18000CC50(
                                                                                                     v205,
                                                                                                     PyModule_Type) != 0 )
                                                                                  v207 = sub_180371D80(
                                                                                           a1,
                                                                                           (_DWORD)v206,
                                                                                           qword_1803CA780,
// Original IDA line 1778 | qword_1803B56D0 -> novels.pipeline.pipeline_runner#760 bytes [3979608, 3979628]: "NovelsStateManager"
                                                                                           qword_1803B56D0,
                                                                                           qword_1803AB250);
                                                                                else
                                                                                  v207 = sub_180371890(
                                                                                           a1,
                                                                                           v206,
// Original IDA line 1784 | qword_1803B56D0 -> novels.pipeline.pipeline_runner#760 bytes [3979608, 3979628]: "NovelsStateManager"
                                                                                           qword_1803B56D0);
                                                                                v208 = v207;
                                                                                if ( *v206 >= 0 )
                                                                                {
                                                                                  v17 = (*(_QWORD *)v206)-- == 1;
                                                                                  if ( v17 )
                                                                                  {
                                                                                    v209 = sub_18000CC40(v206);
                                                                                    (*(void (__fastcall **)(int *))(v209 + 48))(v206);
                                                                                  }
                                                                                }
                                                                                if ( v208 != 0 )
                                                                                {
                                                                                  sub_1800DE7D0(
                                                                                    qword_1803CA780,
// Original IDA line 1799 | qword_1803B56D0 -> novels.pipeline.pipeline_runner#760 bytes [3979608, 3979628]: "NovelsStateManager"
                                                                                    qword_1803B56D0,
                                                                                    v208);
// Original IDA line 1801 | qword_1803B56D8 -> novels.pipeline.pipeline_runner#761 bytes [3979628, 3979645]: "utils.ai_client"
                                                                                  v210 = qword_1803B56D8;
                                                                                  v211 = qword_1803CA780;
// Original IDA line 1803 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                  v589 = qword_1803B4258;
// Original IDA line 1804 | qword_1803B56E0 -> novels.pipeline.pipeline_runner#762 bytes [3979645, 3979663]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979647, 3979663], "value": "NovelsAIClient"}]}
                                                                                  v550 = qword_1803B56E0;
                                                                                  *(_DWORD *)(v18 + 40) = 36;
                                                                                  v212 = sub_1803717E0(
                                                                                           a1,
                                                                                           v210,
                                                                                           v211,
                                                                                           Py_NoneStruct,
                                                                                           v550,
                                                                                           v589);
                                                                                  v213 = (int *)v212;
                                                                                  if ( v212 != 0 )
                                                                                  {
                                                                                    if ( (unsigned int)sub_18000CC50(v212, PyModule_Type) != 0 )
                                                                                      v214 = sub_180371D80(
                                                                                               a1,
                                                                                               (_DWORD)v213,
                                                                                               qword_1803CA780,
// Original IDA line 1821 | qword_1803B3F80 -> novels.pipeline.pipeline_runner#14 bytes [3964014, 3964030]: "NovelsAIClient"
                                                                                               qword_1803B3F80,
                                                                                               qword_1803AB250);
                                                                                    else
                                                                                      v214 = sub_180371890(
                                                                                               a1,
                                                                                               v213,
// Original IDA line 1827 | qword_1803B3F80 -> novels.pipeline.pipeline_runner#14 bytes [3964014, 3964030]: "NovelsAIClient"
                                                                                               qword_1803B3F80);
                                                                                    v215 = v214;
                                                                                    if ( *v213 >= 0 )
                                                                                    {
                                                                                      v17 = (*(_QWORD *)v213)-- == 1;
                                                                                      if ( v17 )
                                                                                      {
                                                                                        v216 = sub_18000CC40(v213);
                                                                                        (*(void (__fastcall **)(int *))(v216 + 48))(v213);
                                                                                      }
                                                                                    }
                                                                                    if ( v215 != 0 )
                                                                                    {
                                                                                      sub_1800DE7D0(
                                                                                        qword_1803CA780,
// Original IDA line 1842 | qword_1803B3F80 -> novels.pipeline.pipeline_runner#14 bytes [3964014, 3964030]: "NovelsAIClient"
                                                                                        qword_1803B3F80,
                                                                                        v215);
// Original IDA line 1844 | qword_1803B56E8 -> novels.pipeline.pipeline_runner#763 bytes [3979663, 3979683]: "utils.cancellation"
                                                                                      v217 = qword_1803B56E8;
                                                                                      v218 = qword_1803CA780;
// Original IDA line 1846 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                      v590 = qword_1803B4258;
// Original IDA line 1847 | qword_1803B56F0 -> novels.pipeline.pipeline_runner#764 bytes [3979683, 3979710]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979685, 3979710], "value": "OperationCancelledError"}]}
                                                                                      v551 = qword_1803B56F0;
                                                                                      *(_DWORD *)(v18 + 40) = 37;
                                                                                      v219 = sub_1803717E0(
                                                                                               a1,
                                                                                               v217,
                                                                                               v218,
                                                                                               Py_NoneStruct,
                                                                                               v551,
                                                                                               v590);
                                                                                      v220 = (int *)v219;
                                                                                      if ( v219 != 0 )
                                                                                      {
                                                                                        if ( (unsigned int)sub_18000CC50(v219, PyModule_Type) != 0 )
                                                                                          v221 = sub_180371D80(
                                                                                                   a1,
                                                                                                   (_DWORD)v220,
                                                                                                   qword_1803CA780,
// Original IDA line 1864 | qword_1803B40C0 -> novels.pipeline.pipeline_runner#54 bytes [3964789, 3964814]: "OperationCancelledError"
                                                                                                   qword_1803B40C0,
                                                                                                   qword_1803AB250);
                                                                                        else
                                                                                          v221 = sub_180371890(
                                                                                                   a1,
                                                                                                   v220,
// Original IDA line 1870 | qword_1803B40C0 -> novels.pipeline.pipeline_runner#54 bytes [3964789, 3964814]: "OperationCancelledError"
                                                                                                   qword_1803B40C0);
                                                                                        v222 = v221;
                                                                                        if ( *v220 >= 0 )
                                                                                        {
                                                                                          v17 = (*(_QWORD *)v220)-- == 1;
                                                                                          if ( v17 )
                                                                                          {
                                                                                            v223 = sub_18000CC40(v220);
                                                                                            (*(void (__fastcall **)(int *))(v223 + 48))(v220);
                                                                                          }
                                                                                        }
                                                                                        if ( v222 != 0 )
                                                                                        {
                                                                                          sub_1800DE7D0(
                                                                                            qword_1803CA780,
// Original IDA line 1885 | qword_1803B40C0 -> novels.pipeline.pipeline_runner#54 bytes [3964789, 3964814]: "OperationCancelledError"
                                                                                            qword_1803B40C0,
                                                                                            v222);
// Original IDA line 1887 | qword_1803B56F8 -> novels.pipeline.pipeline_runner#765 bytes [3979710, 3979734]: "utils.knowledge_bridge"
                                                                                          v224 = qword_1803B56F8;
                                                                                          v225 = qword_1803CA780;
// Original IDA line 1889 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                          v591 = qword_1803B4258;
// Original IDA line 1890 | qword_1803B5700 -> novels.pipeline.pipeline_runner#766 bytes [3979734, 3979785]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979736, 3979760], "value": "load_knowledge_context"}, {"tag": "a", "fileRange": [3979760, 3979785], "value": "load_merged_style_guide"}]}
                                                                                          v552 = qword_1803B5700;
                                                                                          *(_DWORD *)(v18 + 40) = 38;
                                                                                          v226 = sub_1803717E0(
                                                                                                   a1,
                                                                                                   v224,
                                                                                                   v225,
                                                                                                   Py_NoneStruct,
                                                                                                   v552,
                                                                                                   v591);
                                                                                          v140 = (int *)v226;
                                                                                          if ( v226 == 0 )
                                                                                          {
                                                                                            sub_18000CE30(a1, v605);
                                                                                            v22 = 38;
                                                                                            goto LABEL_715;
                                                                                          }
                                                                                          if ( (unsigned int)sub_18000CC50(v226, PyModule_Type) != 0 )
                                                                                            v227 = sub_180371D80(
                                                                                                     a1,
                                                                                                     (_DWORD)v140,
                                                                                                     qword_1803CA780,
// Original IDA line 1911 | qword_1803B4108 -> novels.pipeline.pipeline_runner#63 bytes [3964987, 3965011]: "load_knowledge_context"
                                                                                                     qword_1803B4108,
                                                                                                     qword_1803AB250);
                                                                                          else
                                                                                            v227 = sub_180371890(
                                                                                                     a1,
                                                                                                     v140,
// Original IDA line 1917 | qword_1803B4108 -> novels.pipeline.pipeline_runner#63 bytes [3964987, 3965011]: "load_knowledge_context"
                                                                                                     qword_1803B4108);
                                                                                          if ( v227 == 0
                                                                                            || ((sub_1800DE7D0(
                                                                                                   qword_1803CA780,
// Original IDA line 1921 | qword_1803B4108 -> novels.pipeline.pipeline_runner#63 bytes [3964987, 3965011]: "load_knowledge_context"
                                                                                                   qword_1803B4108,
                                                                                                   v227),
                                                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
// Original IDA line 1924 | qword_1803B4118 -> novels.pipeline.pipeline_runner#65 bytes [3965034, 3965059]: "load_merged_style_guide"
                                                                                              ? (v228 = sub_180371890(a1, v140, qword_1803B4118))
// Original IDA line 1925 | qword_1803B4118 -> novels.pipeline.pipeline_runner#65 bytes [3965034, 3965059]: "load_merged_style_guide"
                                                                                              : (v228 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4118, qword_1803AB250)),
                                                                                                v228 == 0) )
                                                                                          {
                                                                                            v22 = 38;
                                                                                            goto LABEL_206;
                                                                                          }
                                                                                          sub_1800DE7D0(
                                                                                            qword_1803CA780,
// Original IDA line 1933 | qword_1803B4118 -> novels.pipeline.pipeline_runner#65 bytes [3965034, 3965059]: "load_merged_style_guide"
                                                                                            qword_1803B4118,
                                                                                            v228);
                                                                                          if ( *v140 >= 0 )
                                                                                          {
                                                                                            v17 = (*(_QWORD *)v140)-- == 1;
                                                                                            if ( v17 )
                                                                                            {
                                                                                              v229 = sub_18000CC40(v140);
                                                                                              (*(void (__fastcall **)(int *))(v229 + 48))(v140);
                                                                                            }
                                                                                          }
// Original IDA line 1944 | qword_1803B5708 -> novels.pipeline.pipeline_runner#767 bytes [3979785, 3979810]: "services.agent_settings"
                                                                                          v230 = qword_1803B5708;
                                                                                          v231 = qword_1803CA780;
// Original IDA line 1946 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                          v592 = qword_1803B4258;
// Original IDA line 1947 | qword_1803B5710 -> novels.pipeline.pipeline_runner#768 bytes [3979810, 3979832]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979812, 3979832], "value": "AgentSettingsStore"}]}
                                                                                          v553 = qword_1803B5710;
                                                                                          *(_DWORD *)(v18 + 40) = 39;
                                                                                          v232 = sub_1803717E0(
                                                                                                   a1,
                                                                                                   v230,
                                                                                                   v231,
                                                                                                   Py_NoneStruct,
                                                                                                   v553,
                                                                                                   v592);
                                                                                          v233 = (int *)v232;
                                                                                          if ( v232 != 0 )
                                                                                          {
                                                                                            if ( (unsigned int)sub_18000CC50(v232, PyModule_Type) != 0 )
// Original IDA line 1960 | qword_1803B3F58 -> novels.pipeline.pipeline_runner#9 bytes [3963940, 3963960]: "AgentSettingsStore"
                                                                                              v234 = sub_180371D80(a1, (_DWORD)v233, qword_1803CA780, qword_1803B3F58, qword_1803AB250);
                                                                                            else
// Original IDA line 1962 | qword_1803B3F58 -> novels.pipeline.pipeline_runner#9 bytes [3963940, 3963960]: "AgentSettingsStore"
                                                                                              v234 = sub_180371890(a1, v233, qword_1803B3F58);
                                                                                            v235 = v234;
                                                                                            if ( *v233 >= 0 )
                                                                                            {
                                                                                              v17 = (*(_QWORD *)v233)-- == 1;
                                                                                              if ( v17 )
                                                                                              {
                                                                                                v236 = sub_18000CC40(v233);
                                                                                                (*(void (__fastcall **)(int *))(v236 + 48))(v233);
                                                                                              }
                                                                                            }
                                                                                            if ( v235 != 0 )
                                                                                            {
                                                                                              sub_1800DE7D0(
                                                                                                qword_1803CA780,
// Original IDA line 1977 | qword_1803B3F58 -> novels.pipeline.pipeline_runner#9 bytes [3963940, 3963960]: "AgentSettingsStore"
                                                                                                qword_1803B3F58,
                                                                                                v235);
// Original IDA line 1979 | qword_1803B5718 -> novels.pipeline.pipeline_runner#769 bytes [3979832, 3979859]: "services.structured_state"
                                                                                              v237 = qword_1803B5718;
                                                                                              v238 = qword_1803CA780;
// Original IDA line 1981 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                              v593 = qword_1803B4258;
// Original IDA line 1982 | qword_1803B5720 -> novels.pipeline.pipeline_runner#770 bytes [3979859, 3979902]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979861, 3979887], "value": "character_matrix_summary"}, {"tag": "a", "fileRange": [3979887, 3979902], "value": "hooks_summary"}]}
                                                                                              v554 = qword_1803B5720;
                                                                                              *(_DWORD *)(v18 + 40) = 40;
                                                                                              v239 = sub_1803717E0(a1, v237, v238, Py_NoneStruct, v554, v593);
                                                                                              v140 = (int *)v239;
                                                                                              if ( v239 == 0 )
                                                                                              {
                                                                                                sub_18000CE30(a1, v605);
                                                                                                v22 = 40;
                                                                                                goto LABEL_715;
                                                                                              }
                                                                                              if ( (unsigned int)sub_18000CC50(v239, PyModule_Type) != 0 )
// Original IDA line 1993 | qword_1803B54B0 -> novels.pipeline.pipeline_runner#692 bytes [3978068, 3978094]: "character_matrix_summary"
                                                                                                v240 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B54B0, qword_1803AB250);
                                                                                              else
// Original IDA line 1995 | qword_1803B54B0 -> novels.pipeline.pipeline_runner#692 bytes [3978068, 3978094]: "character_matrix_summary"
                                                                                                v240 = sub_180371890(a1, v140, qword_1803B54B0);
                                                                                              if ( v240 == 0
// Original IDA line 1997 | qword_1803B54B0 -> novels.pipeline.pipeline_runner#692 bytes [3978068, 3978094]: "character_matrix_summary"; qword_1803B54A8 -> novels.pipeline.pipeline_runner#691 bytes [3978053, 3978068]: "hooks_summary"
                                                                                                || ((sub_1800DE7D0(qword_1803CA780, qword_1803B54B0, v240), (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0) ? (v241 = sub_180371890(a1, v140, qword_1803B54A8)) : (v241 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B54A8, qword_1803AB250)),
                                                                                                    v241 == 0) )
                                                                                              {
                                                                                                v22 = 40;
                                                                                                goto LABEL_206;
                                                                                              }
                                                                                              sub_1800DE7D0(
                                                                                                qword_1803CA780,
// Original IDA line 2005 | qword_1803B54A8 -> novels.pipeline.pipeline_runner#691 bytes [3978053, 3978068]: "hooks_summary"
                                                                                                qword_1803B54A8,
                                                                                                v241);
                                                                                              if ( *v140 >= 0 )
                                                                                              {
                                                                                                v17 = (*(_QWORD *)v140)-- == 1;
                                                                                                if ( v17 )
                                                                                                {
                                                                                                  v242 = sub_18000CC40(v140);
                                                                                                  (*(void (__fastcall **)(int *))(v242 + 48))(v140);
                                                                                                }
                                                                                              }
// Original IDA line 2016 | qword_1803B5228 -> novels.pipeline.pipeline_runner#611 bytes [3976391, 3976419]: "services.style_fingerprint"
                                                                                              v243 = qword_1803B5228;
                                                                                              v244 = qword_1803CA780;
// Original IDA line 2018 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                              v594 = qword_1803B4258;
// Original IDA line 2019 | qword_1803B5230 -> novels.pipeline.pipeline_runner#612 bytes [3976419, 3976451]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3976421, 3976451], "value": "merge_style_profiles_from_kb"}]}
                                                                                              v555 = qword_1803B5230;
                                                                                              *(_DWORD *)(v18 + 40) = 41;
                                                                                              v245 = sub_1803717E0(a1, v243, v244, Py_NoneStruct, v555, v594);
                                                                                              v246 = (int *)v245;
                                                                                              if ( v245 != 0 )
                                                                                              {
                                                                                                if ( (unsigned int)sub_18000CC50(v245, PyModule_Type) != 0 )
// Original IDA line 2026 | qword_1803B5060 -> novels.pipeline.pipeline_runner#554 bytes [3974826, 3974856]: "merge_style_profiles_from_kb"
                                                                                                  v247 = sub_180371D80(a1, (_DWORD)v246, qword_1803CA780, qword_1803B5060, qword_1803AB250);
                                                                                                else
// Original IDA line 2028 | qword_1803B5060 -> novels.pipeline.pipeline_runner#554 bytes [3974826, 3974856]: "merge_style_profiles_from_kb"
                                                                                                  v247 = sub_180371890(a1, v246, qword_1803B5060);
                                                                                                v248 = v247;
                                                                                                if ( *v246 >= 0 )
                                                                                                {
                                                                                                  v17 = (*(_QWORD *)v246)-- == 1;
                                                                                                  if ( v17 )
                                                                                                  {
                                                                                                    v249 = sub_18000CC40(v246);
                                                                                                    (*(void (__fastcall **)(int *))(v249 + 48))(v246);
                                                                                                  }
                                                                                                }
                                                                                                if ( v248 != 0 )
                                                                                                {
                                                                                                  sub_1800DE7D0(
                                                                                                    qword_1803CA780,
// Original IDA line 2043 | qword_1803B5060 -> novels.pipeline.pipeline_runner#554 bytes [3974826, 3974856]: "merge_style_profiles_from_kb"
                                                                                                    qword_1803B5060,
                                                                                                    v248);
// Original IDA line 2045 | qword_1803B5728 -> novels.pipeline.pipeline_runner#771 bytes [3979902, 3979933]: "services.post_write_validator"
                                                                                                  v250 = qword_1803B5728;
                                                                                                  v251 = qword_1803CA780;
// Original IDA line 2047 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                                  v595 = qword_1803B4258;
// Original IDA line 2048 | qword_1803B5730 -> novels.pipeline.pipeline_runner#772 bytes [3979933, 3979953]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3979935, 3979953], "value": "validate_chapter"}]}
                                                                                                  v556 = qword_1803B5730;
                                                                                                  *(_DWORD *)(v18 + 40) = 42;
                                                                                                  v252 = sub_1803717E0(a1, v250, v251, Py_NoneStruct, v556, v595);
                                                                                                  v253 = (int *)v252;
                                                                                                  if ( v252 != 0 )
                                                                                                  {
                                                                                                    if ( (unsigned int)sub_18000CC50(v252, PyModule_Type) != 0 )
// Original IDA line 2055 | qword_1803B5738 -> novels.pipeline.pipeline_runner#773 bytes [3979953, 3979971]: "validate_chapter"
                                                                                                      v254 = sub_180371D80(a1, (_DWORD)v253, qword_1803CA780, qword_1803B5738, qword_1803AB250);
                                                                                                    else
// Original IDA line 2057 | qword_1803B5738 -> novels.pipeline.pipeline_runner#773 bytes [3979953, 3979971]: "validate_chapter"
                                                                                                      v254 = sub_180371890(a1, v253, qword_1803B5738);
                                                                                                    v255 = v254;
                                                                                                    if ( *v253 >= 0 )
                                                                                                    {
                                                                                                      v17 = (*(_QWORD *)v253)-- == 1;
                                                                                                      if ( v17 )
                                                                                                      {
                                                                                                        v256 = sub_18000CC40(v253);
                                                                                                        (*(void (__fastcall **)(int *))(v256 + 48))(v253);
                                                                                                      }
                                                                                                    }
                                                                                                    if ( v255 != 0 )
                                                                                                    {
// Original IDA line 2070 | qword_1803B46D0 -> novels.pipeline.pipeline_runner#248 bytes [3968818, 3968837]: "_validate_chapter"
                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B46D0, v255);
// Original IDA line 2071 | qword_1803B5740 -> novels.pipeline.pipeline_runner#774 bytes [3979971, 3979999]: "services.ai_tells_detector"
                                                                                                      v257 = qword_1803B5740;
                                                                                                      v258 = qword_1803CA780;
// Original IDA line 2073 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                                      v596 = qword_1803B4258;
// Original IDA line 2074 | qword_1803B5748 -> novels.pipeline.pipeline_runner#775 bytes [3979999, 3980018]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980001, 3980018], "value": "detect_ai_tells"}]}
                                                                                                      v557 = qword_1803B5748;
                                                                                                      *(_DWORD *)(v18 + 40) = 43;
                                                                                                      v259 = sub_1803717E0(a1, v257, v258, Py_NoneStruct, v557, v596);
                                                                                                      v260 = (int *)v259;
                                                                                                      if ( v259 != 0 )
                                                                                                      {
                                                                                                        if ( (unsigned int)sub_18000CC50(v259, PyModule_Type) != 0 )
// Original IDA line 2081 | qword_1803B5750 -> novels.pipeline.pipeline_runner#776 bytes [3980018, 3980035]: "detect_ai_tells"
                                                                                                          v261 = sub_180371D80(a1, (_DWORD)v260, qword_1803CA780, qword_1803B5750, qword_1803AB250);
                                                                                                        else
// Original IDA line 2083 | qword_1803B5750 -> novels.pipeline.pipeline_runner#776 bytes [3980018, 3980035]: "detect_ai_tells"
                                                                                                          v261 = sub_180371890(a1, v260, qword_1803B5750);
                                                                                                        v262 = v261;
                                                                                                        if ( *v260 >= 0 )
                                                                                                        {
                                                                                                          v17 = (*(_QWORD *)v260)-- == 1;
                                                                                                          if ( v17 )
                                                                                                          {
                                                                                                            v263 = sub_18000CC40(v260);
                                                                                                            (*(void (__fastcall **)(int *))(v263 + 48))(v260);
                                                                                                          }
                                                                                                        }
                                                                                                        if ( v262 != 0 )
                                                                                                        {
// Original IDA line 2096 | qword_1803B46F0 -> novels.pipeline.pipeline_runner#252 bytes [3968946, 3968964]: "_detect_ai_tells"
                                                                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B46F0, v262);
// Original IDA line 2097 | qword_1803B5758 -> novels.pipeline.pipeline_runner#777 bytes [3980035, 3980064]: "services.sensitive_detector"
                                                                                                          v264 = qword_1803B5758;
                                                                                                          v265 = qword_1803CA780;
// Original IDA line 2099 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                                          v597 = qword_1803B4258;
// Original IDA line 2100 | qword_1803B5760 -> novels.pipeline.pipeline_runner#778 bytes [3980064, 3980091]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980066, 3980091], "value": "check_sensitive_content"}]}
                                                                                                          v558 = qword_1803B5760;
                                                                                                          *(_DWORD *)(v18 + 40) = 44;
                                                                                                          v266 = sub_1803717E0(a1, v264, v265, Py_NoneStruct, v558, v597);
                                                                                                          v267 = (int *)v266;
                                                                                                          if ( v266 != 0 )
                                                                                                          {
                                                                                                            if ( (unsigned int)sub_18000CC50(v266, PyModule_Type) != 0 )
// Original IDA line 2107 | qword_1803B5768 -> novels.pipeline.pipeline_runner#779 bytes [3980091, 3980116]: "check_sensitive_content"
                                                                                                              v268 = sub_180371D80(a1, (_DWORD)v267, qword_1803CA780, qword_1803B5768, qword_1803AB250);
                                                                                                            else
// Original IDA line 2109 | qword_1803B5768 -> novels.pipeline.pipeline_runner#779 bytes [3980091, 3980116]: "check_sensitive_content"
                                                                                                              v268 = sub_180371890(a1, v267, qword_1803B5768);
                                                                                                            v269 = v268;
                                                                                                            if ( *v267 >= 0 )
                                                                                                            {
                                                                                                              v17 = (*(_QWORD *)v267)-- == 1;
                                                                                                              if ( v17 )
                                                                                                              {
                                                                                                                v270 = sub_18000CC40(v267);
                                                                                                                (*(void (__fastcall **)(int *))(v270 + 48))(v267);
                                                                                                              }
                                                                                                            }
                                                                                                            if ( v269 != 0 )
                                                                                                            {
// Original IDA line 2122 | qword_1803B4708 -> novels.pipeline.pipeline_runner#255 bytes [3968982, 3969000]: "_check_sensitive"
                                                                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4708, v269);
// Original IDA line 2123 | qword_1803B5770 -> novels.pipeline.pipeline_runner#780 bytes [3980116, 3980144]: "services.length_normalizer"
                                                                                                              v271 = qword_1803B5770;
                                                                                                              v272 = qword_1803CA780;
// Original IDA line 2125 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                                              v598 = qword_1803B4258;
// Original IDA line 2126 | qword_1803B5778 -> novels.pipeline.pipeline_runner#781 bytes [3980144, 3980172]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980146, 3980172], "value": "normalize_chapter_length"}]}
                                                                                                              v559 = qword_1803B5778;
                                                                                                              *(_DWORD *)(v18 + 40) = 45;
                                                                                                              v273 = sub_1803717E0(a1, v271, v272, Py_NoneStruct, v559, v598);
                                                                                                              v274 = (int *)v273;
                                                                                                              if ( v273 != 0 )
                                                                                                              {
                                                                                                                if ( (unsigned int)sub_18000CC50(v273, PyModule_Type) != 0 )
// Original IDA line 2133 | qword_1803B5780 -> novels.pipeline.pipeline_runner#782 bytes [3980172, 3980198]: "normalize_chapter_length"
                                                                                                                  v275 = sub_180371D80(a1, (_DWORD)v274, qword_1803CA780, qword_1803B5780, qword_1803AB250);
                                                                                                                else
// Original IDA line 2135 | qword_1803B5780 -> novels.pipeline.pipeline_runner#782 bytes [3980172, 3980198]: "normalize_chapter_length"
                                                                                                                  v275 = sub_180371890(a1, v274, qword_1803B5780);
                                                                                                                v276 = v275;
                                                                                                                if ( *v274 >= 0 )
                                                                                                                {
                                                                                                                  v17 = (*(_QWORD *)v274)-- == 1;
                                                                                                                  if ( v17 )
                                                                                                                  {
                                                                                                                    v277 = sub_18000CC40(v274);
                                                                                                                    (*(void (__fastcall **)(int *))(v277 + 48))(v274);
                                                                                                                  }
                                                                                                                }
                                                                                                                if ( v276 != 0 )
                                                                                                                {
// Original IDA line 2148 | qword_1803B4378 -> novels.pipeline.pipeline_runner#141 bytes [3966597, 3966616]: "_normalize_length"
                                                                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4378, v276);
// Original IDA line 2149 | qword_1803B5788 -> novels.pipeline.pipeline_runner#783 bytes [3980198, 3980236]: "prompts.inkos.genre_profile_defaults"
                                                                                                                  v278 = qword_1803B5788;
                                                                                                                  v279 = qword_1803CA780;
// Original IDA line 2151 | qword_1803B4258 -> novels.pipeline.pipeline_runner#105 bytes [3965902, 3965904]: 2
                                                                                                                  v599 = qword_1803B4258;
// Original IDA line 2152 | qword_1803B5790 -> novels.pipeline.pipeline_runner#784 bytes [3980236, 3980257]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980238, 3980257], "value": "get_genre_profile"}]}
                                                                                                                  v560 = qword_1803B5790;
                                                                                                                  *(_DWORD *)(v18 + 40) = 46;
                                                                                                                  v280 = sub_1803717E0(a1, v278, v279, Py_NoneStruct, v560, v599);
                                                                                                                  v281 = (int *)v280;
                                                                                                                  if ( v280 != 0 )
                                                                                                                  {
                                                                                                                    if ( (unsigned int)sub_18000CC50(v280, PyModule_Type) != 0 )
// Original IDA line 2159 | qword_1803B4490 -> novels.pipeline.pipeline_runner#176 bytes [3967097, 3967116]: "get_genre_profile"
                                                                                                                      v282 = sub_180371D80(a1, (_DWORD)v281, qword_1803CA780, qword_1803B4490, qword_1803AB250);
                                                                                                                    else
// Original IDA line 2161 | qword_1803B4490 -> novels.pipeline.pipeline_runner#176 bytes [3967097, 3967116]: "get_genre_profile"
                                                                                                                      v282 = sub_180371890(a1, v281, qword_1803B4490);
                                                                                                                    v283 = v282;
                                                                                                                    if ( *v281 >= 0 )
                                                                                                                    {
                                                                                                                      v17 = (*(_QWORD *)v281)-- == 1;
                                                                                                                      if ( v17 )
                                                                                                                      {
                                                                                                                        v284 = sub_18000CC40(v281);
                                                                                                                        (*(void (__fastcall **)(int *))(v284 + 48))(v281);
                                                                                                                      }
                                                                                                                    }
                                                                                                                    if ( v283 != 0 )
                                                                                                                    {
// Original IDA line 2174 | qword_1803B4490 -> novels.pipeline.pipeline_runner#176 bytes [3967097, 3967116]: "get_genre_profile"
                                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4490, v283);
// Original IDA line 2175 | qword_1803B5798 -> novels.pipeline.pipeline_runner#785 bytes [3980257, 3980279]: "chapter_review_cycle"
                                                                                                                      v285 = qword_1803B5798;
                                                                                                                      v286 = qword_1803CA780;
                                                                                                                      v600 = qword_1803AB258;
// Original IDA line 2178 | qword_1803B57A0 -> novels.pipeline.pipeline_runner#786 bytes [3980279, 3980307]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980281, 3980307], "value": "run_chapter_review_cycle"}]}
                                                                                                                      v561 = qword_1803B57A0;
                                                                                                                      *(_DWORD *)(v18 + 40) = 47;
                                                                                                                      v287 = sub_1803717E0(a1, v285, v286, Py_NoneStruct, v561, v600);
                                                                                                                      v288 = (int *)v287;
                                                                                                                      if ( v287 != 0 )
                                                                                                                      {
                                                                                                                        if ( (unsigned int)sub_18000CC50(v287, PyModule_Type) != 0 )
// Original IDA line 2185 | qword_1803B44E0 -> novels.pipeline.pipeline_runner#186 bytes [3967321, 3967347]: "run_chapter_review_cycle"
                                                                                                                          v289 = sub_180371D80(a1, (_DWORD)v288, qword_1803CA780, qword_1803B44E0, qword_1803AB250);
                                                                                                                        else
// Original IDA line 2187 | qword_1803B44E0 -> novels.pipeline.pipeline_runner#186 bytes [3967321, 3967347]: "run_chapter_review_cycle"
                                                                                                                          v289 = sub_180371890(a1, v288, qword_1803B44E0);
                                                                                                                        v290 = v289;
                                                                                                                        if ( *v288 >= 0 )
                                                                                                                        {
                                                                                                                          v17 = (*(_QWORD *)v288)-- == 1;
                                                                                                                          if ( v17 )
                                                                                                                          {
                                                                                                                            v291 = sub_18000CC40(v288);
                                                                                                                            (*(void (__fastcall **)(int *))(v291 + 48))(v288);
                                                                                                                          }
                                                                                                                        }
                                                                                                                        if ( v290 != 0 )
                                                                                                                        {
// Original IDA line 2200 | qword_1803B44E0 -> novels.pipeline.pipeline_runner#186 bytes [3967321, 3967347]: "run_chapter_review_cycle"
                                                                                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B44E0, v290);
// Original IDA line 2201 | qword_1803B57A8 -> novels.pipeline.pipeline_runner#787 bytes [3980307, 3980321]: "cover_runner"
                                                                                                                          v292 = qword_1803B57A8;
                                                                                                                          v293 = qword_1803CA780;
                                                                                                                          v601 = qword_1803AB258;
// Original IDA line 2204 | qword_1803B57B0 -> novels.pipeline.pipeline_runner#788 bytes [3980321, 3980336]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980323, 3980336], "value": "CoverRunner"}]}
                                                                                                                          v562 = qword_1803B57B0;
                                                                                                                          *(_DWORD *)(v18 + 40) = 48;
                                                                                                                          v294 = sub_1803717E0(a1, v292, v293, Py_NoneStruct, v562, v601);
                                                                                                                          v295 = (int *)v294;
                                                                                                                          if ( v294 != 0 )
                                                                                                                          {
                                                                                                                            if ( (unsigned int)sub_18000CC50(v294, PyModule_Type) != 0 )
// Original IDA line 2211 | qword_1803B5460 -> novels.pipeline.pipeline_runner#682 bytes [3977801, 3977814]: "CoverRunner"
                                                                                                                              v296 = sub_180371D80(a1, (_DWORD)v295, qword_1803CA780, qword_1803B5460, qword_1803AB250);
                                                                                                                            else
// Original IDA line 2213 | qword_1803B5460 -> novels.pipeline.pipeline_runner#682 bytes [3977801, 3977814]: "CoverRunner"
                                                                                                                              v296 = sub_180371890(a1, v295, qword_1803B5460);
                                                                                                                            v297 = v296;
                                                                                                                            if ( *v295 >= 0 )
                                                                                                                            {
                                                                                                                              v17 = (*(_QWORD *)v295)-- == 1;
                                                                                                                              if ( v17 )
                                                                                                                              {
                                                                                                                                v298 = sub_18000CC40(v295);
                                                                                                                                (*(void (__fastcall **)(int *))(v298 + 48))(v295);
                                                                                                                              }
                                                                                                                            }
                                                                                                                            if ( v297 != 0 )
                                                                                                                            {
// Original IDA line 2226 | qword_1803B5460 -> novels.pipeline.pipeline_runner#682 bytes [3977801, 3977814]: "CoverRunner"
                                                                                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B5460, v297);
// Original IDA line 2227 | qword_1803B57B8 -> novels.pipeline.pipeline_runner#789 bytes [3980336, 3980358]: "pipeline_checkpoints"
                                                                                                                              v299 = qword_1803B57B8;
                                                                                                                              v300 = qword_1803CA780;
                                                                                                                              v602 = qword_1803AB258;
// Original IDA line 2230 | qword_1803B57C0 -> novels.pipeline.pipeline_runner#790 bytes [3980358, 3980401]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980360, 3980374], "value": "is_confirmed"}, {"tag": "a", "fileRange": [3980374, 3980401], "value": "require_secondary_confirm"}]}
                                                                                                                              v563 = qword_1803B57C0;
                                                                                                                              *(_DWORD *)(v18 + 40) = 49;
                                                                                                                              v301 = sub_1803717E0(a1, v299, v300, Py_NoneStruct, v563, v602);
                                                                                                                              v140 = (int *)v301;
                                                                                                                              if ( v301 == 0 )
                                                                                                                              {
                                                                                                                                sub_18000CE30(a1, v605);
                                                                                                                                v22 = 49;
                                                                                                                                goto LABEL_715;
                                                                                                                              }
                                                                                                                              if ( (unsigned int)sub_18000CC50(v301, PyModule_Type) != 0 )
// Original IDA line 2241 | qword_1803B4228 -> novels.pipeline.pipeline_runner#99 bytes [3965723, 3965737]: "is_confirmed"
                                                                                                                                v302 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4228, qword_1803AB250);
                                                                                                                              else
// Original IDA line 2243 | qword_1803B4228 -> novels.pipeline.pipeline_runner#99 bytes [3965723, 3965737]: "is_confirmed"
                                                                                                                                v302 = sub_180371890(a1, v140, qword_1803B4228);
// Original IDA line 2244 | qword_1803B4228 -> novels.pipeline.pipeline_runner#99 bytes [3965723, 3965737]: "is_confirmed"; qword_1803B4220 -> novels.pipeline.pipeline_runner#98 bytes [3965696, 3965723]: "require_secondary_confirm"
                                                                                                                              if ( v302 == 0 || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4228, v302), (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0) ? (v303 = sub_180371890(a1, v140, qword_1803B4220)) : (v303 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4220, qword_1803AB250)), v303 == 0) )
                                                                                                                              {
                                                                                                                                sub_18000CE30(a1, v605);
                                                                                                                                v22 = 49;
                                                                                                                                goto LABEL_207;
                                                                                                                              }
// Original IDA line 2250 | qword_1803B4220 -> novels.pipeline.pipeline_runner#98 bytes [3965696, 3965723]: "require_secondary_confirm"
                                                                                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4220, v303);
                                                                                                                              if ( *v140 >= 0 )
                                                                                                                              {
                                                                                                                                v17 = (*(_QWORD *)v140)-- == 1;
                                                                                                                                if ( v17 )
                                                                                                                                {
                                                                                                                                  v304 = sub_18000CC40(v140);
                                                                                                                                  (*(void (__fastcall **)(int *))(v304 + 48))(v140);
                                                                                                                                }
                                                                                                                              }
// Original IDA line 2260 | qword_1803B57C8 -> novels.pipeline.pipeline_runner#791 bytes [3980401, 3980414]: "play_runner"
                                                                                                                              v305 = qword_1803B57C8;
                                                                                                                              v306 = qword_1803CA780;
                                                                                                                              v603 = qword_1803AB258;
// Original IDA line 2263 | qword_1803B57D0 -> novels.pipeline.pipeline_runner#792 bytes [3980414, 3980428]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980416, 3980428], "value": "PlayRunner"}]}
                                                                                                                              v564 = qword_1803B57D0;
                                                                                                                              *(_DWORD *)(v18 + 40) = 50;
                                                                                                                              v307 = sub_1803717E0(a1, v305, v306, Py_NoneStruct, v564, v603);
                                                                                                                              v308 = (int *)v307;
                                                                                                                              if ( v307 != 0 )
                                                                                                                              {
                                                                                                                                if ( (unsigned int)sub_18000CC50(v307, PyModule_Type) != 0 )
// Original IDA line 2270 | qword_1803B5358 -> novels.pipeline.pipeline_runner#649 bytes [3977219, 3977231]: "PlayRunner"
                                                                                                                                  v309 = sub_180371D80(a1, (_DWORD)v308, qword_1803CA780, qword_1803B5358, qword_1803AB250);
                                                                                                                                else
// Original IDA line 2272 | qword_1803B5358 -> novels.pipeline.pipeline_runner#649 bytes [3977219, 3977231]: "PlayRunner"
                                                                                                                                  v309 = sub_180371890(a1, v308, qword_1803B5358);
                                                                                                                                v310 = v309;
                                                                                                                                if ( *v308 >= 0 )
                                                                                                                                {
                                                                                                                                  v17 = (*(_QWORD *)v308)-- == 1;
                                                                                                                                  if ( v17 )
                                                                                                                                  {
                                                                                                                                    v311 = sub_18000CC40(v308);
                                                                                                                                    (*(void (__fastcall **)(int *))(v311 + 48))(v308);
                                                                                                                                  }
                                                                                                                                }
                                                                                                                                if ( v310 != 0 )
                                                                                                                                {
// Original IDA line 2285 | qword_1803B5358 -> novels.pipeline.pipeline_runner#649 bytes [3977219, 3977231]: "PlayRunner"
                                                                                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B5358, v310);
// Original IDA line 2286 | qword_1803B57D8 -> novels.pipeline.pipeline_runner#793 bytes [3980428, 3980450]: "short_fiction_runner"
                                                                                                                                  v312 = qword_1803B57D8;
                                                                                                                                  v313 = qword_1803CA780;
                                                                                                                                  v604 = qword_1803AB258;
// Original IDA line 2289 | qword_1803B57E0 -> novels.pipeline.pipeline_runner#794 bytes [3980450, 3980472]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3980452, 3980472], "value": "ShortFictionRunner"}]}
                                                                                                                                  v565 = qword_1803B57E0;
                                                                                                                                  *(_DWORD *)(v18 + 40) = 51;
                                                                                                                                  v314 = sub_1803717E0(a1, v312, v313, Py_NoneStruct, v565, v604);
                                                                                                                                  v315 = (int *)v314;
                                                                                                                                  if ( v314 != 0 )
                                                                                                                                  {
                                                                                                                                    if ( (unsigned int)sub_18000CC50(v314, PyModule_Type) != 0 )
// Original IDA line 2296 | qword_1803B52C8 -> novels.pipeline.pipeline_runner#631 bytes [3976749, 3976769]: "ShortFictionRunner"
                                                                                                                                      v316 = sub_180371D80(a1, (_DWORD)v315, qword_1803CA780, qword_1803B52C8, qword_1803AB250);
                                                                                                                                    else
// Original IDA line 2298 | qword_1803B52C8 -> novels.pipeline.pipeline_runner#631 bytes [3976749, 3976769]: "ShortFictionRunner"
                                                                                                                                      v316 = sub_180371890(a1, v315, qword_1803B52C8);
                                                                                                                                    v317 = v316;
                                                                                                                                    if ( *v315 >= 0 )
                                                                                                                                    {
                                                                                                                                      v17 = (*(_QWORD *)v315)-- == 1;
                                                                                                                                      if ( v17 )
                                                                                                                                      {
                                                                                                                                        v318 = sub_18000CC40(v315);
                                                                                                                                        (*(void (__fastcall **)(int *))(v318 + 48))(v315);
                                                                                                                                      }
                                                                                                                                    }
                                                                                                                                    if ( v317 != 0 )
                                                                                                                                    {
// Original IDA line 2311 | qword_1803B52C8 -> novels.pipeline.pipeline_runner#631 bytes [3976749, 3976769]: "ShortFictionRunner"
                                                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B52C8, v317);
                                                                                                                                      v319 = (int *)sub_180370EC0(a1);
                                                                                                                                      v320 = (_DWORD *)sub_180370EC0(a1);
                                                                                                                                      v321 = v320;
                                                                                                                                      if ( *v320 != -1 )
                                                                                                                                        ++*v320;
                                                                                                                                      v322 = sub_1800DFB50(a1);
// Original IDA line 2318 | qword_1803B57E8 -> novels.pipeline.pipeline_runner#795 bytes [3980472, 3980484]: "__module__"
                                                                                                                                      sub_18009B380(v321, qword_1803B57E8, v322);
// Original IDA line 2319 | qword_1803B57F8 -> novels.pipeline.pipeline_runner#797 bytes [3980500, 3980514]: "__qualname__"; qword_1803B57F0 -> novels.pipeline.pipeline_runner#796 bytes [3980484, 3980500]: "PipelineRunner"
                                                                                                                                      sub_18009B380(v321, qword_1803B57F8, qword_1803B57F0);
// Original IDA line 2320 | qword_1803B5800 -> novels.pipeline.pipeline_runner#798 bytes [3980514, 3980613]: {"type": "list", "items": [{"tag": "u", "fileRange": [3980516, 3980536], "value": "保存书籍配置"}, {"tag": "u", "fileRange": [3980536, 3980556], "value": "生成基础设定"}, {"tag": "u", "fileRange": [3980556, 3980576], "value": "审核基础设定"}, {"tag": "u", "f... (truncated literal)
                                                                                                                                      v323 = (int *)sub_180372A40(a1, qword_1803B5800);
// Original IDA line 2321 | qword_1803B4088 -> novels.pipeline.pipeline_runner#47 bytes [3964687, 3964702]: "STAGES_CREATE"
                                                                                                                                      sub_18009B380(v321, qword_1803B4088, v323);
                                                                                                                                      if ( *v323 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v323)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v324 = sub_18000CC40(v323);
                                                                                                                                          (*(void (__fastcall **)(int *))(v324 + 48))(v323);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2331 | qword_1803B5808 -> novels.pipeline.pipeline_runner#799 bytes [3980613, 3980776]: {"type": "list", "items": [{"tag": "u", "fileRange": [3980615, 3980635], "value": "准备章节输入"}, {"tag": "u", "fileRange": [3980635, 3980655], "value": "撰写章节草稿"}, {"tag": "u", "fileRange": [3980655, 3980675], "value": "落盘最终章节"}, {"tag": "u", "f... (truncated literal)
                                                                                                                                      v325 = (int *)sub_180372BB0(a1, qword_1803B5808);
// Original IDA line 2332 | qword_1803B4288 -> novels.pipeline.pipeline_runner#111 bytes [3965972, 3965986]: "STAGES_WRITE"
                                                                                                                                      sub_18009B380(v321, qword_1803B4288, v325);
                                                                                                                                      if ( *v325 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v325)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v326 = sub_18000CC40(v325);
                                                                                                                                          (*(void (__fastcall **)(int *))(v326 + 48))(v325);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2342 | qword_1803B5810 -> novels.pipeline.pipeline_runner#800 bytes [3980776, 3980796]: "整理会话记忆"; qword_1803B5818 -> novels.pipeline.pipeline_runner#801 bytes [3980796, 3980810]: "审稿章节"
                                                                                                                                      v327 = (int *)sub_1803728D0(a1, qword_1803B5810, qword_1803B5818);
// Original IDA line 2343 | qword_1803B4A80 -> novels.pipeline.pipeline_runner#366 bytes [3971484, 3971498]: "STAGES_AUDIT"
                                                                                                                                      sub_18009B380(v321, qword_1803B4A80, v327);
                                                                                                                                      if ( *v327 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v327)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v328 = sub_18000CC40(v327);
                                                                                                                                          (*(void (__fastcall **)(int *))(v328 + 48))(v327);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2353 | qword_1803B5820 -> novels.pipeline.pipeline_runner#802 bytes [3980810, 3980961]: {"type": "list", "items": [{"tag": "u", "fileRange": [3980812, 3980835], "value": "加载修订上下文"}, {"tag": "u", "fileRange": [3980835, 3980849], "value": "修订章节"}, {"tag": "u", "fileRange": [3980849, 3980869], "value": "落盘修订结果"}, {"tag": "u", "fi... (truncated literal)
                                                                                                                                      v329 = (int *)sub_180372BB0(a1, qword_1803B5820);
// Original IDA line 2354 | qword_1803B4AF8 -> novels.pipeline.pipeline_runner#381 bytes [3971810, 3971825]: "STAGES_REVISE"
                                                                                                                                      sub_18009B380(v321, qword_1803B4AF8, v329);
                                                                                                                                      if ( *v329 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v329)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v330 = sub_18000CC40(v329);
                                                                                                                                          (*(void (__fastcall **)(int *))(v330 + 48))(v329);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2364 | qword_1803B5828 -> novels.pipeline.pipeline_runner#803 bytes [3980961, 3980967]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3980963, 3980964], "value": null}, {"tag": "n", "fileRange": [3980964, 3980965], "value": null}, {"tag": "s", "fileRange": [3980965, 3980966], "value": ""}, {"tag": "n", "fileRange": [3... (truncated literal)
                                                                                                                                      v331 = (_DWORD *)qword_1803B5828;
// Original IDA line 2365 | qword_1803B5830 -> novels.pipeline.pipeline_runner#804 bytes [3980967, 3981137]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3980969, 3980976], "value": "state"}, {"tag": "a", "fileRange": [3980976, 3980992], "value": "cancel_checker"}, {"tag": "a", "fileRange": [3980992, 3981007], "value": "event_emitter"}, {"... (truncated literal)
                                                                                                                                      v332 = sub_18036FB80(a1, qword_1803B5830);
                                                                                                                                      if ( *v331 != -1 )
                                                                                                                                        ++*v331;
                                                                                                                                      v333 = (int *)sub_18012F5E0(a1, v331, v332);
// Original IDA line 2369 | qword_1803B5838 -> novels.pipeline.pipeline_runner#805 bytes [3981137, 3981147]: "__init__"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5838, v333);
                                                                                                                                      if ( *v333 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v333)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v334 = sub_18000CC40(v333);
                                                                                                                                          (*(void (__fastcall **)(int *))(v334 + 48))(v333);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2379 | qword_1803B5848 -> novels.pipeline.pipeline_runner#807 bytes [3981172, 3981219]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981174, 3981186], "value": "event_type"}, {"tag": "a", "fileRange": [3981186, 3981195], "value": "payload"}, {"tag": "a", "fileRange": [3981195, 3981203], "value": "return"}], "values": ... (truncated literal)
                                                                                                                                      v335 = sub_18036FB80(a1, qword_1803B5848);
                                                                                                                                      v336 = (int *)sub_18012FA60(a1, v335);
// Original IDA line 2381 | qword_1803B3FA8 -> novels.pipeline.pipeline_runner#19 bytes [3964107, 3964114]: "_emit"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B3FA8, v336);
                                                                                                                                      if ( *v336 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v336)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v337 = sub_18000CC40(v336);
                                                                                                                                          (*(void (__fastcall **)(int *))(v337 + 48))(v336);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2391 | qword_1803B5858 -> novels.pipeline.pipeline_runner#809 bytes [3981241, 3981271]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981243, 3981252], "value": "book_id"}, {"tag": "a", "fileRange": [3981252, 3981260], "value": "return"}], "values": [{"tag": "a", "fileRange": [3981260, 3981265], "value": "str"}, {"tag"... (truncated literal)
                                                                                                                                      v338 = sub_18036FB80(a1, qword_1803B5858);
                                                                                                                                      v339 = (int *)sub_18012FE80(a1, v338);
// Original IDA line 2393 | qword_1803B3FC0 -> novels.pipeline.pipeline_runner#22 bytes [3964232, 3964244]: "_cancelled"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B3FC0, v339);
                                                                                                                                      if ( *v339 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v339)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v340 = sub_18000CC40(v339);
                                                                                                                                          (*(void (__fastcall **)(int *))(v340 + 48))(v339);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2403 | qword_1803B5868 -> novels.pipeline.pipeline_runner#811 bytes [3981298, 3981328]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981300, 3981308], "value": "return"}], "values": [{"tag": "a", "fileRange": [3981308, 3981328], "value": "AgentSettingsStore"}]}
                                                                                                                                      v341 = sub_18036FB80(a1, qword_1803B5868);
                                                                                                                                      v342 = (int *)sub_1801302A0(a1, v341);
// Original IDA line 2405 | qword_1803B3F68 -> novels.pipeline.pipeline_runner#11 bytes [3963966, 3963983]: "_settings_store"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B3F68, v342);
                                                                                                                                      if ( *v342 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v342)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v343 = sub_18000CC40(v342);
                                                                                                                                          (*(void (__fastcall **)(int *))(v343 + 48))(v342);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2415 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v344 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2416 | qword_1803B5878 -> novels.pipeline.pipeline_runner#813 bytes [3981360, 3981422]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981362, 3981369], "value": "agent"}, {"tag": "a", "fileRange": [3981369, 3981379], "value": "explicit"}, {"tag": "a", "fileRange": [3981379, 3981387], "value": "return"}], "values": [{"t... (truncated literal)
                                                                                                                                      v345 = sub_18036FB80(a1, qword_1803B5878);
                                                                                                                                      if ( *v344 != -1 )
                                                                                                                                        ++*v344;
                                                                                                                                      v346 = (int *)sub_1801303C0(a1, v344, v345);
// Original IDA line 2420 | qword_1803B3F90 -> novels.pipeline.pipeline_runner#16 bytes [3964042, 3964058]: "_resolve_model"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B3F90, v346);
                                                                                                                                      if ( *v346 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v346)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v347 = sub_18000CC40(v346);
                                                                                                                                          (*(void (__fastcall **)(int *))(v347 + 48))(v346);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2430 | qword_1803B5888 -> novels.pipeline.pipeline_runner#815 bytes [3981453, 3981467]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3981455, 3981456], "value": null}, {"tag": "a", "fileRange": [3981456, 3981467], "value": "architect"}]}
                                                                                                                                      v348 = (_DWORD *)qword_1803B5888;
// Original IDA line 2431 | qword_1803B5890 -> novels.pipeline.pipeline_runner#816 bytes [3981467, 3981539]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981469, 3981478], "value": "book_id"}, {"tag": "a", "fileRange": [3981478, 3981485], "value": "model"}, {"tag": "a", "fileRange": [3981485, 3981492], "value": "agent"}, {"tag": "a", "fil... (truncated literal)
                                                                                                                                      v349 = sub_18036FB80(a1, qword_1803B5890);
                                                                                                                                      if ( *v348 != -1 )
                                                                                                                                        ++*v348;
                                                                                                                                      v350 = (int *)sub_180130420(a1, v348, v349);
// Original IDA line 2435 | qword_1803B3FD8 -> novels.pipeline.pipeline_runner#25 bytes [3964268, 3964274]: "_ctx"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B3FD8, v350);
                                                                                                                                      if ( *v350 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v350)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v351 = sub_18000CC40(v350);
                                                                                                                                          (*(void (__fastcall **)(int *))(v351 + 48))(v350);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2445 | qword_1803B58A0 -> novels.pipeline.pipeline_runner#818 bytes [3981560, 3981618]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981562, 3981569], "value": "model"}, {"tag": "a", "fileRange": [3981569, 3981576], "value": "agent"}, {"tag": "a", "fileRange": [3981576, 3981584], "value": "return"}], "values": [{"tag"... (truncated literal)
                                                                                                                                      v352 = sub_18036FB80(a1, qword_1803B58A0);
                                                                                                                                      v353 = (int *)sub_180130480(a1, v352);
// Original IDA line 2447 | qword_1803B52F0 -> novels.pipeline.pipeline_runner#636 bytes [3976886, 3976910]: "_ctx_for_short_fiction"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B52F0, v353);
                                                                                                                                      if ( *v353 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v353)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v354 = sub_18000CC40(v353);
                                                                                                                                          (*(void (__fastcall **)(int *))(v354 + 48))(v353);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2457 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v355 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2458 | qword_1803B58B0 -> novels.pipeline.pipeline_runner#820 bytes [3981657, 3981726]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981659, 3981668], "value": "payload"}, {"tag": "a", "fileRange": [3981668, 3981675], "value": "model"}, {"tag": "a", "fileRange": [3981675, 3981683], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v356 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v355 != -1 )
                                                                                                                                        ++*v355;
                                                                                                                                      v357 = (int *)sub_1801304E0(a1, v355, v356);
// Original IDA line 2462 | qword_1803B5038 -> novels.pipeline.pipeline_runner#549 bytes [3974738, 3974751]: "create_book"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5038, v357);
                                                                                                                                      if ( *v357 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v357)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v358 = sub_18000CC40(v357);
                                                                                                                                          (*(void (__fastcall **)(int *))(v358 + 48))(v357);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2472 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v359 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2473 | qword_1803B58C0 -> novels.pipeline.pipeline_runner#822 bytes [3981754, 3981816]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981756, 3981765], "value": "book_id"}, {"tag": "a", "fileRange": [3981765, 3981772], "value": "model"}, {"tag": "a", "fileRange": [3981772, 3981780], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v360 = sub_18036FB80(a1, qword_1803B58C0);
                                                                                                                                      if ( *v359 != -1 )
                                                                                                                                        ++*v359;
                                                                                                                                      v361 = (int *)sub_1801305A0(a1, v359, v360);
// Original IDA line 2477 | qword_1803B4780 -> novels.pipeline.pipeline_runner#270 bytes [3969248, 3969268]: "write_next_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4780, v361);
                                                                                                                                      if ( *v361 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v361)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v362 = sub_18000CC40(v361);
                                                                                                                                          (*(void (__fastcall **)(int *))(v362 + 48))(v361);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2487 | qword_1803B58D0 -> novels.pipeline.pipeline_runner#824 bytes [3981851, 3981856]: {"type": "tuple", "items": [{"tag": "l", "fileRange": [3981853, 3981855], "value": 1}, {"tag": "n", "fileRange": [3981855, 3981856], "value": null}]}
                                                                                                                                      v363 = (_DWORD *)qword_1803B58D0;
// Original IDA line 2488 | qword_1803B58D8 -> novels.pipeline.pipeline_runner#825 bytes [3981856, 3981886]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981858, 3981884], "value": "max_consecutive_failures"}], "values": [{"tag": "l", "fileRange": [3981884, 3981886], "value": 3}]}
                                                                                                                                      v364 = sub_18036FB80(a1, qword_1803B58D8);
// Original IDA line 2489 | qword_1803B58E0 -> novels.pipeline.pipeline_runner#826 bytes [3981886, 3981991]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981888, 3981897], "value": "book_id"}, {"tag": "a", "fileRange": [3981897, 3981904], "value": "count"}, {"tag": "a", "fileRange": [3981904, 3981911], "value": "model"}, {"tag": "a", "fil... (truncated literal)
                                                                                                                                      v365 = sub_18036FB80(a1, qword_1803B58E0);
                                                                                                                                      if ( *v363 != -1 )
                                                                                                                                        ++*v363;
                                                                                                                                      v366 = (int *)sub_18012F220(a1, v363, v364, v365);
// Original IDA line 2493 | qword_1803B58E8 -> novels.pipeline.pipeline_runner#827 bytes [3981991, 3982013]: "write_batch_chapters"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B58E8, v366);
                                                                                                                                      if ( *v366 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v366)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v367 = sub_18000CC40(v366);
                                                                                                                                          (*(void (__fastcall **)(int *))(v367 + 48))(v366);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2503 | qword_1803B58F8 -> novels.pipeline.pipeline_runner#829 bytes [3982050, 3982090]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982052, 3982061], "value": "book_id"}, {"tag": "a", "fileRange": [3982061, 3982069], "value": "return"}], "values": [{"tag": "a", "fileRange": [3982069, 3982074], "value": "str"}, {"tag"... (truncated literal)
                                                                                                                                      v368 = sub_18036FB80(a1, qword_1803B58F8);
                                                                                                                                      v369 = (int *)sub_18012F280(a1, v368);
// Original IDA line 2505 | qword_1803B4E40 -> novels.pipeline.pipeline_runner#486 bytes [3974094, 3974112]: "get_book_payload"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4E40, v369);
                                                                                                                                      if ( *v369 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v369)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v370 = sub_18000CC40(v369);
                                                                                                                                          (*(void (__fastcall **)(int *))(v370 + 48))(v369);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2515 | qword_1803B5908 -> novels.pipeline.pipeline_runner#831 bytes [3982123, 3982184]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982125, 3982138], "value": "story_files"}, {"tag": "a", "fileRange": [3982138, 3982146], "value": "return"}], "values": [{"tag": "u", "fileRange": [3982146, 3982162], "value": "Dict[str,... (truncated literal)
                                                                                                                                      v371 = sub_18036FB80(a1, qword_1803B5908);
                                                                                                                                      v372 = (int *)sub_18012F2E0(a1, v371);
// Original IDA line 2517 | qword_1803B4820 -> novels.pipeline.pipeline_runner#290 bytes [3969783, 3969806]: "_foundation_file_list"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4820, v372);
                                                                                                                                      if ( *v372 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v372)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v373 = sub_18000CC40(v372);
                                                                                                                                          (*(void (__fastcall **)(int *))(v373 + 48))(v372);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2527 | qword_1803B5918 -> novels.pipeline.pipeline_runner#833 bytes [3982222, 3982268]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982224, 3982233], "value": "book_id"}, {"tag": "a", "fileRange": [3982233, 3982241], "value": "return"}], "values": [{"tag": "a", "fileRange": [3982241, 3982246], "value": "str"}, {"tag"... (truncated literal)
                                                                                                                                      v374 = sub_18036FB80(a1, qword_1803B5918);
                                                                                                                                      v375 = (int *)sub_18012F340(a1, v374);
// Original IDA line 2529 | qword_1803B4828 -> novels.pipeline.pipeline_runner#291 bytes [3969806, 3969824]: "_list_characters"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4828, v375);
                                                                                                                                      if ( *v375 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v375)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v376 = sub_18000CC40(v375);
                                                                                                                                          (*(void (__fastcall **)(int *))(v376 + 48))(v375);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2539 | qword_1803B5928 -> novels.pipeline.pipeline_runner#835 bytes [3982301, 3982332]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982303, 3982312], "value": "book_id"}, {"tag": "a", "fileRange": [3982312, 3982317], "value": "key"}, {"tag": "a", "fileRange": [3982317, 3982325], "value": "return"}], "values": [{"tag"... (truncated literal)
                                                                                                                                      v377 = sub_18036FB80(a1, qword_1803B5928);
                                                                                                                                      v378 = (int *)sub_18012F3A0(a1, v377);
// Original IDA line 2541 | qword_1803B5930 -> novels.pipeline.pipeline_runner#836 bytes [3982332, 3982349]: "read_story_file"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5930, v378);
                                                                                                                                      if ( *v378 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v378)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v379 = sub_18000CC40(v378);
                                                                                                                                          (*(void (__fastcall **)(int *))(v379 + 48))(v378);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2551 | qword_1803B5940 -> novels.pipeline.pipeline_runner#838 bytes [3982381, 3982427]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982383, 3982392], "value": "book_id"}, {"tag": "a", "fileRange": [3982392, 3982397], "value": "key"}, {"tag": "a", "fileRange": [3982397, 3982406], "value": "content"}, {"tag": "a", "fil... (truncated literal)
                                                                                                                                      v380 = sub_18036FB80(a1, qword_1803B5940);
                                                                                                                                      v381 = (int *)sub_18012F400(a1, v380);
// Original IDA line 2553 | qword_1803B5948 -> novels.pipeline.pipeline_runner#839 bytes [3982427, 3982444]: "save_story_file"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5948, v381);
                                                                                                                                      if ( *v381 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v381)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v382 = sub_18000CC40(v381);
                                                                                                                                          (*(void (__fastcall **)(int *))(v382 + 48))(v381);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2563 | qword_1803B5958 -> novels.pipeline.pipeline_runner#841 bytes [3982476, 3982513]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982478, 3982490], "value": "sync_truth"}, {"tag": "a", "fileRange": [3982490, 3982503], "value": "replay_from"}, {"tag": "a", "fileRange": [3982503, 3982510], "value": "model"}], "values... (truncated literal)
                                                                                                                                      v383 = sub_18036FB80(a1, qword_1803B5958);
// Original IDA line 2564 | qword_1803B5960 -> novels.pipeline.pipeline_runner#842 bytes [3982513, 3982642]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982515, 3982524], "value": "book_id"}, {"tag": "a", "fileRange": [3982524, 3982540], "value": "chapter_number"}, {"tag": "a", "fileRange": [3982540, 3982549], "value": "content"}, {"tag"... (truncated literal)
                                                                                                                                      v384 = sub_18036FB80(a1, qword_1803B5960);
                                                                                                                                      v385 = (int *)sub_18012F460(a1, v383, v384);
// Original IDA line 2566 | qword_1803B5968 -> novels.pipeline.pipeline_runner#843 bytes [3982642, 3982656]: "save_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5968, v385);
                                                                                                                                      if ( *v385 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v385)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v386 = sub_18000CC40(v385);
                                                                                                                                          (*(void (__fastcall **)(int *))(v386 + 48))(v385);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2576 | qword_1803B5978 -> novels.pipeline.pipeline_runner#845 bytes [3982685, 3982689]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3982687, 3982688], "value": null}, {"tag": "n", "fileRange": [3982688, 3982689], "value": null}]}
                                                                                                                                      v387 = (_DWORD *)qword_1803B5978;
// Original IDA line 2577 | qword_1803B5980 -> novels.pipeline.pipeline_runner#846 bytes [3982689, 3982782]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982691, 3982700], "value": "book_id"}, {"tag": "a", "fileRange": [3982700, 3982716], "value": "chapter_number"}, {"tag": "a", "fileRange": [3982716, 3982723], "value": "model"}, {"tag": ... (truncated literal)
                                                                                                                                      v388 = sub_18036FB80(a1, qword_1803B5980);
                                                                                                                                      if ( *v387 != -1 )
                                                                                                                                        ++*v387;
                                                                                                                                      v389 = (int *)sub_18012F4C0(a1, v387, v388);
// Original IDA line 2581 | qword_1803B5988 -> novels.pipeline.pipeline_runner#847 bytes [3982782, 3982797]: "audit_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5988, v389);
                                                                                                                                      if ( *v389 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v389)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v390 = sub_18000CC40(v389);
                                                                                                                                          (*(void (__fastcall **)(int *))(v390 + 48))(v389);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2591 | qword_1803B5998 -> novels.pipeline.pipeline_runner#849 bytes [3982827, 3982838]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3982829, 3982830], "value": null}, {"tag": "a", "fileRange": [3982830, 3982836], "value": "auto"}, {"tag": "s", "fileRange": [3982836, 3982837], "value": ""}, {"tag": "n", "fileRange": ... (truncated literal)
                                                                                                                                      v391 = (_DWORD *)qword_1803B5998;
// Original IDA line 2592 | qword_1803B59A0 -> novels.pipeline.pipeline_runner#850 bytes [3982838, 3982956]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3982840, 3982849], "value": "book_id"}, {"tag": "a", "fileRange": [3982849, 3982865], "value": "chapter_number"}, {"tag": "a", "fileRange": [3982865, 3982871], "value": "mode"}, {"tag": "... (truncated literal)
                                                                                                                                      v392 = sub_18036FB80(a1, qword_1803B59A0);
                                                                                                                                      if ( *v391 != -1 )
                                                                                                                                        ++*v391;
                                                                                                                                      v393 = (int *)sub_18012F520(a1, v391, v392);
// Original IDA line 2596 | qword_1803B59A8 -> novels.pipeline.pipeline_runner#851 bytes [3982956, 3982972]: "revise_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B59A8, v393);
                                                                                                                                      if ( *v393 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v393)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v394 = sub_18000CC40(v393);
                                                                                                                                          (*(void (__fastcall **)(int *))(v394 + 48))(v393);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2606 | qword_1803B59B8 -> novels.pipeline.pipeline_runner#853 bytes [3983003, 3983102]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983005, 3983014], "value": "book_id"}, {"tag": "a", "fileRange": [3983014, 3983020], "value": "book"}, {"tag": "a", "fileRange": [3983020, 3983036], "value": "chapter_number"}, {"tag": "... (truncated literal)
                                                                                                                                      v395 = sub_18036FB80(a1, qword_1803B59B8);
                                                                                                                                      v396 = (int *)sub_18012F580(a1, v395);
// Original IDA line 2608 | qword_1803B4A28 -> novels.pipeline.pipeline_runner#355 bytes [3971222, 3971252]: "_settle_validate_write_truth"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4A28, v396);
                                                                                                                                      if ( *v396 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v396)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v397 = sub_18000CC40(v396);
                                                                                                                                          (*(void (__fastcall **)(int *))(v397 + 48))(v396);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2618 | qword_1803B59C8 -> novels.pipeline.pipeline_runner#855 bytes [3983147, 3983350]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983149, 3983158], "value": "book_id"}, {"tag": "a", "fileRange": [3983158, 3983164], "value": "book"}, {"tag": "a", "fileRange": [3983164, 3983180], "value": "chapter_number"}, {"tag": "... (truncated literal)
                                                                                                                                      v398 = sub_18036FB80(a1, qword_1803B59C8);
                                                                                                                                      v399 = (int *)sub_18012F640(a1, v398);
// Original IDA line 2620 | qword_1803B4568 -> novels.pipeline.pipeline_runner#203 bytes [3967811, 3967836]: "_apply_truth_and_memory"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4568, v399);
                                                                                                                                      if ( *v399 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v399)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v400 = sub_18000CC40(v399);
                                                                                                                                          (*(void (__fastcall **)(int *))(v400 + 48))(v399);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2630 | qword_1803B59D8 -> novels.pipeline.pipeline_runner#857 bytes [3983390, 3983483]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983392, 3983403], "value": "story_dir"}, {"tag": "a", "fileRange": [3983403, 3983419], "value": "chapter_number"}, {"tag": "a", "fileRange": [3983419, 3983426], "value": "title"}, {"tag"... (truncated literal)
                                                                                                                                      v401 = sub_18036FB80(a1, qword_1803B59D8);
                                                                                                                                      v402 = (int *)sub_18012F6A0(a1, v401);
// Original IDA line 2632 | qword_1803B4C58 -> novels.pipeline.pipeline_runner#425 bytes [3972954, 3972977]: "_sync_temporal_memory"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4C58, v402);
                                                                                                                                      if ( *v402 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v402)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v403 = sub_18000CC40(v402);
                                                                                                                                          (*(void (__fastcall **)(int *))(v403 + 48))(v402);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2642 | qword_1803B59E8 -> novels.pipeline.pipeline_runner#859 bytes [3983521, 3983636]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983523, 3983532], "value": "book_id"}, {"tag": "a", "fileRange": [3983532, 3983538], "value": "book"}, {"tag": "a", "fileRange": [3983538, 3983545], "value": "index"}, {"tag": "a", "file... (truncated literal)
                                                                                                                                      v404 = sub_18036FB80(a1, qword_1803B59E8);
                                                                                                                                      v405 = (int *)sub_18012F700(a1, v404);
// Original IDA line 2644 | qword_1803B5278 -> novels.pipeline.pipeline_runner#621 bytes [3976541, 3976569]: "_replay_import_truth_files"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5278, v405);
                                                                                                                                      if ( *v405 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v405)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v406 = sub_18000CC40(v405);
                                                                                                                                          (*(void (__fastcall **)(int *))(v406 + 48))(v405);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2654 | qword_1803B59F8 -> novels.pipeline.pipeline_runner#861 bytes [3983679, 3983707]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983681, 3983693], "value": "log_prefix"}], "values": [{"tag": "u", "fileRange": [3983693, 3983707], "value": "真相回放"}]}
                                                                                                                                      v407 = sub_18036FB80(a1, qword_1803B59F8);
// Original IDA line 2655 | qword_1803B5A00 -> novels.pipeline.pipeline_runner#862 bytes [3983707, 3983813]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983709, 3983718], "value": "book_id"}, {"tag": "a", "fileRange": [3983718, 3983724], "value": "book"}, {"tag": "a", "fileRange": [3983724, 3983738], "value": "from_chapter"}, {"tag": "a"... (truncated literal)
                                                                                                                                      v408 = sub_18036FB80(a1, qword_1803B5A00);
                                                                                                                                      v409 = (int *)sub_18012F760(a1, v407, v408);
// Original IDA line 2657 | qword_1803B4A18 -> novels.pipeline.pipeline_runner#353 bytes [3971160, 3971188]: "_replay_truth_from_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4A18, v409);
                                                                                                                                      if ( *v409 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v409)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v410 = sub_18000CC40(v409);
                                                                                                                                          (*(void (__fastcall **)(int *))(v410 + 48))(v409);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2667 | qword_1803B5A10 -> novels.pipeline.pipeline_runner#864 bytes [3983856, 3983947]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983858, 3983867], "value": "book_id"}, {"tag": "a", "fileRange": [3983867, 3983873], "value": "book"}, {"tag": "a", "fileRange": [3983873, 3983880], "value": "model"}, {"tag": "a", "file... (truncated literal)
                                                                                                                                      v411 = sub_18036FB80(a1, qword_1803B5A10);
                                                                                                                                      v412 = (int *)sub_18012F820(a1, v411);
// Original IDA line 2669 | qword_1803B45E8 -> novels.pipeline.pipeline_runner#219 bytes [3968284, 3968312]: "_maybe_consolidate_volumes"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B45E8, v412);
                                                                                                                                      if ( *v412 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v412)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v413 = sub_18000CC40(v412);
                                                                                                                                          (*(void (__fastcall **)(int *))(v413 + 48))(v412);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2679 | qword_1803B5A20 -> novels.pipeline.pipeline_runner#866 bytes [3983990, 3984050]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3983992, 3983999], "value": "audit"}, {"tag": "a", "fileRange": [3983999, 3984015], "value": "chapter_number"}, {"tag": "a", "fileRange": [3984015, 3984023], "value": "return"}], "values"... (truncated literal)
                                                                                                                                      v414 = sub_18036FB80(a1, qword_1803B5A20);
                                                                                                                                      v415 = (int *)sub_18012F880(a1, v414);
// Original IDA line 2681 | qword_1803B4530 -> novels.pipeline.pipeline_runner#196 bytes [3967723, 3967743]: "_append_audit_logs"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4530, v415);
                                                                                                                                      if ( *v415 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v415)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v416 = sub_18000CC40(v415);
                                                                                                                                          (*(void (__fastcall **)(int *))(v416 + 48))(v415);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2691 | qword_1803B5A30 -> novels.pipeline.pipeline_runner#868 bytes [3984085, 3984091]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3984087, 3984091], "value": "md"}]}
                                                                                                                                      v417 = (_DWORD *)qword_1803B5A30;
// Original IDA line 2692 | qword_1803B5A38 -> novels.pipeline.pipeline_runner#869 bytes [3984091, 3984122]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984093, 3984102], "value": "book_id"}, {"tag": "a", "fileRange": [3984102, 3984107], "value": "fmt"}, {"tag": "a", "fileRange": [3984107, 3984115], "value": "return"}], "values": [{"tag"... (truncated literal)
                                                                                                                                      v418 = sub_18036FB80(a1, qword_1803B5A38);
                                                                                                                                      if ( *v417 != -1 )
                                                                                                                                        ++*v417;
                                                                                                                                      v419 = (int *)sub_18012F8E0(a1, v417, v418);
// Original IDA line 2696 | qword_1803B5A40 -> novels.pipeline.pipeline_runner#870 bytes [3984122, 3984135]: "export_book"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5A40, v419);
                                                                                                                                      if ( *v419 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v419)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v420 = sub_18000CC40(v419);
                                                                                                                                          (*(void (__fastcall **)(int *))(v420 + 48))(v419);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2706 | qword_1803B5A50 -> novels.pipeline.pipeline_runner#872 bytes [3984163, 3984178]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984165, 3984173], "value": "return"}], "values": [{"tag": "a", "fileRange": [3984173, 3984178], "value": "str"}]}
                                                                                                                                      v421 = sub_18036FB80(a1, qword_1803B5A50);
                                                                                                                                      v422 = (int *)sub_18012F940(a1, v421);
// Original IDA line 2708 | qword_1803B4E68 -> novels.pipeline.pipeline_runner#491 bytes [3974150, 3974162]: "_export_md"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4E68, v422);
                                                                                                                                      if ( *v422 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v422)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v423 = sub_18000CC40(v422);
                                                                                                                                          (*(void (__fastcall **)(int *))(v423 + 48))(v422);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2718 | qword_1803B5A50 -> novels.pipeline.pipeline_runner#872 bytes [3984163, 3984178]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984165, 3984173], "value": "return"}], "values": [{"tag": "a", "fileRange": [3984173, 3984178], "value": "str"}]}
                                                                                                                                      v424 = sub_18036FB80(a1, qword_1803B5A50);
                                                                                                                                      v425 = (int *)sub_18012F9A0(a1, v424);
// Original IDA line 2720 | qword_1803B4E50 -> novels.pipeline.pipeline_runner#488 bytes [3974117, 3974130]: "_export_txt"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4E50, v425);
                                                                                                                                      if ( *v425 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v425)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v426 = sub_18000CC40(v425);
                                                                                                                                          (*(void (__fastcall **)(int *))(v426 + 48))(v425);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2730 | qword_1803B5A68 -> novels.pipeline.pipeline_runner#875 bytes [3984233, 3984258]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984235, 3984244], "value": "book_id"}, {"tag": "a", "fileRange": [3984244, 3984252], "value": "return"}], "values": [{"tag": "a", "fileRange": [3984252, 3984257], "value": "str"}, {"tag"... (truncated literal)
                                                                                                                                      v427 = sub_18036FB80(a1, qword_1803B5A68);
                                                                                                                                      v428 = (int *)sub_18012FA00(a1, v427);
// Original IDA line 2732 | qword_1803B4E60 -> novels.pipeline.pipeline_runner#490 bytes [3974136, 3974150]: "_export_epub"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4E60, v428);
                                                                                                                                      if ( *v428 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v428)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v429 = sub_18000CC40(v428);
                                                                                                                                          (*(void (__fastcall **)(int *))(v429 + 48))(v428);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2742 | qword_1803B5A78 -> novels.pipeline.pipeline_runner#877 bytes [3984287, 3984337]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984289, 3984298], "value": "book_id"}, {"tag": "a", "fileRange": [3984298, 3984314], "value": "chapter_number"}, {"tag": "a", "fileRange": [3984314, 3984322], "value": "return"}], "value... (truncated literal)
                                                                                                                                      v430 = sub_18036FB80(a1, qword_1803B5A78);
                                                                                                                                      v431 = (int *)sub_18012FAC0(a1, v430);
// Original IDA line 2744 | qword_1803B4A68 -> novels.pipeline.pipeline_runner#363 bytes [3971446, 3971460]: "read_chapter"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4A68, v431);
                                                                                                                                      if ( *v431 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v431)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v432 = sub_18000CC40(v431);
                                                                                                                                          (*(void (__fastcall **)(int *))(v432 + 48))(v431);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2754 | qword_1803B4EE8 -> novels.pipeline.pipeline_runner#507 bytes [3974277, 3974280]: {"type": "tuple", "items": [{"tag": "s", "fileRange": [3974279, 3974280], "value": ""}]}
                                                                                                                                      v433 = (_DWORD *)qword_1803B4EE8;
// Original IDA line 2755 | qword_1803B5A88 -> novels.pipeline.pipeline_runner#879 bytes [3984366, 3984429]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984368, 3984375], "value": "index"}, {"tag": "a", "fileRange": [3984375, 3984383], "value": "status"}, {"tag": "a", "fileRange": [3984383, 3984392], "value": "message"}, {"tag": "a", "fi... (truncated literal)
                                                                                                                                      v434 = sub_18036FB80(a1, qword_1803B5A88);
                                                                                                                                      if ( *v433 != -1 )
                                                                                                                                        ++*v433;
                                                                                                                                      v435 = (int *)sub_18012FB20(a1, v433, v434);
// Original IDA line 2759 | qword_1803B4060 -> novels.pipeline.pipeline_runner#42 bytes [3964618, 3964626]: "_stage"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4060, v435);
                                                                                                                                      if ( *v435 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v435)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v436 = sub_18000CC40(v435);
                                                                                                                                          (*(void (__fastcall **)(int *))(v436 + 48))(v435);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2769 | qword_1803B5A98 -> novels.pipeline.pipeline_runner#881 bytes [3984452, 3984482]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984454, 3984463], "value": "message"}, {"tag": "a", "fileRange": [3984463, 3984471], "value": "return"}], "values": [{"tag": "a", "fileRange": [3984471, 3984476], "value": "str"}, {"tag"... (truncated literal)
                                                                                                                                      v437 = sub_18036FB80(a1, qword_1803B5A98);
                                                                                                                                      v438 = (int *)sub_18012FB80(a1, v437);
// Original IDA line 2771 | qword_1803B4260 -> novels.pipeline.pipeline_runner#106 bytes [3965904, 3965910]: "_log"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4260, v438);
                                                                                                                                      if ( *v438 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v438)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v439 = sub_18000CC40(v438);
                                                                                                                                          (*(void (__fastcall **)(int *))(v439 + 48))(v438);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2781 | qword_1803B5AA8 -> novels.pipeline.pipeline_runner#883 bytes [3984503, 3984559]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984505, 3984514], "value": "payload"}, {"tag": "a", "fileRange": [3984514, 3984522], "value": "return"}], "values": [{"tag": "u", "fileRange": [3984522, 3984538], "value": "Dict[str, Any... (truncated literal)
                                                                                                                                      v440 = sub_18036FB80(a1, qword_1803B5AA8);
                                                                                                                                      v441 = (int *)sub_18012FBE0(a1, v440);
                                                                                                                                      v442 = (int *)sub_18036F0F0(a1, v441);
                                                                                                                                      if ( *v441 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v441)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v443 = sub_18000CC40(v441);
                                                                                                                                          (*(void (__fastcall **)(int *))(v443 + 48))(v441);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2793 | qword_1803B4048 -> novels.pipeline.pipeline_runner#39 bytes [3964453, 3964472]: "_normalize_kb_ids"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B4048, v442);
                                                                                                                                      if ( *v442 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v442)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v444 = sub_18000CC40(v442);
                                                                                                                                          (*(void (__fastcall **)(int *))(v444 + 48))(v442);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2803 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v445 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2804 | qword_1803B58B0 -> novels.pipeline.pipeline_runner#820 bytes [3981657, 3981726]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981659, 3981668], "value": "payload"}, {"tag": "a", "fileRange": [3981668, 3981675], "value": "model"}, {"tag": "a", "fileRange": [3981675, 3981683], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v446 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v445 != -1 )
                                                                                                                                        ++*v445;
                                                                                                                                      v447 = (int *)sub_18012FC40(a1, v445, v446);
// Original IDA line 2808 | qword_1803B5AB8 -> novels.pipeline.pipeline_runner#885 bytes [3984593, 3984614]: "init_imitation_book"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5AB8, v447);
                                                                                                                                      if ( *v447 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v447)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v448 = sub_18000CC40(v447);
                                                                                                                                          (*(void (__fastcall **)(int *))(v448 + 48))(v447);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2818 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v449 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2819 | qword_1803B58B0 -> novels.pipeline.pipeline_runner#820 bytes [3981657, 3981726]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981659, 3981668], "value": "payload"}, {"tag": "a", "fileRange": [3981668, 3981675], "value": "model"}, {"tag": "a", "fileRange": [3981675, 3981683], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v450 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v449 != -1 )
                                                                                                                                        ++*v449;
                                                                                                                                      v451 = (int *)sub_18012FCA0(a1, v449, v450);
// Original IDA line 2823 | qword_1803B5AC8 -> novels.pipeline.pipeline_runner#887 bytes [3984650, 3984668]: "init_fanfic_book"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5AC8, v451);
                                                                                                                                      if ( *v451 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v451)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v452 = sub_18000CC40(v451);
                                                                                                                                          (*(void (__fastcall **)(int *))(v452 + 48))(v451);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2833 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v453 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2834 | qword_1803B58B0 -> novels.pipeline.pipeline_runner#820 bytes [3981657, 3981726]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3981659, 3981668], "value": "payload"}, {"tag": "a", "fileRange": [3981668, 3981675], "value": "model"}, {"tag": "a", "fileRange": [3981675, 3981683], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v454 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v453 != -1 )
                                                                                                                                        ++*v453;
                                                                                                                                      v455 = (int *)sub_18012FD00(a1, v453, v454);
// Original IDA line 2838 | qword_1803B5AD8 -> novels.pipeline.pipeline_runner#889 bytes [3984701, 3984720]: "init_spinoff_book"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5AD8, v455);
                                                                                                                                      if ( *v455 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v455)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v456 = sub_18000CC40(v455);
                                                                                                                                          (*(void (__fastcall **)(int *))(v456 + 48))(v455);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2848 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v457 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2849 | qword_1803B5AE8 -> novels.pipeline.pipeline_runner#891 bytes [3984754, 3984848]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984756, 3984765], "value": "book_id"}, {"tag": "a", "fileRange": [3984765, 3984775], "value": "chapters"}, {"tag": "a", "fileRange": [3984775, 3984782], "value": "model"}, {"tag": "a", "... (truncated literal)
                                                                                                                                      v458 = sub_18036FB80(a1, qword_1803B5AE8);
                                                                                                                                      if ( *v457 != -1 )
                                                                                                                                        ++*v457;
                                                                                                                                      v459 = (int *)sub_18012FD60(a1, v457, v458);
// Original IDA line 2853 | qword_1803B5208 -> novels.pipeline.pipeline_runner#607 bytes [3976277, 3976294]: "import_chapters"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5208, v459);
                                                                                                                                      if ( *v459 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v459)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v460 = sub_18000CC40(v459);
                                                                                                                                          (*(void (__fastcall **)(int *))(v460 + 48))(v459);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2863 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v461 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2864 | qword_1803B5AF8 -> novels.pipeline.pipeline_runner#893 bytes [3984880, 3984953]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984882, 3984891], "value": "payload"}, {"tag": "a", "fileRange": [3984891, 3984898], "value": "model"}, {"tag": "a", "fileRange": [3984898, 3984906], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v462 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v461 != -1 )
                                                                                                                                        ++*v461;
                                                                                                                                      v463 = (int *)sub_18012FDC0(a1, v461, v462);
// Original IDA line 2868 | qword_1803B5B00 -> novels.pipeline.pipeline_runner#894 bytes [3984953, 3984972]: "run_short_fiction"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B00, v463);
                                                                                                                                      if ( *v463 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v463)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v464 = sub_18000CC40(v463);
                                                                                                                                          (*(void (__fastcall **)(int *))(v464 + 48))(v463);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2878 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v465 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2879 | qword_1803B5AF8 -> novels.pipeline.pipeline_runner#893 bytes [3984880, 3984953]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984882, 3984891], "value": "payload"}, {"tag": "a", "fileRange": [3984891, 3984898], "value": "model"}, {"tag": "a", "fileRange": [3984898, 3984906], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v466 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v465 != -1 )
                                                                                                                                        ++*v465;
                                                                                                                                      v467 = (int *)sub_18012FE20(a1, v465, v466);
// Original IDA line 2883 | qword_1803B5B10 -> novels.pipeline.pipeline_runner#896 bytes [3985006, 3985027]: "run_imitation_short"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B10, v467);
                                                                                                                                      if ( *v467 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v467)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v468 = sub_18000CC40(v467);
                                                                                                                                          (*(void (__fastcall **)(int *))(v468 + 48))(v467);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2893 | qword_1803B5B20 -> novels.pipeline.pipeline_runner#898 bytes [3985063, 3985104]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985065, 3985075], "value": "story_id"}, {"tag": "a", "fileRange": [3985075, 3985083], "value": "return"}], "values": [{"tag": "a", "fileRange": [3985083, 3985088], "value": "str"}, {"tag... (truncated literal)
                                                                                                                                      v469 = sub_18036FB80(a1, qword_1803B5B20);
                                                                                                                                      v470 = (int *)sub_18012FEE0(a1, v469);
// Original IDA line 2895 | qword_1803B5330 -> novels.pipeline.pipeline_runner#644 bytes [3977098, 3977123]: "get_short_story_payload"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5330, v470);
                                                                                                                                      if ( *v470 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v470)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v471 = sub_18000CC40(v470);
                                                                                                                                          (*(void (__fastcall **)(int *))(v471 + 48))(v470);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2905 | qword_1803B5B30 -> novels.pipeline.pipeline_runner#900 bytes [3985144, 3985176]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985146, 3985156], "value": "story_id"}, {"tag": "a", "fileRange": [3985156, 3985161], "value": "key"}, {"tag": "a", "fileRange": [3985161, 3985169], "value": "return"}], "values": [{"tag... (truncated literal)
                                                                                                                                      v472 = sub_18036FB80(a1, qword_1803B5B30);
                                                                                                                                      v473 = (int *)sub_18012FF40(a1, v472);
// Original IDA line 2907 | qword_1803B5340 -> novels.pipeline.pipeline_runner#646 bytes [3977148, 3977171]: "read_short_story_file"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5340, v473);
                                                                                                                                      if ( *v473 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v473)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v474 = sub_18000CC40(v473);
                                                                                                                                          (*(void (__fastcall **)(int *))(v474 + 48))(v473);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2917 | qword_1803B5B40 -> novels.pipeline.pipeline_runner#902 bytes [3985214, 3985261]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985216, 3985226], "value": "story_id"}, {"tag": "a", "fileRange": [3985226, 3985231], "value": "key"}, {"tag": "a", "fileRange": [3985231, 3985240], "value": "content"}, {"tag": "a", "fi... (truncated literal)
                                                                                                                                      v475 = sub_18036FB80(a1, qword_1803B5B40);
                                                                                                                                      v476 = (int *)sub_18012FFA0(a1, v475);
// Original IDA line 2919 | qword_1803B5350 -> novels.pipeline.pipeline_runner#648 bytes [3977196, 3977219]: "save_short_story_file"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5350, v476);
                                                                                                                                      if ( *v476 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v476)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v477 = sub_18000CC40(v476);
                                                                                                                                          (*(void (__fastcall **)(int *))(v477 + 48))(v476);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2929 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v478 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2930 | qword_1803B5AF8 -> novels.pipeline.pipeline_runner#893 bytes [3984880, 3984953]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3984882, 3984891], "value": "payload"}, {"tag": "a", "fileRange": [3984891, 3984898], "value": "model"}, {"tag": "a", "fileRange": [3984898, 3984906], "value": "return"}], "values": [{"ta... (truncated literal)
                                                                                                                                      v479 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v478 != -1 )
                                                                                                                                        ++*v478;
                                                                                                                                      v480 = (int *)sub_180130000(a1, v478, v479);
// Original IDA line 2934 | qword_1803B5B50 -> novels.pipeline.pipeline_runner#904 bytes [3985299, 3985311]: "start_play"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B50, v480);
                                                                                                                                      if ( *v480 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v480)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v481 = sub_18000CC40(v480);
                                                                                                                                          (*(void (__fastcall **)(int *))(v481 + 48))(v480);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2944 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v482 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2945 | qword_1803B5B60 -> novels.pipeline.pipeline_runner#906 bytes [3985338, 3985416]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985340, 3985350], "value": "world_id"}, {"tag": "a", "fileRange": [3985350, 3985364], "value": "player_input"}, {"tag": "a", "fileRange": [3985364, 3985371], "value": "model"}, {"tag": "... (truncated literal)
                                                                                                                                      v483 = sub_18036FB80(a1, qword_1803B5B60);
                                                                                                                                      if ( *v482 != -1 )
                                                                                                                                        ++*v482;
                                                                                                                                      v484 = (int *)sub_180130060(a1, v482, v483);
// Original IDA line 2949 | qword_1803B5B68 -> novels.pipeline.pipeline_runner#907 bytes [3985416, 3985427]: "play_step"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B68, v484);
                                                                                                                                      if ( *v484 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v484)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v485 = sub_18000CC40(v484);
                                                                                                                                          (*(void (__fastcall **)(int *))(v485 + 48))(v484);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2959 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v486 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2960 | qword_1803B5B78 -> novels.pipeline.pipeline_runner#909 bytes [3985453, 3985539]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985455, 3985465], "value": "world_id"}, {"tag": "a", "fileRange": [3985465, 3985472], "value": "patch"}, {"tag": "a", "fileRange": [3985472, 3985479], "value": "model"}, {"tag": "a", "fi... (truncated literal)
                                                                                                                                      v487 = sub_18036FB80(a1, qword_1803B5B78);
                                                                                                                                      if ( *v486 != -1 )
                                                                                                                                        ++*v486;
                                                                                                                                      v488 = (int *)sub_1801300C0(a1, v486, v487);
// Original IDA line 2964 | qword_1803B5B80 -> novels.pipeline.pipeline_runner#910 bytes [3985539, 3985550]: "play_edit"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B80, v488);
                                                                                                                                      if ( *v488 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v488)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v489 = sub_18000CC40(v488);
                                                                                                                                          (*(void (__fastcall **)(int *))(v489 + 48))(v488);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2974 | qword_1803B44E8 -> novels.pipeline.pipeline_runner#187 bytes [3967347, 3967350]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3967349, 3967350], "value": null}]}
                                                                                                                                      v490 = (_DWORD *)qword_1803B44E8;
// Original IDA line 2975 | qword_1803B5B90 -> novels.pipeline.pipeline_runner#912 bytes [3985576, 3985663]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985578, 3985588], "value": "world_id"}, {"tag": "a", "fileRange": [3985588, 3985596], "value": "params"}, {"tag": "a", "fileRange": [3985596, 3985603], "value": "model"}, {"tag": "a", "f... (truncated literal)
                                                                                                                                      v491 = sub_18036FB80(a1, qword_1803B5B90);
                                                                                                                                      if ( *v490 != -1 )
                                                                                                                                        ++*v490;
                                                                                                                                      v492 = (int *)sub_180130120(a1, v490, v491);
// Original IDA line 2979 | qword_1803B5B98 -> novels.pipeline.pipeline_runner#913 bytes [3985663, 3985676]: "play_revise"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5B98, v492);
                                                                                                                                      if ( *v492 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v492)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v493 = sub_18000CC40(v492);
                                                                                                                                          (*(void (__fastcall **)(int *))(v493 + 48))(v492);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 2989 | qword_1803B5BA8 -> novels.pipeline.pipeline_runner#915 bytes [3985704, 3985742]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985706, 3985715], "value": "book_id"}, {"tag": "a", "fileRange": [3985715, 3985722], "value": "model"}, {"tag": "a", "fileRange": [3985722, 3985739], "value": "image_generator"}], "value... (truncated literal)
                                                                                                                                      v494 = sub_18036FB80(a1, qword_1803B5BA8);
// Original IDA line 2990 | qword_1803B5BB0 -> novels.pipeline.pipeline_runner#916 bytes [3985742, 3985839]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985744, 3985753], "value": "payload"}, {"tag": "a", "fileRange": [3985753, 3985762], "value": "book_id"}, {"tag": "a", "fileRange": [3985762, 3985769], "value": "model"}, {"tag": "a", "f... (truncated literal)
                                                                                                                                      v495 = sub_18036FB80(a1, qword_1803B5BB0);
                                                                                                                                      v496 = (int *)sub_180130180(a1, v494, v495);
// Original IDA line 2992 | qword_1803B5BB8 -> novels.pipeline.pipeline_runner#917 bytes [3985839, 3985855]: "generate_cover"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5BB8, v496);
                                                                                                                                      if ( *v496 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v496)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v497 = sub_18000CC40(v496);
                                                                                                                                          (*(void (__fastcall **)(int *))(v497 + 48))(v496);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3002 | qword_1803B5BC8 -> novels.pipeline.pipeline_runner#919 bytes [3985886, 3985918]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985888, 3985904], "value": "parent_book_id"}, {"tag": "a", "fileRange": [3985904, 3985912], "value": "return"}], "values": [{"tag": "a", "fileRange": [3985912, 3985917], "value": "str"},... (truncated literal)
                                                                                                                                      v498 = sub_18036FB80(a1, qword_1803B5BC8);
                                                                                                                                      v499 = (int *)sub_1801301E0(a1, v498);
// Original IDA line 3004 | qword_1803B51B0 -> novels.pipeline.pipeline_runner#596 bytes [3975990, 3976018]: "_build_parent_canon_bundle"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B51B0, v499);
                                                                                                                                      if ( *v499 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v499)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v500 = sub_18000CC40(v499);
                                                                                                                                          (*(void (__fastcall **)(int *))(v500 + 48))(v499);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3014 | qword_1803B5BC8 -> novels.pipeline.pipeline_runner#919 bytes [3985886, 3985918]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3985888, 3985904], "value": "parent_book_id"}, {"tag": "a", "fileRange": [3985904, 3985912], "value": "return"}], "values": [{"tag": "a", "fileRange": [3985912, 3985917], "value": "str"},... (truncated literal)
                                                                                                                                      v501 = sub_18036FB80(a1, qword_1803B5BC8);
                                                                                                                                      v502 = (int *)sub_180130240(a1, v501);
// Original IDA line 3016 | qword_1803B5BD8 -> novels.pipeline.pipeline_runner#921 bytes [3985961, 3985982]: "_build_parent_canon"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5BD8, v502);
                                                                                                                                      if ( *v502 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v502)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v503 = sub_18000CC40(v502);
                                                                                                                                          (*(void (__fastcall **)(int *))(v503 + 48))(v502);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3026 | qword_1803B5BE8 -> novels.pipeline.pipeline_runner#923 bytes [3986018, 3986079]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3986020, 3986029], "value": "book_id"}, {"tag": "a", "fileRange": [3986029, 3986045], "value": "reference_text"}, {"tag": "a", "fileRange": [3986045, 3986058], "value": "source_name"}, {"... (truncated literal)
                                                                                                                                      v504 = sub_18036FB80(a1, qword_1803B5BE8);
                                                                                                                                      v505 = (int *)sub_180130300(a1, v504);
// Original IDA line 3028 | qword_1803B5050 -> novels.pipeline.pipeline_runner#552 bytes [3974792, 3974812]: "_write_style_guide"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5050, v505);
                                                                                                                                      if ( *v505 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v505)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v506 = sub_18000CC40(v505);
                                                                                                                                          (*(void (__fastcall **)(int *))(v506 + 48))(v505);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3038 | qword_1803B5BF8 -> novels.pipeline.pipeline_runner#925 bytes [3986114, 3986148]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3986116, 3986122], "value": "book"}, {"tag": "a", "fileRange": [3986122, 3986130], "value": "return"}], "values": [{"tag": "a", "fileRange": [3986130, 3986142], "value": "BookConfig"}, {"... (truncated literal)
                                                                                                                                      v507 = sub_18036FB80(a1, qword_1803B5BF8);
                                                                                                                                      v508 = (int *)sub_180130360(a1, v507);
// Original IDA line 3040 | qword_1803B5130 -> novels.pipeline.pipeline_runner#580 bytes [3975549, 3975566]: "_apply_kb_style"
                                                                                                                                      sub_18009B380(v321, *(_QWORD *)&qword_1803B5130, v508);
                                                                                                                                      if ( *v508 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v508)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v509 = sub_18000CC40(v508);
                                                                                                                                          (*(void (__fastcall **)(int *))(v509 + 48))(v508);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3050 | qword_1803B57F0 -> novels.pipeline.pipeline_runner#796 bytes [3980484, 3980500]: "PipelineRunner"
                                                                                                                                      v510 = (_DWORD *)qword_1803B57F0;
                                                                                                                                      v511 = sub_180374A40(a1, 3);
                                                                                                                                      v512 = (int *)v511;
                                                                                                                                      if ( *v510 != -1 )
                                                                                                                                        ++*v510;
                                                                                                                                      sub_18000CC80(v511, 0, v510);
                                                                                                                                      v513 = qword_1803AB240;
                                                                                                                                      if ( *(_DWORD *)qword_1803AB240 != -1 )
                                                                                                                                        ++*(_DWORD *)qword_1803AB240;
                                                                                                                                      sub_18000CC80(v512, 1, v513);
                                                                                                                                      if ( *v321 != -1 )
                                                                                                                                        ++*v321;
                                                                                                                                      sub_18000CC80(v512, 2, v321);
                                                                                                                                      *(_DWORD *)(v606 + 40) = 54;
                                                                                                                                      v514 = (_DWORD *)sub_180001790(a1, PyType_Type, v512, v319);
                                                                                                                                      if ( *v512 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v512)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v515 = sub_18000CC40(v512);
                                                                                                                                          (*(void (__fastcall **)(int *))(v515 + 48))(v512);
                                                                                                                                        }
                                                                                                                                      }
                                                                                                                                      if ( v514 == nullptr )
                                                                                                                                      {
                                                                                                                                        sub_18000CE30(a1, v605);
                                                                                                                                        v42 = v605[0];
                                                                                                                                        sub_18000CE70(v605);
                                                                                                                                        if ( (int)*v321 >= 0 )
                                                                                                                                        {
                                                                                                                                          v17 = (*(_QWORD *)v321)-- == 1;
                                                                                                                                          if ( v17 )
                                                                                                                                          {
                                                                                                                                            v516 = sub_18000CC40(v321);
                                                                                                                                            (*(void (__fastcall **)(_DWORD *))(v516 + 48))(v321);
                                                                                                                                          }
                                                                                                                                        }
                                                                                                                                        v605[0] = v42;
                                                                                                                                        sub_18000CE70(v605);
                                                                                                                                        v605[0] = v42;
                                                                                                                                        sub_18000CE70(v605);
                                                                                                                                        if ( *v319 >= 0 )
                                                                                                                                        {
                                                                                                                                          v17 = (*(_QWORD *)v319)-- == 1;
                                                                                                                                          if ( v17 )
                                                                                                                                          {
                                                                                                                                            v517 = sub_18000CC40(v319);
                                                                                                                                            (*(void (__fastcall **)(int *))(v517 + 48))(v319);
                                                                                                                                          }
                                                                                                                                        }
                                                                                                                                        if ( (int)*v321 >= 0 )
                                                                                                                                        {
                                                                                                                                          v17 = (*(_QWORD *)v321)-- == 1;
                                                                                                                                          if ( v17 )
                                                                                                                                          {
                                                                                                                                            v518 = sub_18000CC40(v321);
                                                                                                                                            (*(void (__fastcall **)(_DWORD *))(v518 + 48))(v321);
                                                                                                                                          }
                                                                                                                                        }
                                                                                                                                        v18 = v606;
                                                                                                                                        v22 = 54;
                                                                                                                                        goto LABEL_714;
                                                                                                                                      }
                                                                                                                                      if ( *v514 != -1 )
                                                                                                                                        ++*v514;
                                                                                                                                      if ( (int)*v321 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v321)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v522 = sub_18000CC40(v321);
                                                                                                                                          (*(void (__fastcall **)(_DWORD *))(v522 + 48))(v321);
                                                                                                                                        }
                                                                                                                                      }
                                                                                                                                      if ( (int)*v514 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v514)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v523 = sub_18000CC40(v514);
                                                                                                                                          (*(void (__fastcall **)(_DWORD *))(v523 + 48))(v514);
                                                                                                                                        }
                                                                                                                                      }
// Original IDA line 3134 | qword_1803B57F0 -> novels.pipeline.pipeline_runner#796 bytes [3980484, 3980500]: "PipelineRunner"
                                                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B57F0, v514);
                                                                                                                                      sub_180001D80(a1, v524);
                                                                                                                                      if ( *v319 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v319)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v525 = sub_18000CC40(v319);
                                                                                                                                          (*(void (__fastcall **)(int *))(v525 + 48))(v319);
                                                                                                                                        }
                                                                                                                                      }
                                                                                                                                      if ( (int)*v321 >= 0 )
                                                                                                                                      {
                                                                                                                                        v17 = (*(_QWORD *)v321)-- == 1;
                                                                                                                                        if ( v17 )
                                                                                                                                        {
                                                                                                                                          v526 = sub_18000CC40(v321);
                                                                                                                                          (*(void (__fastcall **)(_DWORD *))(v526 + 48))(v321);
                                                                                                                                        }
                                                                                                                                      }
                                                                                                                                      result = qword_1803CA778;
                                                                                                                                      if ( *(_DWORD *)qword_1803CA778 != -1 )
                                                                                                                                        ++*(_DWORD *)qword_1803CA778;
                                                                                                                                      return result;
                                                                                                                                    }
                                                                                                                                  }
                                                                                                                                  sub_18000CE30(a1, v605);
                                                                                                                                  v22 = 51;
                                                                                                                                  goto LABEL_715;
                                                                                                                                }
                                                                                                                              }
                                                                                                                              sub_18000CE30(a1, v605);
                                                                                                                              v22 = 50;
                                                                                                                              goto LABEL_715;
                                                                                                                            }
                                                                                                                          }
                                                                                                                          sub_18000CE30(a1, v605);
                                                                                                                          v22 = 48;
                                                                                                                          goto LABEL_715;
                                                                                                                        }
                                                                                                                      }
                                                                                                                      sub_18000CE30(a1, v605);
                                                                                                                      v22 = 47;
                                                                                                                      goto LABEL_715;
                                                                                                                    }
                                                                                                                  }
                                                                                                                  sub_18000CE30(a1, v605);
                                                                                                                  v22 = 46;
                                                                                                                  goto LABEL_715;
                                                                                                                }
                                                                                                              }
                                                                                                              sub_18000CE30(a1, v605);
                                                                                                              v22 = 45;
                                                                                                              goto LABEL_715;
                                                                                                            }
                                                                                                          }
                                                                                                          sub_18000CE30(a1, v605);
                                                                                                          v22 = 44;
                                                                                                          goto LABEL_715;
                                                                                                        }
                                                                                                      }
                                                                                                      sub_18000CE30(a1, v605);
                                                                                                      v22 = 43;
                                                                                                      goto LABEL_715;
                                                                                                    }
                                                                                                  }
                                                                                                  sub_18000CE30(
                                                                                                    a1,
                                                                                                    v605);
                                                                                                  v22 = 42;
                                                                                                  goto LABEL_715;
                                                                                                }
                                                                                              }
                                                                                              sub_18000CE30(a1, v605);
                                                                                              v22 = 41;
                                                                                              goto LABEL_715;
                                                                                            }
                                                                                          }
                                                                                          sub_18000CE30(a1, v605);
                                                                                          v22 = 39;
                                                                                          goto LABEL_715;
                                                                                        }
                                                                                      }
                                                                                      sub_18000CE30(a1, v605);
                                                                                      v22 = 37;
                                                                                      goto LABEL_715;
                                                                                    }
                                                                                  }
                                                                                  sub_18000CE30(a1, v605);
                                                                                  v22 = 36;
                                                                                  goto LABEL_715;
                                                                                }
                                                                              }
                                                                              sub_18000CE30(a1, v605);
                                                                              v22 = 35;
                                                                              goto LABEL_715;
                                                                            }
                                                                          }
                                                                          sub_18000CE30(a1, v605);
                                                                          v22 = 34;
                                                                          goto LABEL_715;
                                                                        }
                                                                      }
                                                                      sub_18000CE30(a1, v605);
                                                                      v22 = 32;
                                                                      goto LABEL_715;
                                                                    }
                                                                  }
                                                                  sub_18000CE30(a1, v605);
                                                                  v22 = 30;
                                                                  goto LABEL_715;
                                                                }
                                                              }
                                                              sub_18000CE30(a1, v605);
                                                              v22 = 29;
                                                              goto LABEL_715;
                                                            }
                                                          }
                                                          sub_18000CE30(a1, v605);
                                                          v22 = 28;
                                                          goto LABEL_715;
                                                        }
                                                      }
                                                      sub_18000CE30(a1, v605);
                                                      v22 = 20;
                                                      goto LABEL_715;
                                                    }
                                                  }
                                                  sub_18000CE30(a1, v605);
                                                  v22 = 19;
                                                  goto LABEL_715;
                                                }
                                              }
                                              sub_18000CE30(a1, v605);
                                              v22 = 18;
                                              goto LABEL_715;
                                            }
                                          }
                                          sub_18000CE30(a1, v605);
                                          v22 = 17;
                                          goto LABEL_715;
                                        }
                                      }
                                      sub_18000CE30(a1, v605);
                                      v22 = 16;
                                      goto LABEL_715;
                                    }
                                  }
                                  sub_18000CE30(a1, v605);
                                  v22 = 15;
                                  goto LABEL_715;
                                }
                              }
                              sub_18000CE30(a1, v605);
                              v22 = 14;
                              goto LABEL_715;
                            }
                          }
                          sub_18000CE30(a1, v605);
                          v22 = 13;
                          goto LABEL_715;
                        }
                      }
                      sub_18000CE30(a1, v605);
                      v22 = 12;
                      goto LABEL_715;
                    }
                  }
                  sub_18000CE30(a1, v605);
                  v22 = 11;
                  goto LABEL_715;
                }
              }
              sub_18000CE30(a1, v605);
              v22 = 10;
              goto LABEL_715;
            }
          }
          sub_18000CE30(a1, v605);
          v22 = 9;
          goto LABEL_715;
        }
      }
      sub_18000CE30(a1, v605);
      v22 = 8;
      goto LABEL_715;
    }
  }
  v22 = 5;
  v605[0] = *(_QWORD *)(a1 + 96);
  *(_QWORD *)(a1 + 96) = 0;
  goto LABEL_715;
}
