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
  sub_1800DE740(qword_1803CA780, qword_1803B54E8, Py_NoneStruct);
  sub_1800DE740(qword_1803CA780, qword_1803B54F0, qword_1803B3F00);
  v606 = sub_18036CCF0(qword_1803B3BA8, qword_1803CA778);
  v18 = v606;
  sub_180001D00(a1, v606);
  v19 = qword_1803B3F00;
  v20 = sub_1800DFBA0(a1);
  if ( (unsigned __int8)sub_18037F3A0(a1, v20, qword_1803B54F8, v19) == 0
    || (v23 = sub_1800DFBA0(a1), (unsigned __int8)sub_18037F3A0(a1, v23, qword_1803B5500, Py_TrueStruct) == 0) )
  {
    v22 = 1;
    v605[0] = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_715;
  }
  sub_1800DE740(qword_1803CA780, qword_1803B5508, Py_NoneStruct);
  v24 = sub_180372070();
  v25 = sub_18037F0C0(a1, v24, qword_1803B5510);
  sub_1800DE7D0(qword_1803CA780, qword_1803B5510, v25);
  v26 = qword_1803AB250;
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
  sub_1800DE7D0(qword_1803CA780, qword_1803B5518, v28);
  v29 = qword_1803AB250;
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
  sub_1800DE7D0(qword_1803CA780, qword_1803B4EA0, v31);
  v32 = qword_1803B5528;
  v33 = qword_1803AB250;
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
      v37 = sub_180371D80(a1, (_DWORD)v36, qword_1803CA780, qword_1803B5530, qword_1803AB250);
    else
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
      sub_1800DE7D0(qword_1803CA780, qword_1803B5530, v38);
      v39 = (int *)sub_180372110();
      v40 = v39;
      if ( *v39 != -1 )
        ++*v39;
      if ( (unsigned int)sub_18000CC50(v39, PyModule_Type) != 0 )
        v41 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5538, qword_1803AB250);
      else
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
      sub_1800DE7D0(qword_1803CA780, qword_1803B5538, v41);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
        v43 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5540, qword_1803AB250);
      else
        v43 = sub_180371890(a1, v40, qword_1803B5540);
      if ( v43 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
      sub_1800DE7D0(qword_1803CA780, qword_1803B5540, v43);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
        v44 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5548, qword_1803AB250);
      else
        v44 = sub_180371890(a1, v40, qword_1803B5548);
      if ( v44 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
      sub_1800DE7D0(qword_1803CA780, qword_1803B5548, v44);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
        v45 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5550, qword_1803AB250);
      else
        v45 = sub_180371890(a1, v40, qword_1803B5550);
      if ( v45 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_68;
      }
      sub_1800DE7D0(qword_1803CA780, qword_1803B5550, v45);
      if ( (unsigned int)sub_18000CC50(v40, PyModule_Type) != 0 )
        v46 = sub_180371D80(a1, (_DWORD)v40, qword_1803CA780, qword_1803B5558, qword_1803AB250);
      else
        v46 = sub_180371890(a1, v40, qword_1803B5558);
      if ( v46 == 0 )
      {
        sub_18000CE30(a1, v605);
        v42 = v605[0];
        goto LABEL_68;
      }
      sub_1800DE7D0(qword_1803CA780, qword_1803B5558, v46);
      if ( *v40 >= 0 )
      {
        v17 = (*(_QWORD *)v40)-- == 1;
        if ( v17 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v40 + 1) + 48LL))(v40);
      }
      v47 = qword_1803B5560;
      v48 = qword_1803CA780;
      v566 = qword_1803B4258;
      v527 = qword_1803B5568;
      *(_DWORD *)(v18 + 40) = 8;
      v49 = sub_1803717E0(a1, v47, v48, Py_NoneStruct, v527, v566);
      v50 = (int *)v49;
      if ( v49 != 0 )
      {
        if ( (unsigned int)sub_18000CC50(v49, PyModule_Type) != 0 )
          v51 = sub_180371D80(a1, (_DWORD)v50, qword_1803CA780, qword_1803B40D8, qword_1803AB250);
        else
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
          sub_1800DE7D0(qword_1803CA780, qword_1803B40D8, v52);
          v53 = qword_1803B5570;
          v54 = qword_1803CA780;
          v567 = qword_1803B4258;
          v528 = qword_1803B5578;
          *(_DWORD *)(v18 + 40) = 9;
          v55 = sub_1803717E0(a1, v53, v54, Py_NoneStruct, v528, v567);
          v56 = (int *)v55;
          if ( v55 != 0 )
          {
            if ( (unsigned int)sub_18000CC50(v55, PyModule_Type) != 0 )
              v57 = sub_180371D80(a1, (_DWORD)v56, qword_1803CA780, qword_1803B40E8, qword_1803AB250);
            else
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
              sub_1800DE7D0(qword_1803CA780, qword_1803B40E8, v58);
              v60 = qword_1803B5580;
              v61 = qword_1803CA780;
              v568 = qword_1803B4258;
              v529 = qword_1803B5588;
              *(_DWORD *)(v18 + 40) = 10;
              v62 = sub_1803717E0(a1, v60, v61, Py_NoneStruct, v529, v568);
              v63 = (int *)v62;
              if ( v62 != 0 )
              {
                if ( (unsigned int)sub_18000CC50(v62, PyModule_Type) != 0 )
                  v64 = sub_180371D80(a1, (_DWORD)v63, qword_1803CA780, qword_1803B43F8, qword_1803AB250);
                else
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
                  sub_1800DE7D0(qword_1803CA780, qword_1803B43F8, v65);
                  v67 = qword_1803B5590;
                  v68 = qword_1803CA780;
                  v569 = qword_1803B4258;
                  v530 = qword_1803B5598;
                  *(_DWORD *)(v18 + 40) = 11;
                  v69 = sub_1803717E0(a1, v67, v68, Py_NoneStruct, v530, v569);
                  v70 = (int *)v69;
                  if ( v69 != 0 )
                  {
                    if ( (unsigned int)sub_18000CC50(v69, PyModule_Type) != 0 )
                      v71 = sub_180371D80(a1, (_DWORD)v70, qword_1803CA780, qword_1803B3F78, qword_1803AB250);
                    else
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
                      sub_1800DE7D0(qword_1803CA780, qword_1803B3F78, v72);
                      v74 = qword_1803B55A0;
                      v75 = qword_1803CA780;
                      v570 = qword_1803B4258;
                      v531 = qword_1803B55A8;
                      *(_DWORD *)(v18 + 40) = 12;
                      v76 = sub_1803717E0(a1, v74, v75, Py_NoneStruct, v531, v570);
                      v77 = (int *)v76;
                      if ( v76 != 0 )
                      {
                        if ( (unsigned int)sub_18000CC50(v76, PyModule_Type) != 0 )
                          v78 = sub_180371D80(a1, (_DWORD)v77, qword_1803CA780, qword_1803B42E8, qword_1803AB250);
                        else
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
                          sub_1800DE7D0(qword_1803CA780, qword_1803B42E8, v79);
                          v81 = qword_1803B55B0;
                          v82 = qword_1803CA780;
                          v571 = qword_1803B4258;
                          v532 = qword_1803B55B8;
                          *(_DWORD *)(v18 + 40) = 13;
                          v83 = sub_1803717E0(a1, v81, v82, Py_NoneStruct, v532, v571);
                          v84 = (int *)v83;
                          if ( v83 != 0 )
                          {
                            if ( (unsigned int)sub_18000CC50(v83, PyModule_Type) != 0 )
                              v85 = sub_180371D80(a1, (_DWORD)v84, qword_1803CA780, qword_1803B4D80, qword_1803AB250);
                            else
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
                              sub_1800DE7D0(qword_1803CA780, qword_1803B4D80, v86);
                              v88 = qword_1803B55C0;
                              v89 = qword_1803CA780;
                              v572 = qword_1803B4258;
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
                                          qword_1803B42B8,
                                          qword_1803AB250);
                                else
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
                                  sub_1800DE7D0(qword_1803CA780, qword_1803B42B8, v93);
                                  v95 = qword_1803B55D0;
                                  v96 = qword_1803CA780;
                                  v573 = qword_1803B4258;
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
                                              qword_1803B4408,
                                              qword_1803AB250);
                                    else
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
                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4408, v100);
                                      v102 = qword_1803B55E0;
                                      v103 = qword_1803CA780;
                                      v574 = qword_1803B4258;
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
                                                   qword_1803B4B88,
                                                   qword_1803AB250);
                                        else
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
                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4B88, v107);
                                          v109 = qword_1803B55F0;
                                          v110 = qword_1803CA780;
                                          v575 = qword_1803B4258;
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
                                                       qword_1803B4BA0,
                                                       qword_1803AB250);
                                            else
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
                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4BA0, v114);
                                              v116 = qword_1803B5600;
                                              v117 = qword_1803CA780;
                                              v576 = qword_1803B4258;
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
                                                           qword_1803B4318,
                                                           qword_1803AB250);
                                                else
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
                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4318, v121);
                                                  v123 = qword_1803B5610;
                                                  v124 = qword_1803CA780;
                                                  v577 = qword_1803B4258;
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
                                                               qword_1803B4308,
                                                               qword_1803AB250);
                                                    else
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
                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4308, v128);
                                                      v130 = qword_1803B5620;
                                                      v131 = qword_1803CA780;
                                                      v578 = qword_1803B4258;
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
                                                                   qword_1803B4160,
                                                                   qword_1803AB250);
                                                        else
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
                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B4160, v135);
                                                          v137 = qword_1803B5630;
                                                          v138 = qword_1803CA780;
                                                          v579 = qword_1803B4258;
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
                                                                     qword_1803B4D10,
                                                                     qword_1803AB250);
                                                          else
                                                            v141 = sub_180371890(a1, v140, qword_1803B4D10);
                                                          if ( v141 == 0
                                                            || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4D10, v141),
                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                              ? (v142 = sub_180371890(a1, v140, qword_1803B4A90))
                                                              : (v142 = sub_180371D80(
                                                                          a1,
                                                                          (_DWORD)v140,
                                                                          qword_1803CA780,
                                                                          qword_1803B4A90,
                                                                          qword_1803AB250)),
                                                                v142 == 0
                                                             || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4A90, v142),
                                                                  (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                               ? (v143 = sub_180371890(a1, v140, qword_1803B4A30))
                                                               : (v143 = sub_180371D80(
                                                                           a1,
                                                                           (_DWORD)v140,
                                                                           qword_1803CA780,
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
                                                          v146 = qword_1803B5640;
                                                          v147 = qword_1803CA780;
                                                          v580 = qword_1803B4258;
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
                                                                     qword_1803B4AD0,
                                                                     qword_1803AB250);
                                                          else
                                                            v149 = sub_180371890(a1, v140, qword_1803B4AD0);
                                                          if ( v149 == 0
                                                            || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4AD0, v149),
                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                              ? (v150 = sub_180371890(a1, v140, qword_1803B4650))
                                                              : (v150 = sub_180371D80(
                                                                          a1,
                                                                          (_DWORD)v140,
                                                                          qword_1803CA780,
                                                                          qword_1803B4650,
                                                                          qword_1803AB250)),
                                                                v150 == 0
                                                             || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4650, v150),
                                                                  (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                               ? (v151 = sub_180371890(a1, v140, qword_1803B4648))
                                                               : (v151 = sub_180371D80(
                                                                           a1,
                                                                           (_DWORD)v140,
                                                                           qword_1803CA780,
                                                                           qword_1803B4648,
                                                                           qword_1803AB250)),
                                                                 v151 == 0
                                                              || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4648, v151),
                                                                   (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                                ? (v152 = sub_180371890(a1, v140, qword_1803B4588))
                                                                : (v152 = sub_180371D80(
                                                                            a1,
                                                                            (_DWORD)v140,
                                                                            qword_1803CA780,
                                                                            qword_1803B4588,
                                                                            qword_1803AB250)),
                                                                  v152 == 0))) )
                                                          {
                                                            v22 = 22;
                                                            goto LABEL_206;
                                                          }
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
                                                          v154 = qword_1803B5650;
                                                          v155 = qword_1803CA780;
                                                          v581 = qword_1803B4258;
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
                                                                       qword_1803B4198,
                                                                       qword_1803AB250);
                                                            else
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
                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4198, v159);
                                                              v161 = qword_1803B5660;
                                                              v162 = qword_1803CA780;
                                                              v582 = qword_1803B4258;
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
                                                                           qword_1803B4D70,
                                                                           qword_1803AB250);
                                                                else
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
                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4D70, v166);
                                                                  v168 = qword_1803B5670;
                                                                  v169 = qword_1803CA780;
                                                                  v583 = qword_1803B4258;
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
                                                                               qword_1803B4C08,
                                                                               qword_1803AB250);
                                                                    else
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
                                                                        qword_1803B4C08,
                                                                        v173);
                                                                      v175 = qword_1803B5680;
                                                                      v176 = qword_1803CA780;
                                                                      v584 = qword_1803B4258;
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
                                                                                 qword_1803B4CF8,
                                                                                 qword_1803AB250);
                                                                      else
                                                                        v178 = sub_180371890(a1, v140, qword_1803B4CF8);
                                                                      if ( v178 == 0
                                                                        || ((sub_1800DE7D0(
                                                                               qword_1803CA780,
                                                                               qword_1803B4CF8,
                                                                               v178),
                                                                             (unsigned int)sub_18000CC50(
                                                                                             v140,
                                                                                             PyModule_Type) == 0)
                                                                          ? (v179 = sub_180371890(
                                                                                      a1,
                                                                                      v140,
                                                                                      qword_1803B4B80))
                                                                          : (v179 = sub_180371D80(
                                                                                      a1,
                                                                                      (_DWORD)v140,
                                                                                      qword_1803CA780,
                                                                                      qword_1803B4B80,
                                                                                      qword_1803AB250)),
                                                                            v179 == 0
                                                                         || ((sub_1800DE7D0(
                                                                                qword_1803CA780,
                                                                                qword_1803B4B80,
                                                                                v179),
                                                                              (unsigned int)sub_18000CC50(
                                                                                              v140,
                                                                                              PyModule_Type) == 0)
                                                                           ? (v180 = sub_180371890(
                                                                                       a1,
                                                                                       v140,
                                                                                       qword_1803B4BF0))
                                                                           : (v180 = sub_180371D80(
                                                                                       a1,
                                                                                       (_DWORD)v140,
                                                                                       qword_1803CA780,
                                                                                       qword_1803B4BF0,
                                                                                       qword_1803AB250)),
                                                                             v180 == 0)) )
                                                                      {
                                                                        v22 = 31;
                                                                        goto LABEL_206;
                                                                      }
                                                                      sub_1800DE7D0(
                                                                        qword_1803CA780,
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
                                                                      v182 = qword_1803B5690;
                                                                      v183 = qword_1803CA780;
                                                                      v585 = qword_1803B4258;
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
                                                                                   qword_1803B4D50,
                                                                                   qword_1803AB250);
                                                                        else
                                                                          v186 = sub_180371890(
                                                                                   a1,
                                                                                   v185,
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
                                                                            qword_1803B4D50,
                                                                            v187);
                                                                          v189 = qword_1803B56A0;
                                                                          v190 = qword_1803CA780;
                                                                          v586 = qword_1803B4258;
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
                                                                                     qword_1803B4010,
                                                                                     qword_1803AB250);
                                                                          else
                                                                            v192 = sub_180371890(
                                                                                     a1,
                                                                                     v140,
                                                                                     qword_1803B4010);
                                                                          if ( v192 == 0
                                                                            || ((sub_1800DE7D0(
                                                                                   qword_1803CA780,
                                                                                   qword_1803B4010,
                                                                                   v192),
                                                                                 (unsigned int)sub_18000CC50(
                                                                                                 v140,
                                                                                                 PyModule_Type) == 0)
                                                                              ? (v193 = sub_180371890(
                                                                                          a1,
                                                                                          v140,
                                                                                          qword_1803B45C0))
                                                                              : (v193 = sub_180371D80(
                                                                                          a1,
                                                                                          (_DWORD)v140,
                                                                                          qword_1803CA780,
                                                                                          qword_1803B45C0,
                                                                                          qword_1803AB250)),
                                                                                v193 == 0
                                                                             || ((sub_1800DE7D0(
                                                                                    qword_1803CA780,
                                                                                    qword_1803B45C0,
                                                                                    v193),
                                                                                  (unsigned int)sub_18000CC50(
                                                                                                  v140,
                                                                                                  PyModule_Type) == 0)
                                                                               ? (v194 = sub_180371890(
                                                                                           a1,
                                                                                           v140,
                                                                                           qword_1803B45D0))
                                                                               : (v194 = sub_180371D80(
                                                                                           a1,
                                                                                           (_DWORD)v140,
                                                                                           qword_1803CA780,
                                                                                           qword_1803B45D0,
                                                                                           qword_1803AB250)),
                                                                                 v194 == 0)) )
                                                                          {
                                                                            v22 = 33;
                                                                            goto LABEL_206;
                                                                          }
                                                                          sub_1800DE7D0(
                                                                            qword_1803CA780,
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
                                                                          v196 = qword_1803B56B0;
                                                                          v197 = qword_1803CA780;
                                                                          v587 = qword_1803B4258;
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
                                                                                       qword_1803B41B8,
                                                                                       qword_1803AB250);
                                                                            else
                                                                              v200 = sub_180371890(
                                                                                       a1,
                                                                                       v199,
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
                                                                                qword_1803B41B8,
                                                                                v201);
                                                                              v203 = qword_1803B56C0;
                                                                              v204 = qword_1803CA780;
                                                                              v588 = qword_1803B4258;
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
                                                                                           qword_1803B56D0,
                                                                                           qword_1803AB250);
                                                                                else
                                                                                  v207 = sub_180371890(
                                                                                           a1,
                                                                                           v206,
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
                                                                                    qword_1803B56D0,
                                                                                    v208);
                                                                                  v210 = qword_1803B56D8;
                                                                                  v211 = qword_1803CA780;
                                                                                  v589 = qword_1803B4258;
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
                                                                                               qword_1803B3F80,
                                                                                               qword_1803AB250);
                                                                                    else
                                                                                      v214 = sub_180371890(
                                                                                               a1,
                                                                                               v213,
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
                                                                                        qword_1803B3F80,
                                                                                        v215);
                                                                                      v217 = qword_1803B56E8;
                                                                                      v218 = qword_1803CA780;
                                                                                      v590 = qword_1803B4258;
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
                                                                                                   qword_1803B40C0,
                                                                                                   qword_1803AB250);
                                                                                        else
                                                                                          v221 = sub_180371890(
                                                                                                   a1,
                                                                                                   v220,
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
                                                                                            qword_1803B40C0,
                                                                                            v222);
                                                                                          v224 = qword_1803B56F8;
                                                                                          v225 = qword_1803CA780;
                                                                                          v591 = qword_1803B4258;
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
                                                                                                     qword_1803B4108,
                                                                                                     qword_1803AB250);
                                                                                          else
                                                                                            v227 = sub_180371890(
                                                                                                     a1,
                                                                                                     v140,
                                                                                                     qword_1803B4108);
                                                                                          if ( v227 == 0
                                                                                            || ((sub_1800DE7D0(
                                                                                                   qword_1803CA780,
                                                                                                   qword_1803B4108,
                                                                                                   v227),
                                                                                                 (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0)
                                                                                              ? (v228 = sub_180371890(a1, v140, qword_1803B4118))
                                                                                              : (v228 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4118, qword_1803AB250)),
                                                                                                v228 == 0) )
                                                                                          {
                                                                                            v22 = 38;
                                                                                            goto LABEL_206;
                                                                                          }
                                                                                          sub_1800DE7D0(
                                                                                            qword_1803CA780,
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
                                                                                          v230 = qword_1803B5708;
                                                                                          v231 = qword_1803CA780;
                                                                                          v592 = qword_1803B4258;
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
                                                                                              v234 = sub_180371D80(a1, (_DWORD)v233, qword_1803CA780, qword_1803B3F58, qword_1803AB250);
                                                                                            else
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
                                                                                                qword_1803B3F58,
                                                                                                v235);
                                                                                              v237 = qword_1803B5718;
                                                                                              v238 = qword_1803CA780;
                                                                                              v593 = qword_1803B4258;
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
                                                                                                v240 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B54B0, qword_1803AB250);
                                                                                              else
                                                                                                v240 = sub_180371890(a1, v140, qword_1803B54B0);
                                                                                              if ( v240 == 0
                                                                                                || ((sub_1800DE7D0(qword_1803CA780, qword_1803B54B0, v240), (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0) ? (v241 = sub_180371890(a1, v140, qword_1803B54A8)) : (v241 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B54A8, qword_1803AB250)),
                                                                                                    v241 == 0) )
                                                                                              {
                                                                                                v22 = 40;
                                                                                                goto LABEL_206;
                                                                                              }
                                                                                              sub_1800DE7D0(
                                                                                                qword_1803CA780,
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
                                                                                              v243 = qword_1803B5228;
                                                                                              v244 = qword_1803CA780;
                                                                                              v594 = qword_1803B4258;
                                                                                              v555 = qword_1803B5230;
                                                                                              *(_DWORD *)(v18 + 40) = 41;
                                                                                              v245 = sub_1803717E0(a1, v243, v244, Py_NoneStruct, v555, v594);
                                                                                              v246 = (int *)v245;
                                                                                              if ( v245 != 0 )
                                                                                              {
                                                                                                if ( (unsigned int)sub_18000CC50(v245, PyModule_Type) != 0 )
                                                                                                  v247 = sub_180371D80(a1, (_DWORD)v246, qword_1803CA780, qword_1803B5060, qword_1803AB250);
                                                                                                else
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
                                                                                                    qword_1803B5060,
                                                                                                    v248);
                                                                                                  v250 = qword_1803B5728;
                                                                                                  v251 = qword_1803CA780;
                                                                                                  v595 = qword_1803B4258;
                                                                                                  v556 = qword_1803B5730;
                                                                                                  *(_DWORD *)(v18 + 40) = 42;
                                                                                                  v252 = sub_1803717E0(a1, v250, v251, Py_NoneStruct, v556, v595);
                                                                                                  v253 = (int *)v252;
                                                                                                  if ( v252 != 0 )
                                                                                                  {
                                                                                                    if ( (unsigned int)sub_18000CC50(v252, PyModule_Type) != 0 )
                                                                                                      v254 = sub_180371D80(a1, (_DWORD)v253, qword_1803CA780, qword_1803B5738, qword_1803AB250);
                                                                                                    else
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
                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B46D0, v255);
                                                                                                      v257 = qword_1803B5740;
                                                                                                      v258 = qword_1803CA780;
                                                                                                      v596 = qword_1803B4258;
                                                                                                      v557 = qword_1803B5748;
                                                                                                      *(_DWORD *)(v18 + 40) = 43;
                                                                                                      v259 = sub_1803717E0(a1, v257, v258, Py_NoneStruct, v557, v596);
                                                                                                      v260 = (int *)v259;
                                                                                                      if ( v259 != 0 )
                                                                                                      {
                                                                                                        if ( (unsigned int)sub_18000CC50(v259, PyModule_Type) != 0 )
                                                                                                          v261 = sub_180371D80(a1, (_DWORD)v260, qword_1803CA780, qword_1803B5750, qword_1803AB250);
                                                                                                        else
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
                                                                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B46F0, v262);
                                                                                                          v264 = qword_1803B5758;
                                                                                                          v265 = qword_1803CA780;
                                                                                                          v597 = qword_1803B4258;
                                                                                                          v558 = qword_1803B5760;
                                                                                                          *(_DWORD *)(v18 + 40) = 44;
                                                                                                          v266 = sub_1803717E0(a1, v264, v265, Py_NoneStruct, v558, v597);
                                                                                                          v267 = (int *)v266;
                                                                                                          if ( v266 != 0 )
                                                                                                          {
                                                                                                            if ( (unsigned int)sub_18000CC50(v266, PyModule_Type) != 0 )
                                                                                                              v268 = sub_180371D80(a1, (_DWORD)v267, qword_1803CA780, qword_1803B5768, qword_1803AB250);
                                                                                                            else
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
                                                                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B4708, v269);
                                                                                                              v271 = qword_1803B5770;
                                                                                                              v272 = qword_1803CA780;
                                                                                                              v598 = qword_1803B4258;
                                                                                                              v559 = qword_1803B5778;
                                                                                                              *(_DWORD *)(v18 + 40) = 45;
                                                                                                              v273 = sub_1803717E0(a1, v271, v272, Py_NoneStruct, v559, v598);
                                                                                                              v274 = (int *)v273;
                                                                                                              if ( v273 != 0 )
                                                                                                              {
                                                                                                                if ( (unsigned int)sub_18000CC50(v273, PyModule_Type) != 0 )
                                                                                                                  v275 = sub_180371D80(a1, (_DWORD)v274, qword_1803CA780, qword_1803B5780, qword_1803AB250);
                                                                                                                else
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
                                                                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B4378, v276);
                                                                                                                  v278 = qword_1803B5788;
                                                                                                                  v279 = qword_1803CA780;
                                                                                                                  v599 = qword_1803B4258;
                                                                                                                  v560 = qword_1803B5790;
                                                                                                                  *(_DWORD *)(v18 + 40) = 46;
                                                                                                                  v280 = sub_1803717E0(a1, v278, v279, Py_NoneStruct, v560, v599);
                                                                                                                  v281 = (int *)v280;
                                                                                                                  if ( v280 != 0 )
                                                                                                                  {
                                                                                                                    if ( (unsigned int)sub_18000CC50(v280, PyModule_Type) != 0 )
                                                                                                                      v282 = sub_180371D80(a1, (_DWORD)v281, qword_1803CA780, qword_1803B4490, qword_1803AB250);
                                                                                                                    else
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
                                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B4490, v283);
                                                                                                                      v285 = qword_1803B5798;
                                                                                                                      v286 = qword_1803CA780;
                                                                                                                      v600 = qword_1803AB258;
                                                                                                                      v561 = qword_1803B57A0;
                                                                                                                      *(_DWORD *)(v18 + 40) = 47;
                                                                                                                      v287 = sub_1803717E0(a1, v285, v286, Py_NoneStruct, v561, v600);
                                                                                                                      v288 = (int *)v287;
                                                                                                                      if ( v287 != 0 )
                                                                                                                      {
                                                                                                                        if ( (unsigned int)sub_18000CC50(v287, PyModule_Type) != 0 )
                                                                                                                          v289 = sub_180371D80(a1, (_DWORD)v288, qword_1803CA780, qword_1803B44E0, qword_1803AB250);
                                                                                                                        else
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
                                                                                                                          sub_1800DE7D0(qword_1803CA780, qword_1803B44E0, v290);
                                                                                                                          v292 = qword_1803B57A8;
                                                                                                                          v293 = qword_1803CA780;
                                                                                                                          v601 = qword_1803AB258;
                                                                                                                          v562 = qword_1803B57B0;
                                                                                                                          *(_DWORD *)(v18 + 40) = 48;
                                                                                                                          v294 = sub_1803717E0(a1, v292, v293, Py_NoneStruct, v562, v601);
                                                                                                                          v295 = (int *)v294;
                                                                                                                          if ( v294 != 0 )
                                                                                                                          {
                                                                                                                            if ( (unsigned int)sub_18000CC50(v294, PyModule_Type) != 0 )
                                                                                                                              v296 = sub_180371D80(a1, (_DWORD)v295, qword_1803CA780, qword_1803B5460, qword_1803AB250);
                                                                                                                            else
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
                                                                                                                              sub_1800DE7D0(qword_1803CA780, qword_1803B5460, v297);
                                                                                                                              v299 = qword_1803B57B8;
                                                                                                                              v300 = qword_1803CA780;
                                                                                                                              v602 = qword_1803AB258;
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
                                                                                                                                v302 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4228, qword_1803AB250);
                                                                                                                              else
                                                                                                                                v302 = sub_180371890(a1, v140, qword_1803B4228);
                                                                                                                              if ( v302 == 0 || ((sub_1800DE7D0(qword_1803CA780, qword_1803B4228, v302), (unsigned int)sub_18000CC50(v140, PyModule_Type) == 0) ? (v303 = sub_180371890(a1, v140, qword_1803B4220)) : (v303 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA780, qword_1803B4220, qword_1803AB250)), v303 == 0) )
                                                                                                                              {
                                                                                                                                sub_18000CE30(a1, v605);
                                                                                                                                v22 = 49;
                                                                                                                                goto LABEL_207;
                                                                                                                              }
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
                                                                                                                              v305 = qword_1803B57C8;
                                                                                                                              v306 = qword_1803CA780;
                                                                                                                              v603 = qword_1803AB258;
                                                                                                                              v564 = qword_1803B57D0;
                                                                                                                              *(_DWORD *)(v18 + 40) = 50;
                                                                                                                              v307 = sub_1803717E0(a1, v305, v306, Py_NoneStruct, v564, v603);
                                                                                                                              v308 = (int *)v307;
                                                                                                                              if ( v307 != 0 )
                                                                                                                              {
                                                                                                                                if ( (unsigned int)sub_18000CC50(v307, PyModule_Type) != 0 )
                                                                                                                                  v309 = sub_180371D80(a1, (_DWORD)v308, qword_1803CA780, qword_1803B5358, qword_1803AB250);
                                                                                                                                else
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
                                                                                                                                  sub_1800DE7D0(qword_1803CA780, qword_1803B5358, v310);
                                                                                                                                  v312 = qword_1803B57D8;
                                                                                                                                  v313 = qword_1803CA780;
                                                                                                                                  v604 = qword_1803AB258;
                                                                                                                                  v565 = qword_1803B57E0;
                                                                                                                                  *(_DWORD *)(v18 + 40) = 51;
                                                                                                                                  v314 = sub_1803717E0(a1, v312, v313, Py_NoneStruct, v565, v604);
                                                                                                                                  v315 = (int *)v314;
                                                                                                                                  if ( v314 != 0 )
                                                                                                                                  {
                                                                                                                                    if ( (unsigned int)sub_18000CC50(v314, PyModule_Type) != 0 )
                                                                                                                                      v316 = sub_180371D80(a1, (_DWORD)v315, qword_1803CA780, qword_1803B52C8, qword_1803AB250);
                                                                                                                                    else
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
                                                                                                                                      sub_1800DE7D0(qword_1803CA780, qword_1803B52C8, v317);
                                                                                                                                      v319 = (int *)sub_180370EC0(a1);
                                                                                                                                      v320 = (_DWORD *)sub_180370EC0(a1);
                                                                                                                                      v321 = v320;
                                                                                                                                      if ( *v320 != -1 )
                                                                                                                                        ++*v320;
                                                                                                                                      v322 = sub_1800DFB50(a1);
                                                                                                                                      sub_18009B380(v321, qword_1803B57E8, v322);
                                                                                                                                      sub_18009B380(v321, qword_1803B57F8, qword_1803B57F0);
                                                                                                                                      v323 = (int *)sub_180372A40(a1, qword_1803B5800);
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
                                                                                                                                      v325 = (int *)sub_180372BB0(a1, qword_1803B5808);
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
                                                                                                                                      v327 = (int *)sub_1803728D0(a1, qword_1803B5810, qword_1803B5818);
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
                                                                                                                                      v329 = (int *)sub_180372BB0(a1, qword_1803B5820);
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
                                                                                                                                      v331 = (_DWORD *)qword_1803B5828;
                                                                                                                                      v332 = sub_18036FB80(a1, qword_1803B5830);
                                                                                                                                      if ( *v331 != -1 )
                                                                                                                                        ++*v331;
                                                                                                                                      v333 = (int *)sub_18012F5E0(a1, v331, v332);
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
                                                                                                                                      v335 = sub_18036FB80(a1, qword_1803B5848);
                                                                                                                                      v336 = (int *)sub_18012FA60(a1, v335);
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
                                                                                                                                      v338 = sub_18036FB80(a1, qword_1803B5858);
                                                                                                                                      v339 = (int *)sub_18012FE80(a1, v338);
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
                                                                                                                                      v341 = sub_18036FB80(a1, qword_1803B5868);
                                                                                                                                      v342 = (int *)sub_1801302A0(a1, v341);
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
                                                                                                                                      v344 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v345 = sub_18036FB80(a1, qword_1803B5878);
                                                                                                                                      if ( *v344 != -1 )
                                                                                                                                        ++*v344;
                                                                                                                                      v346 = (int *)sub_1801303C0(a1, v344, v345);
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
                                                                                                                                      v348 = (_DWORD *)qword_1803B5888;
                                                                                                                                      v349 = sub_18036FB80(a1, qword_1803B5890);
                                                                                                                                      if ( *v348 != -1 )
                                                                                                                                        ++*v348;
                                                                                                                                      v350 = (int *)sub_180130420(a1, v348, v349);
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
                                                                                                                                      v352 = sub_18036FB80(a1, qword_1803B58A0);
                                                                                                                                      v353 = (int *)sub_180130480(a1, v352);
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
                                                                                                                                      v355 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v356 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v355 != -1 )
                                                                                                                                        ++*v355;
                                                                                                                                      v357 = (int *)sub_1801304E0(a1, v355, v356);
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
                                                                                                                                      v359 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v360 = sub_18036FB80(a1, qword_1803B58C0);
                                                                                                                                      if ( *v359 != -1 )
                                                                                                                                        ++*v359;
                                                                                                                                      v361 = (int *)sub_1801305A0(a1, v359, v360);
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
                                                                                                                                      v363 = (_DWORD *)qword_1803B58D0;
                                                                                                                                      v364 = sub_18036FB80(a1, qword_1803B58D8);
                                                                                                                                      v365 = sub_18036FB80(a1, qword_1803B58E0);
                                                                                                                                      if ( *v363 != -1 )
                                                                                                                                        ++*v363;
                                                                                                                                      v366 = (int *)sub_18012F220(a1, v363, v364, v365);
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
                                                                                                                                      v368 = sub_18036FB80(a1, qword_1803B58F8);
                                                                                                                                      v369 = (int *)sub_18012F280(a1, v368);
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
                                                                                                                                      v371 = sub_18036FB80(a1, qword_1803B5908);
                                                                                                                                      v372 = (int *)sub_18012F2E0(a1, v371);
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
                                                                                                                                      v374 = sub_18036FB80(a1, qword_1803B5918);
                                                                                                                                      v375 = (int *)sub_18012F340(a1, v374);
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
                                                                                                                                      v377 = sub_18036FB80(a1, qword_1803B5928);
                                                                                                                                      v378 = (int *)sub_18012F3A0(a1, v377);
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
                                                                                                                                      v380 = sub_18036FB80(a1, qword_1803B5940);
                                                                                                                                      v381 = (int *)sub_18012F400(a1, v380);
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
                                                                                                                                      v383 = sub_18036FB80(a1, qword_1803B5958);
                                                                                                                                      v384 = sub_18036FB80(a1, qword_1803B5960);
                                                                                                                                      v385 = (int *)sub_18012F460(a1, v383, v384);
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
                                                                                                                                      v387 = (_DWORD *)qword_1803B5978;
                                                                                                                                      v388 = sub_18036FB80(a1, qword_1803B5980);
                                                                                                                                      if ( *v387 != -1 )
                                                                                                                                        ++*v387;
                                                                                                                                      v389 = (int *)sub_18012F4C0(a1, v387, v388);
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
                                                                                                                                      v391 = (_DWORD *)qword_1803B5998;
                                                                                                                                      v392 = sub_18036FB80(a1, qword_1803B59A0);
                                                                                                                                      if ( *v391 != -1 )
                                                                                                                                        ++*v391;
                                                                                                                                      v393 = (int *)sub_18012F520(a1, v391, v392);
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
                                                                                                                                      v395 = sub_18036FB80(a1, qword_1803B59B8);
                                                                                                                                      v396 = (int *)sub_18012F580(a1, v395);
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
                                                                                                                                      v398 = sub_18036FB80(a1, qword_1803B59C8);
                                                                                                                                      v399 = (int *)sub_18012F640(a1, v398);
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
                                                                                                                                      v401 = sub_18036FB80(a1, qword_1803B59D8);
                                                                                                                                      v402 = (int *)sub_18012F6A0(a1, v401);
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
                                                                                                                                      v404 = sub_18036FB80(a1, qword_1803B59E8);
                                                                                                                                      v405 = (int *)sub_18012F700(a1, v404);
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
                                                                                                                                      v407 = sub_18036FB80(a1, qword_1803B59F8);
                                                                                                                                      v408 = sub_18036FB80(a1, qword_1803B5A00);
                                                                                                                                      v409 = (int *)sub_18012F760(a1, v407, v408);
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
                                                                                                                                      v411 = sub_18036FB80(a1, qword_1803B5A10);
                                                                                                                                      v412 = (int *)sub_18012F820(a1, v411);
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
                                                                                                                                      v414 = sub_18036FB80(a1, qword_1803B5A20);
                                                                                                                                      v415 = (int *)sub_18012F880(a1, v414);
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
                                                                                                                                      v417 = (_DWORD *)qword_1803B5A30;
                                                                                                                                      v418 = sub_18036FB80(a1, qword_1803B5A38);
                                                                                                                                      if ( *v417 != -1 )
                                                                                                                                        ++*v417;
                                                                                                                                      v419 = (int *)sub_18012F8E0(a1, v417, v418);
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
                                                                                                                                      v421 = sub_18036FB80(a1, qword_1803B5A50);
                                                                                                                                      v422 = (int *)sub_18012F940(a1, v421);
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
                                                                                                                                      v424 = sub_18036FB80(a1, qword_1803B5A50);
                                                                                                                                      v425 = (int *)sub_18012F9A0(a1, v424);
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
                                                                                                                                      v427 = sub_18036FB80(a1, qword_1803B5A68);
                                                                                                                                      v428 = (int *)sub_18012FA00(a1, v427);
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
                                                                                                                                      v430 = sub_18036FB80(a1, qword_1803B5A78);
                                                                                                                                      v431 = (int *)sub_18012FAC0(a1, v430);
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
                                                                                                                                      v433 = (_DWORD *)qword_1803B4EE8;
                                                                                                                                      v434 = sub_18036FB80(a1, qword_1803B5A88);
                                                                                                                                      if ( *v433 != -1 )
                                                                                                                                        ++*v433;
                                                                                                                                      v435 = (int *)sub_18012FB20(a1, v433, v434);
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
                                                                                                                                      v437 = sub_18036FB80(a1, qword_1803B5A98);
                                                                                                                                      v438 = (int *)sub_18012FB80(a1, v437);
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
                                                                                                                                      v445 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v446 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v445 != -1 )
                                                                                                                                        ++*v445;
                                                                                                                                      v447 = (int *)sub_18012FC40(a1, v445, v446);
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
                                                                                                                                      v449 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v450 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v449 != -1 )
                                                                                                                                        ++*v449;
                                                                                                                                      v451 = (int *)sub_18012FCA0(a1, v449, v450);
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
                                                                                                                                      v453 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v454 = sub_18036FB80(a1, qword_1803B58B0);
                                                                                                                                      if ( *v453 != -1 )
                                                                                                                                        ++*v453;
                                                                                                                                      v455 = (int *)sub_18012FD00(a1, v453, v454);
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
                                                                                                                                      v457 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v458 = sub_18036FB80(a1, qword_1803B5AE8);
                                                                                                                                      if ( *v457 != -1 )
                                                                                                                                        ++*v457;
                                                                                                                                      v459 = (int *)sub_18012FD60(a1, v457, v458);
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
                                                                                                                                      v461 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v462 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v461 != -1 )
                                                                                                                                        ++*v461;
                                                                                                                                      v463 = (int *)sub_18012FDC0(a1, v461, v462);
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
                                                                                                                                      v465 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v466 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v465 != -1 )
                                                                                                                                        ++*v465;
                                                                                                                                      v467 = (int *)sub_18012FE20(a1, v465, v466);
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
                                                                                                                                      v469 = sub_18036FB80(a1, qword_1803B5B20);
                                                                                                                                      v470 = (int *)sub_18012FEE0(a1, v469);
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
                                                                                                                                      v472 = sub_18036FB80(a1, qword_1803B5B30);
                                                                                                                                      v473 = (int *)sub_18012FF40(a1, v472);
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
                                                                                                                                      v475 = sub_18036FB80(a1, qword_1803B5B40);
                                                                                                                                      v476 = (int *)sub_18012FFA0(a1, v475);
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
                                                                                                                                      v478 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v479 = sub_18036FB80(a1, qword_1803B5AF8);
                                                                                                                                      if ( *v478 != -1 )
                                                                                                                                        ++*v478;
                                                                                                                                      v480 = (int *)sub_180130000(a1, v478, v479);
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
                                                                                                                                      v482 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v483 = sub_18036FB80(a1, qword_1803B5B60);
                                                                                                                                      if ( *v482 != -1 )
                                                                                                                                        ++*v482;
                                                                                                                                      v484 = (int *)sub_180130060(a1, v482, v483);
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
                                                                                                                                      v486 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v487 = sub_18036FB80(a1, qword_1803B5B78);
                                                                                                                                      if ( *v486 != -1 )
                                                                                                                                        ++*v486;
                                                                                                                                      v488 = (int *)sub_1801300C0(a1, v486, v487);
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
                                                                                                                                      v490 = (_DWORD *)qword_1803B44E8;
                                                                                                                                      v491 = sub_18036FB80(a1, qword_1803B5B90);
                                                                                                                                      if ( *v490 != -1 )
                                                                                                                                        ++*v490;
                                                                                                                                      v492 = (int *)sub_180130120(a1, v490, v491);
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
                                                                                                                                      v494 = sub_18036FB80(a1, qword_1803B5BA8);
                                                                                                                                      v495 = sub_18036FB80(a1, qword_1803B5BB0);
                                                                                                                                      v496 = (int *)sub_180130180(a1, v494, v495);
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
                                                                                                                                      v498 = sub_18036FB80(a1, qword_1803B5BC8);
                                                                                                                                      v499 = (int *)sub_1801301E0(a1, v498);
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
                                                                                                                                      v501 = sub_18036FB80(a1, qword_1803B5BC8);
                                                                                                                                      v502 = (int *)sub_180130240(a1, v501);
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
                                                                                                                                      v504 = sub_18036FB80(a1, qword_1803B5BE8);
                                                                                                                                      v505 = (int *)sub_180130300(a1, v504);
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
                                                                                                                                      v507 = sub_18036FB80(a1, qword_1803B5BF8);
                                                                                                                                      v508 = (int *)sub_180130360(a1, v507);
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
