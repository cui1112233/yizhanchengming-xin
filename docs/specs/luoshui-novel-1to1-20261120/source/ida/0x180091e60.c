__int64 __fastcall sub_180091E60(_QWORD *a1, __int64 a2, __int64 a3)
{
  __int64 *v5; // rax
  __int64 v6; // rbx
  __int64 Length; // rax
  __int64 v8; // rax
  __int64 v9; // rax
  __int64 v10; // rcx
  _QWORD *v11; // rax
  __int64 Dict; // rax
  __int64 v13; // rcx
  __int64 v14; // rax
  __int64 v15; // rsi
  int *v16; // rbx
  __int64 v17; // rbp
  int **v18; // rax
  int *v19; // r8
  bool v20; // zf
  __int64 v21; // rax
  int *AttrString; // rbx
  __int64 v23; // rcx
  __int64 v24; // rsi
  __int64 v25; // rcx
  __int64 v26; // rcx
  __int64 v27; // rcx
  __int64 v28; // rdx
  _DWORD *v29; // r8
  __int64 v30; // r13
  __int64 v31; // rbx
  __int64 *v32; // rax
  __int64 v33; // rcx
  __int64 v34; // rdx
  __int64 *v35; // rax
  __int64 v36; // rbx
  unsigned int v37; // esi
  __int64 v38; // rax
  __int64 v39; // rcx
  __int64 v40; // rcx
  __int64 v41; // rax
  __int64 v42; // rcx
  __int64 v43; // rax
  __int64 v44; // rcx
  int v45; // edx
  int v46; // r8d
  __int64 v47; // rax
  __int64 v48; // rcx
  int v49; // edx
  int v50; // r8d
  __int64 v51; // rax
  __int64 v52; // rcx
  int v53; // edx
  int v54; // r8d
  __int64 v55; // rax
  int *v56; // rbx
  __int64 v57; // rax
  __int64 v58; // rcx
  __int64 v59; // rsi
  __int64 v60; // rax
  int *v61; // r14
  __int64 v62; // rax
  __int64 v63; // rcx
  __int64 v64; // rax
  __int64 v65; // rcx
  __int64 v66; // rax
  __int64 v67; // rcx
  __int64 v68; // rax
  __int64 v69; // rcx
  int v70; // edx
  int v71; // r8d
  __int64 v72; // rax
  __int64 v73; // rax
  __int64 v74; // rcx
  __int64 v75; // rax
  __int64 v76; // rcx
  int v77; // edx
  int v78; // r8d
  __int64 v79; // rax
  int *v80; // rbx
  __int64 v81; // rax
  __int64 v82; // rcx
  __int64 v83; // rsi
  int v84; // edx
  int v85; // r8d
  __int64 v86; // rax
  int *v87; // rbx
  __int64 v88; // rax
  __int64 v89; // rcx
  __int64 v90; // rsi
  int v91; // edx
  int v92; // r8d
  __int64 v93; // rax
  __int64 v94; // rax
  __int64 v95; // rcx
  __int64 v96; // rax
  __int64 v97; // rcx
  __int64 v98; // rax
  __int64 v99; // rcx
  int v100; // edx
  int v101; // r8d
  __int64 v102; // rax
  __int64 v103; // rax
  __int64 v104; // rcx
  __int64 v105; // rax
  __int64 v106; // rcx
  __int64 v107; // rax
  __int64 v108; // rcx
  __int64 v109; // rax
  __int64 v110; // rcx
  __int64 v111; // rax
  __int64 v112; // rcx
  int v113; // edx
  int v114; // r8d
  __int64 v115; // rax
  int *v116; // rbx
  __int64 v117; // rax
  __int64 v118; // rcx
  __int64 v119; // rsi
  int v120; // edx
  int v121; // r8d
  __int64 v122; // rax
  int *v123; // rbx
  __int64 v124; // rax
  __int64 v125; // rcx
  __int64 v126; // rsi
  int v127; // edx
  int v128; // r8d
  __int64 v129; // rax
  int *v130; // rbx
  __int64 v131; // rax
  __int64 v132; // rcx
  __int64 v133; // rsi
  int v134; // edx
  int v135; // r8d
  __int64 v136; // rax
  __int64 v137; // rax
  __int64 v138; // rcx
  __int64 v139; // rax
  __int64 v140; // rcx
  int v141; // edx
  int v142; // r8d
  __int64 v143; // rax
  int *v144; // rbx
  __int64 v145; // rax
  __int64 v146; // rcx
  __int64 v147; // rsi
  int v148; // edx
  int v149; // r8d
  __int64 v150; // rax
  int *v151; // rbx
  __int64 v152; // rax
  __int64 v153; // rcx
  __int64 v154; // rsi
  int *v155; // r12
  __int64 v156; // r14
  __int64 *v157; // rax
  __int64 v158; // r8
  __int64 *v159; // rax
  int *v160; // rbx
  __int64 v161; // rax
  int *v162; // rbx
  __int64 v163; // rax
  int *v164; // rbx
  __int64 v165; // rax
  int *v166; // rbx
  int *v167; // rbx
  _DWORD *v168; // rbx
  __int64 v169; // rax
  int *v170; // rbx
  __int64 v171; // rax
  int *v172; // rbx
  _DWORD *v173; // rbx
  __int64 v174; // rax
  int *v175; // rbx
  _DWORD *v176; // rbx
  __int64 v177; // rax
  int *v178; // rbx
  __int64 v179; // rax
  int *v180; // rbx
  __int64 v181; // rax
  int *v182; // rbx
  __int64 v183; // rax
  int *v184; // rbx
  __int64 v185; // rax
  int *v186; // rbx
  __int64 v187; // rax
  int *v188; // rbx
  __int64 v189; // rax
  int *v190; // rbx
  __int64 v191; // rax
  int *v192; // rbx
  _DWORD *v193; // rbx
  __int64 v194; // rax
  int *v195; // rbx
  __int64 v196; // rax
  int *v197; // rbx
  __int64 v198; // rax
  int *v199; // rbx
  __int64 v200; // rax
  int *v201; // rbx
  __int64 v202; // rax
  int *v203; // rbx
  __int64 v204; // rax
  int *v205; // rbx
  _DWORD *v206; // rbx
  __int64 v207; // rax
  int *v208; // rbx
  __int64 v209; // rax
  int *v210; // rbx
  __int64 v211; // rax
  int *v212; // rbx
  __int64 v213; // rax
  int *v214; // rbx
  __int64 v215; // rax
  int *v216; // rbx
  __int64 v217; // rax
  int *v218; // rbx
  __int64 v219; // rax
  int *v220; // rbx
  __int64 v221; // rax
  int *v222; // rbx
  __int64 v223; // rax
  int *v224; // rbx
  __int64 v225; // rax
  int *v226; // rbx
  __int64 v227; // rax
  int *v228; // rbx
  __int64 v229; // rax
  int *v230; // rbx
  __int64 v231; // rax
  int *v232; // rbx
  __int64 v233; // rax
  int *v234; // rbx
  __int64 v235; // rax
  int *v236; // rbx
  __int64 v237; // rax
  int *v238; // rbx
  __int64 v239; // rax
  int *v240; // rbx
  __int64 v241; // rax
  int *v242; // rbx
  __int64 v243; // rax
  int *v244; // rbx
  __int64 v245; // rax
  int *v246; // rbx
  __int64 v247; // rax
  int *v248; // rbx
  __int64 v249; // rax
  int *v250; // rbx
  _DWORD *v251; // rbx
  __int64 v252; // rax
  int *v253; // rbx
  __int64 v254; // rax
  int *v255; // rbx
  __int64 v256; // rax
  int *v257; // rbx
  __int64 v258; // rax
  int *v259; // rbx
  __int64 v260; // rax
  int *v261; // rbx
  __int64 v262; // rax
  int *v263; // rbx
  __int64 v264; // rax
  int *v265; // rbx
  __int64 v266; // rax
  int *v267; // rbx
  __int64 v268; // rax
  int *v269; // rbx
  __int64 v270; // rax
  int *v271; // rbx
  __int64 v272; // rax
  int *v273; // rbx
  __int64 v274; // rax
  int *v275; // rbx
  __int64 v276; // rax
  int *v277; // rbx
  __int64 v278; // rax
  int *v279; // rbx
  _DWORD *v280; // rbx
  __int64 v281; // rax
  int *v282; // rbx
  __int64 v283; // rcx
  _DWORD *v284; // r15
  int *v285; // rbx
  __int64 v286; // rcx
  __int64 v287; // rcx
  __int64 v288; // rsi
  __int64 v289; // r14
  __int64 v290; // rax
  int *v291; // rsi
  int *v292; // rcx
  __int64 v293; // rcx
  __int64 v294; // rsi
  int *v295; // rcx
  int *v296; // rcx
  __int64 result; // rax
  __int64 v298; // rdx
  __int64 v299; // [rsp+20h] [rbp-68h]
  __int64 v300; // [rsp+20h] [rbp-68h]
  __int64 v301; // [rsp+20h] [rbp-68h]
  __int64 v302; // [rsp+20h] [rbp-68h]
  __int64 v303; // [rsp+20h] [rbp-68h]
  __int64 v304; // [rsp+20h] [rbp-68h]
  __int64 v305; // [rsp+20h] [rbp-68h]
  __int64 v306; // [rsp+20h] [rbp-68h]
  __int64 v307; // [rsp+20h] [rbp-68h]
  __int64 v308; // [rsp+20h] [rbp-68h]
  __int64 v309; // [rsp+20h] [rbp-68h]
  __int64 v310; // [rsp+20h] [rbp-68h]
  __int64 v311; // [rsp+28h] [rbp-60h]
  __int64 v312; // [rsp+28h] [rbp-60h]
  __int64 v313; // [rsp+28h] [rbp-60h]
  __int64 v314; // [rsp+28h] [rbp-60h]
  __int64 v315; // [rsp+28h] [rbp-60h]
  __int64 v316; // [rsp+28h] [rbp-60h]
  __int64 v317; // [rsp+28h] [rbp-60h]
  __int64 v318; // [rsp+28h] [rbp-60h]
  __int64 v319; // [rsp+28h] [rbp-60h]
  __int64 v320; // [rsp+28h] [rbp-60h]
  __int64 v321; // [rsp+28h] [rbp-60h]
  __int64 v322; // [rsp+28h] [rbp-60h]
  __int64 v323; // [rsp+28h] [rbp-60h]
  __int64 v324; // [rsp+28h] [rbp-60h]

  qword_1803CA830 = *(_QWORD *)(a2 + 16);
  qword_1803CA828 = a2;
  if ( byte_1803AD098 == 0 )
  {
    if ( byte_1803AD099 == 0 )
    {
      sub_180388A30(a1, &qword_1803AFF50, "novels.api.novels_api");
      byte_1803AD099 = 1;
    }
    sub_1800668E0();
    byte_1803AD098 = 1;
  }
  sub_180065C90(a1, qword_1803AB330, qword_1803AB560);
  v5 = (__int64 *)sub_1800015A0(qword_1803CA830, qword_1803AB2A8);
  v6 = 0;
  if ( v5 != nullptr )
    v6 = *v5;
  Length = PyUnicode_GetLength(v6);
  v8 = PyUnicode_Find(v6, qword_1803AB4A8, 0, Length, -1);
  if ( v8 != -1 )
  {
    v9 = PyUnicode_Substring(v6, 0, v8);
    sub_180065D20(v10, qword_1803AB2B0, v9);
  }
  v11 = (_QWORD *)sub_1800015A0(qword_1803CA830, qword_1803AB300);
  if ( v11 == nullptr || *v11 == 0 )
  {
    Dict = PyModule_GetDict(qword_1803AB780);
    sub_180065C90(v13, qword_1803AB300, Dict);
  }
  v14 = sub_18038D790(a3);
  v15 = qword_1803AB328;
  v16 = (int *)v14;
  v17 = qword_1803CA830;
  v18 = (int **)sub_1800015A0(qword_1803CA830, qword_1803AB328);
  if ( v18 == nullptr )
    goto LABEL_20;
  v19 = *v18;
  if ( v16 != *v18 )
  {
    if ( v19 != nullptr )
    {
      if ( *v16 != -1 )
        ++*v16;
      *v18 = v16;
      if ( *v19 >= 0 )
      {
        v20 = (*(_QWORD *)v19)-- == 1;
        if ( v20 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v19 + 1) + 48LL))(v19);
      }
    }
    else
    {
LABEL_20:
      PyDict_SetItem(v17, v15, v16);
    }
  }
  v21 = qword_1803C9FD0;
  if ( qword_1803C9FD0 == 0 )
  {
    v21 = PyImport_ImportModule("importlib._bootstrap");
    qword_1803C9FD0 = v21;
  }
  AttrString = (int *)PyObject_GetAttrString(v21, "_spec_from_module");
  v24 = sub_180376770(a1, AttrString, qword_1803CA828);
  if ( *AttrString >= 0 )
  {
    v20 = (*(_QWORD *)AttrString)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)AttrString + 1) + 48LL))(AttrString);
  }
  if ( v24 == 0 )
  {
    PyErr_PrintEx(0);
    abort();
  }
  sub_18037F3A0(v23, v24, qword_1803AB4D0, Py_TrueStruct);
  sub_180065D20(v25, qword_1803AB4C8, v24);
  sub_180065C90(v26, qword_1803B0830, Py_NoneStruct);
  sub_180065C90(v27, qword_1803B0838, qword_1803AFF30);
  v28 = qword_1803CA828;
  v29 = *(_DWORD **)(qword_1803CA828 + 16);
  if ( *v29 != -1 )
    ++*v29;
  v30 = sub_18036CBB0(qword_1803AFE30, v28, v29, 0);
  sub_180001D00(a1, v30);
  v31 = qword_1803AFF30;
  v32 = (__int64 *)sub_1800015A0(qword_1803CA830, qword_1803B0F98);
  if ( v32 == nullptr || (v34 = *v32, *v32 == 0) )
  {
    v35 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B0F98);
    v34 = 0;
    if ( v35 != nullptr )
      v34 = *v35;
  }
  if ( (unsigned __int8)sub_18037F3A0(v33, v34, qword_1803B0840, v31) == 0
    || (v38 = sub_180066560(), (unsigned __int8)sub_18037F3A0(v39, v38, qword_1803B0848, Py_TrueStruct) == 0) )
  {
    v36 = a1[12];
    v37 = 1;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065C90(v40, qword_1803B0850, Py_NoneStruct);
  v41 = sub_180372070();
  v43 = sub_18037F0C0(v42, v41, qword_1803B0858);
  sub_180065D20(v44, qword_1803B0858, v43);
  v45 = qword_1803AFF58;
  v46 = qword_1803CA830;
  v311 = qword_1803AB250;
  *(_DWORD *)(v30 + 40) = 3;
  v47 = sub_1803717E0((_DWORD)a1, v45, v46, Py_NoneStruct, Py_NoneStruct, v311);
  if ( v47 == 0 )
  {
    v36 = a1[12];
    v37 = 3;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065D20(v48, qword_1803AFF58, v47);
  v49 = qword_1803B0490;
  v50 = qword_1803CA830;
  v312 = qword_1803AB250;
  *(_DWORD *)(v30 + 40) = 4;
  v51 = sub_1803717E0((_DWORD)a1, v49, v50, Py_NoneStruct, Py_NoneStruct, v312);
  if ( v51 == 0 )
  {
    v36 = a1[12];
    v37 = 4;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065D20(v52, qword_1803B0490, v51);
  v53 = qword_1803B0860;
  v54 = qword_1803CA830;
  v313 = qword_1803AB250;
  v299 = qword_1803B0868;
  *(_DWORD *)(v30 + 40) = 5;
  v55 = sub_1803717E0((_DWORD)a1, v53, v54, Py_NoneStruct, v299, v313);
  v56 = (int *)v55;
  if ( v55 == 0 )
  {
    v36 = a1[12];
    v37 = 5;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v55 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v57 = sub_180371D80((_DWORD)a1, (_DWORD)v56, qword_1803CA830, qword_1803AFFA8, qword_1803AB250);
  else
    v57 = sub_180371890(a1, v56, qword_1803AFFA8);
  v59 = v57;
  if ( *v56 >= 0 )
  {
    v20 = (*(_QWORD *)v56)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v56 + 1) + 48LL))(v56);
  }
  if ( v59 == 0 )
  {
    v36 = a1[12];
    v37 = 5;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065D20(v58, qword_1803AFFA8, v59);
  v60 = sub_180372110();
  v61 = (int *)v60;
  if ( *(_DWORD *)v60 != -1 )
    ++*(_DWORD *)v60;
  if ( *(_QWORD *)(v60 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v62 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0870, qword_1803AB250);
  else
    v62 = sub_180371890(a1, v61, qword_1803B0870);
  if ( v62 == 0
    || ((sub_180065D20(v63, qword_1803B0870, v62), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
      ? (v64 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0878, qword_1803AB250))
      : (v64 = sub_180371890(a1, v61, qword_1803B0878)),
        v64 == 0
     || ((sub_180065D20(v65, qword_1803B0878, v64), *((_QWORD *)v61 + 1) == PyModule_Type)
      || (unsigned int)PyType_IsSubtype() != 0
       ? (v66 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0880, qword_1803AB250))
       : (v66 = sub_180371890(a1, v61, qword_1803B0880)),
         v66 == 0
      || ((sub_180065D20(v67, qword_1803B0880, v66), *((_QWORD *)v61 + 1) == PyModule_Type)
       || (unsigned int)PyType_IsSubtype() != 0
        ? (v68 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0888, qword_1803AB250))
        : (v68 = sub_180371890(a1, v61, qword_1803B0888)),
          v68 == 0))) )
  {
    v37 = 6;
LABEL_75:
    v36 = a1[12];
    a1[12] = 0;
    if ( *v61 >= 0 )
    {
      v20 = (*(_QWORD *)v61)-- == 1;
      if ( v20 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
    }
    goto LABEL_462;
  }
  sub_180065D20(v69, qword_1803B0888, v68);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
  v70 = qword_1803B0890;
  v71 = qword_1803CA830;
  v314 = qword_1803B0288;
  v300 = qword_1803B0898;
  *(_DWORD *)(v30 + 40) = 8;
  v72 = sub_1803717E0((_DWORD)a1, v70, v71, Py_NoneStruct, v300, v314);
  v61 = (int *)v72;
  if ( v72 == 0 )
  {
    v36 = a1[12];
    v37 = 8;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v72 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v73 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B02E0, qword_1803AB250);
  else
    v73 = sub_180371890(a1, v61, qword_1803B02E0);
  if ( v73 == 0
    || ((sub_180065D20(v74, qword_1803B02E0, v73), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
      ? (v75 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0500, qword_1803AB250))
      : (v75 = sub_180371890(a1, v61, qword_1803B0500)),
        v75 == 0) )
  {
    v37 = 8;
    goto LABEL_75;
  }
  sub_180065D20(v76, qword_1803B0500, v75);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
  v77 = qword_1803B08A0;
  v78 = qword_1803CA830;
  v315 = qword_1803B0288;
  v301 = qword_1803B08A8;
  *(_DWORD *)(v30 + 40) = 9;
  v79 = sub_1803717E0((_DWORD)a1, v77, v78, Py_NoneStruct, v301, v315);
  v80 = (int *)v79;
  if ( v79 == 0 )
  {
    v36 = a1[12];
    v37 = 9;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v79 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v81 = sub_180371D80((_DWORD)a1, (_DWORD)v80, qword_1803CA830, qword_1803B00A0, qword_1803AB250);
  else
    v81 = sub_180371890(a1, v80, qword_1803B00A0);
  v83 = v81;
  if ( *v80 >= 0 )
  {
    v20 = (*(_QWORD *)v80)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v80 + 1) + 48LL))(v80);
  }
  if ( v83 == 0 )
  {
    v36 = a1[12];
    v37 = 9;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065D20(v82, qword_1803B00A0, v83);
  v84 = qword_1803B08B0;
  v85 = qword_1803CA830;
  v316 = qword_1803B0288;
  v302 = qword_1803B08B8;
  *(_DWORD *)(v30 + 40) = 10;
  v86 = sub_1803717E0((_DWORD)a1, v84, v85, Py_NoneStruct, v302, v316);
  v87 = (int *)v86;
  if ( v86 == 0 )
  {
    v36 = a1[12];
    v37 = 10;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v86 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v88 = sub_180371D80((_DWORD)a1, (_DWORD)v87, qword_1803CA830, qword_1803B00C8, qword_1803AB250);
  else
    v88 = sub_180371890(a1, v87, qword_1803B00C8);
  v90 = v88;
  if ( *v87 >= 0 )
  {
    v20 = (*(_QWORD *)v87)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v87 + 1) + 48LL))(v87);
  }
  if ( v90 == 0 )
  {
    v36 = a1[12];
    v37 = 10;
    a1[12] = 0;
    goto LABEL_462;
  }
  sub_180065D20(v89, qword_1803B00C8, v90);
  v91 = qword_1803B08C0;
  v92 = qword_1803CA830;
  v317 = qword_1803B0288;
  v303 = qword_1803B08C8;
  *(_DWORD *)(v30 + 40) = 11;
  v93 = sub_1803717E0((_DWORD)a1, v91, v92, Py_NoneStruct, v303, v317);
  v61 = (int *)v93;
  if ( v93 == 0 )
  {
    v36 = a1[12];
    v37 = 11;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v93 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v94 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B05C8, qword_1803AB250);
  else
    v94 = sub_180371890(a1, v61, qword_1803B05C8);
  if ( v94 == 0
    || ((sub_180065D20(v95, qword_1803B05C8, v94), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
      ? (v96 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0050, qword_1803AB250))
      : (v96 = sub_180371890(a1, v61, qword_1803B0050)),
        v96 == 0
     || ((sub_180065D20(v97, qword_1803B0050, v96), *((_QWORD *)v61 + 1) == PyModule_Type)
      || (unsigned int)PyType_IsSubtype() != 0
       ? (v98 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B05D8, qword_1803AB250))
       : (v98 = sub_180371890(a1, v61, qword_1803B05D8)),
         v98 == 0)) )
  {
    v37 = 11;
    goto LABEL_75;
  }
  sub_180065D20(v99, qword_1803B05D8, v98);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
  v100 = qword_1803B08D0;
  v101 = qword_1803CA830;
  v318 = qword_1803B0288;
  v304 = qword_1803B08D8;
  *(_DWORD *)(v30 + 40) = 16;
  v102 = sub_1803717E0((_DWORD)a1, v100, v101, Py_NoneStruct, v304, v318);
  v61 = (int *)v102;
  if ( v102 == 0 )
  {
    v36 = a1[12];
    v37 = 16;
    a1[12] = 0;
    goto LABEL_462;
  }
  if ( *(_QWORD *)(v102 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v103 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0408, qword_1803AB250);
  else
    v103 = sub_180371890(a1, v61, qword_1803B0408);
  if ( v103 != 0 )
  {
    sub_180065D20(v104, qword_1803B0408, v103);
    if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
      v105 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B03C0, qword_1803AB250);
    else
      v105 = sub_180371890(a1, v61, qword_1803B03C0);
    if ( v105 != 0 )
    {
      sub_180065D20(v106, qword_1803B03C0, v105);
      if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
        v107 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0340, qword_1803AB250);
      else
        v107 = sub_180371890(a1, v61, qword_1803B0340);
      if ( v107 != 0 )
      {
        sub_180065D20(v108, qword_1803B0340, v107);
        if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
          v109 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B03F8, qword_1803AB250);
        else
          v109 = sub_180371890(a1, v61, qword_1803B03F8);
        if ( v109 != 0 )
        {
          sub_180065D20(v110, qword_1803B03F8, v109);
          if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
            v111 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0390, qword_1803AB250);
          else
            v111 = sub_180371890(a1, v61, qword_1803B0390);
          if ( v111 != 0 )
          {
            sub_180065D20(v112, qword_1803B0390, v111);
            if ( *v61 >= 0 )
            {
              v20 = (*(_QWORD *)v61)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
            }
            v113 = qword_1803B08E0;
            v114 = qword_1803CA830;
            v319 = qword_1803B0288;
            v305 = qword_1803B08E8;
            *(_DWORD *)(v30 + 40) = 23;
            v115 = sub_1803717E0((_DWORD)a1, v113, v114, Py_NoneStruct, v305, v319);
            v116 = (int *)v115;
            if ( v115 == 0 )
            {
              v36 = a1[12];
              v37 = 23;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v115 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v117 = sub_180371D80((_DWORD)a1, (_DWORD)v116, qword_1803CA830, qword_1803B0368, qword_1803AB250);
            else
              v117 = sub_180371890(a1, v116, qword_1803B0368);
            v119 = v117;
            if ( *v116 >= 0 )
            {
              v20 = (*(_QWORD *)v116)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v116 + 1) + 48LL))(v116);
            }
            if ( v119 == 0 )
            {
              v36 = a1[12];
              v37 = 23;
              a1[12] = 0;
              goto LABEL_462;
            }
            sub_180065D20(v118, qword_1803B0368, v119);
            v120 = qword_1803B08F0;
            v121 = qword_1803CA830;
            v320 = qword_1803B0288;
            v306 = qword_1803B08F8;
            *(_DWORD *)(v30 + 40) = 24;
            v122 = sub_1803717E0((_DWORD)a1, v120, v121, Py_NoneStruct, v306, v320);
            v123 = (int *)v122;
            if ( v122 == 0 )
            {
              v36 = a1[12];
              v37 = 24;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v122 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v124 = sub_180371D80((_DWORD)a1, (_DWORD)v123, qword_1803CA830, qword_1803AFFB8, qword_1803AB250);
            else
              v124 = sub_180371890(a1, v123, qword_1803AFFB8);
            v126 = v124;
            if ( *v123 >= 0 )
            {
              v20 = (*(_QWORD *)v123)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v123 + 1) + 48LL))(v123);
            }
            if ( v126 == 0 )
            {
              v36 = a1[12];
              v37 = 24;
              a1[12] = 0;
              goto LABEL_462;
            }
            sub_180065D20(v125, qword_1803AFFB8, v126);
            v127 = qword_1803B0900;
            v128 = qword_1803CA830;
            v321 = qword_1803B0288;
            v307 = qword_1803B0908;
            *(_DWORD *)(v30 + 40) = 25;
            v129 = sub_1803717E0((_DWORD)a1, v127, v128, Py_NoneStruct, v307, v321);
            v130 = (int *)v129;
            if ( v129 == 0 )
            {
              v36 = a1[12];
              v37 = 25;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v129 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v131 = sub_180371D80((_DWORD)a1, (_DWORD)v130, qword_1803CA830, qword_1803AFFE0, qword_1803AB250);
            else
              v131 = sub_180371890(a1, v130, qword_1803AFFE0);
            v133 = v131;
            if ( *v130 >= 0 )
            {
              v20 = (*(_QWORD *)v130)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v130 + 1) + 48LL))(v130);
            }
            if ( v133 == 0 )
            {
              v36 = a1[12];
              v37 = 25;
              a1[12] = 0;
              goto LABEL_462;
            }
            sub_180065D20(v132, qword_1803AFFE0, v133);
            v134 = qword_1803B0910;
            v135 = qword_1803CA830;
            v322 = qword_1803B0288;
            v308 = qword_1803B0918;
            *(_DWORD *)(v30 + 40) = 26;
            v136 = sub_1803717E0((_DWORD)a1, v134, v135, Py_NoneStruct, v308, v322);
            v61 = (int *)v136;
            if ( v136 == 0 )
            {
              v36 = a1[12];
              v37 = 26;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v136 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v137 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0740, qword_1803AB250);
            else
              v137 = sub_180371890(a1, v61, qword_1803B0740);
            if ( v137 == 0
              || ((sub_180065D20(v138, qword_1803B0740, v137), *((_QWORD *)v61 + 1) == PyModule_Type)
               || (unsigned int)PyType_IsSubtype() != 0
                ? (v139 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0748, qword_1803AB250))
                : (v139 = sub_180371890(a1, v61, qword_1803B0748)),
                  v139 == 0) )
            {
              v37 = 26;
              goto LABEL_75;
            }
            sub_180065D20(v140, qword_1803B0748, v139);
            if ( *v61 >= 0 )
            {
              v20 = (*(_QWORD *)v61)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
            }
            v141 = qword_1803B0920;
            v142 = qword_1803CA830;
            v323 = qword_1803B0288;
            v309 = qword_1803B0928;
            *(_DWORD *)(v30 + 40) = 27;
            v143 = sub_1803717E0((_DWORD)a1, v141, v142, Py_NoneStruct, v309, v323);
            v144 = (int *)v143;
            if ( v143 == 0 )
            {
              v36 = a1[12];
              v37 = 27;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v143 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v145 = sub_180371D80((_DWORD)a1, (_DWORD)v144, qword_1803CA830, qword_1803B0268, qword_1803AB250);
            else
              v145 = sub_180371890(a1, v144, qword_1803B0268);
            v147 = v145;
            if ( *v144 >= 0 )
            {
              v20 = (*(_QWORD *)v144)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v144 + 1) + 48LL))(v144);
            }
            if ( v147 == 0 )
            {
              v36 = a1[12];
              v37 = 27;
              a1[12] = 0;
              goto LABEL_462;
            }
            sub_180065D20(v146, qword_1803B0268, v147);
            v148 = qword_1803B0930;
            v149 = qword_1803CA830;
            v324 = qword_1803B0288;
            v310 = qword_1803B0938;
            *(_DWORD *)(v30 + 40) = 28;
            v150 = sub_1803717E0((_DWORD)a1, v148, v149, Py_NoneStruct, v310, v324);
            v151 = (int *)v150;
            if ( v150 == 0 )
            {
              v36 = a1[12];
              v37 = 28;
              a1[12] = 0;
              goto LABEL_462;
            }
            if ( *(_QWORD *)(v150 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v152 = sub_180371D80((_DWORD)a1, (_DWORD)v151, qword_1803CA830, qword_1803B0338, qword_1803AB250);
            else
              v152 = sub_180371890(a1, v151, qword_1803B0338);
            v154 = v152;
            if ( *v151 >= 0 )
            {
              v20 = (*(_QWORD *)v151)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v151 + 1) + 48LL))(v151);
            }
            if ( v154 == 0 )
            {
              v36 = a1[12];
              v37 = 28;
              a1[12] = 0;
              goto LABEL_462;
            }
            sub_180065D20(v153, qword_1803B0338, v154);
            v155 = (int *)sub_180370EC0(a1);
            v156 = sub_180370EC0(a1);
            if ( *(_DWORD *)v156 != -1 )
              ++*(_DWORD *)v156;
            v157 = (__int64 *)sub_1800015A0(qword_1803CA830, qword_1803B0F90);
            if ( v157 == nullptr || (v158 = *v157, *v157 == 0) )
            {
              v159 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B0F90);
              v158 = 0;
              if ( v159 != nullptr )
                v158 = *v159;
            }
            PyDict_SetItem(v156, qword_1803B0940, v158);
            PyDict_SetItem(v156, qword_1803B0950, qword_1803B0948);
            v160 = (int *)sub_1803609A0(
                            (int)sub_180067D50,
                            qword_1803B0958,
                            qword_1803B0960,
                            qword_1803AFE90,
                            0,
                            0,
                            0,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0958, v160);
            if ( *v160 >= 0 )
            {
              v20 = (*(_QWORD *)v160)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v160 + 1) + 48LL))(v160);
            }
            v161 = sub_18036FB80(a1, qword_1803B0968);
            v162 = (int *)sub_1803609A0(
                            (int)sub_1800681B0,
                            qword_1803AFFE8,
                            qword_1803B0970,
                            qword_1803AFC60,
                            0,
                            0,
                            v161,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803AFFE8, v162);
            if ( *v162 >= 0 )
            {
              v20 = (*(_QWORD *)v162)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v162 + 1) + 48LL))(v162);
            }
            v163 = sub_18036FB80(a1, qword_1803B0978);
            v164 = (int *)sub_1803609A0(
                            (int)sub_1800688D0,
                            qword_1803AFFA0,
                            qword_1803B0980,
                            qword_1803AFD50,
                            0,
                            0,
                            v163,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803AFFA0, v164);
            if ( *v164 >= 0 )
            {
              v20 = (*(_QWORD *)v164)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v164 + 1) + 48LL))(v164);
            }
            v165 = sub_18036FB80(a1, qword_1803B0988);
            v166 = (int *)sub_1803609A0(
                            (int)sub_180068CE0,
                            qword_1803B0058,
                            qword_1803B0990,
                            qword_1803AFC80,
                            0,
                            0,
                            v165,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0058, v166);
            if ( *v166 >= 0 )
            {
              v20 = (*(_QWORD *)v166)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v166 + 1) + 48LL))(v166);
            }
            v167 = (int *)sub_1803609A0(
                            (int)sub_1800690A0,
                            qword_1803B00A8,
                            qword_1803B0998,
                            qword_1803AFCB0,
                            0,
                            0,
                            0,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00A8, v167);
            if ( *v167 >= 0 )
            {
              v20 = (*(_QWORD *)v167)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v167 + 1) + 48LL))(v167);
            }
            v168 = (_DWORD *)qword_1803B09A0;
            v169 = sub_18036FB80(a1, qword_1803B09A8);
            if ( *v168 != -1 )
              ++*v168;
            v170 = (int *)sub_1803609A0(
                            (int)sub_180069340,
                            qword_1803B00B8,
                            qword_1803B09B0,
                            qword_1803AFF38,
                            (__int64)v168,
                            0,
                            v169,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00B8, v170);
            if ( *v170 >= 0 )
            {
              v20 = (*(_QWORD *)v170)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v170 + 1) + 48LL))(v170);
            }
            v171 = sub_18036FB80(a1, qword_1803B09B8);
            v172 = (int *)sub_1803609A0(
                            (int)sub_180069B90,
                            qword_1803B00E0,
                            qword_1803B09C0,
                            qword_1803AFE38,
                            0,
                            0,
                            v171,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00E0, v172);
            if ( *v172 >= 0 )
            {
              v20 = (*(_QWORD *)v172)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v172 + 1) + 48LL))(v172);
            }
            v173 = (_DWORD *)qword_1803B09C8;
            v174 = sub_18036FB80(a1, qword_1803B09D0);
            if ( *v173 != -1 )
              ++*v173;
            v175 = (int *)sub_1803609A0(
                            (int)sub_18006A620,
                            qword_1803B00D0,
                            qword_1803B09D8,
                            qword_1803AFF00,
                            (__int64)v173,
                            0,
                            v174,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00D0, v175);
            if ( *v175 >= 0 )
            {
              v20 = (*(_QWORD *)v175)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v175 + 1) + 48LL))(v175);
            }
            v176 = (_DWORD *)qword_1803B09C8;
            v177 = sub_18036FB80(a1, qword_1803B09E0);
            if ( *v176 != -1 )
              ++*v176;
            v178 = (int *)sub_1803609A0(
                            (int)sub_18006AC60,
                            qword_1803B0248,
                            qword_1803B09E8,
                            qword_1803AFD48,
                            (__int64)v176,
                            0,
                            v177,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0248, v178);
            if ( *v178 >= 0 )
            {
              v20 = (*(_QWORD *)v178)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v178 + 1) + 48LL))(v178);
            }
            v179 = sub_18036FB80(a1, qword_1803B09F0);
            v180 = (int *)sub_1803609A0(
                            (int)sub_18006B5E0,
                            qword_1803B0628,
                            qword_1803B09F8,
                            qword_1803AFB70,
                            0,
                            0,
                            v179,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0628, v180);
            if ( *v180 >= 0 )
            {
              v20 = (*(_QWORD *)v180)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v180 + 1) + 48LL))(v180);
            }
            v181 = sub_18036FB80(a1, qword_1803B0A00);
            v182 = (int *)sub_1803609A0(
                            (int)sub_18006BD80,
                            qword_1803B0220,
                            qword_1803B0A08,
                            qword_1803AFB20,
                            0,
                            0,
                            v181,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0220, v182);
            if ( *v182 >= 0 )
            {
              v20 = (*(_QWORD *)v182)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v182 + 1) + 48LL))(v182);
            }
            v183 = sub_18036FB80(a1, qword_1803B0A10);
            v184 = (int *)sub_1803609A0(
                            (int)sub_18006C590,
                            qword_1803B00B0,
                            qword_1803B0A18,
                            qword_1803AFD58,
                            0,
                            0,
                            v183,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00B0, v184);
            if ( *v184 >= 0 )
            {
              v20 = (*(_QWORD *)v184)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v184 + 1) + 48LL))(v184);
            }
            v185 = sub_18036FB80(a1, qword_1803B0A20);
            v186 = (int *)sub_1803609A0(
                            (int)sub_18006CDE0,
                            qword_1803B0A28,
                            qword_1803B0A30,
                            qword_1803AFF18,
                            0,
                            0,
                            v185,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A28, v186);
            if ( *v186 >= 0 )
            {
              v20 = (*(_QWORD *)v186)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v186 + 1) + 48LL))(v186);
            }
            v187 = sub_18036FB80(a1, qword_1803B0A38);
            v188 = (int *)sub_1803609A0(
                            (int)sub_18006D730,
                            qword_1803B0A40,
                            qword_1803B0A48,
                            qword_1803AFB50,
                            0,
                            0,
                            v187,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A40, v188);
            if ( *v188 >= 0 )
            {
              v20 = (*(_QWORD *)v188)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v188 + 1) + 48LL))(v188);
            }
            v189 = sub_18036FB80(a1, qword_1803B0A50);
            v190 = (int *)sub_1803609A0(
                            (int)sub_18006DEE0,
                            qword_1803B0A58,
                            qword_1803B0A60,
                            qword_1803B0FC8,
                            0,
                            0,
                            v189,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A58, v190);
            if ( *v190 >= 0 )
            {
              v20 = (*(_QWORD *)v190)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v190 + 1) + 48LL))(v190);
            }
            v191 = sub_18036FB80(a1, qword_1803B0A38);
            v192 = (int *)sub_1803609A0(
                            (int)sub_18006E800,
                            qword_1803B0A68,
                            qword_1803B0A70,
                            qword_1803AFEE8,
                            0,
                            0,
                            v191,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A68, v192);
            if ( *v192 >= 0 )
            {
              v20 = (*(_QWORD *)v192)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v192 + 1) + 48LL))(v192);
            }
            v193 = (_DWORD *)qword_1803B09C8;
            v194 = sub_18036FB80(a1, qword_1803B0A78);
            if ( *v193 != -1 )
              ++*v193;
            v195 = (int *)sub_1803609A0(
                            (int)sub_18006ED60,
                            qword_1803B0A80,
                            qword_1803B0A88,
                            qword_1803AFF10,
                            (__int64)v193,
                            0,
                            v194,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A80, v195);
            if ( *v195 >= 0 )
            {
              v20 = (*(_QWORD *)v195)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v195 + 1) + 48LL))(v195);
            }
            v196 = sub_18036FB80(a1, qword_1803B0A90);
            v197 = (int *)sub_1803609A0(
                            (int)sub_18006F6F0,
                            qword_1803B0A98,
                            qword_1803B0AA0,
                            qword_1803AFEF0,
                            0,
                            0,
                            v196,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A98, v197);
            if ( *v197 >= 0 )
            {
              v20 = (*(_QWORD *)v197)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v197 + 1) + 48LL))(v197);
            }
            v198 = sub_18036FB80(a1, qword_1803B0AA8);
            v199 = (int *)sub_1803609A0(
                            (int)sub_180071030,
                            qword_1803B0AB0,
                            qword_1803B0AB8,
                            qword_1803AFD38,
                            0,
                            0,
                            v198,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AB0, v199);
            if ( *v199 >= 0 )
            {
              v20 = (*(_QWORD *)v199)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v199 + 1) + 48LL))(v199);
            }
            v200 = sub_18036FB80(a1, qword_1803B0AC0);
            v201 = (int *)sub_1803609A0(
                            (int)sub_1800717E0,
                            qword_1803B0AC8,
                            qword_1803B0AD0,
                            qword_1803AFB10,
                            0,
                            0,
                            v200,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AC8, v201);
            if ( *v201 >= 0 )
            {
              v20 = (*(_QWORD *)v201)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v201 + 1) + 48LL))(v201);
            }
            v202 = sub_18036FB80(a1, qword_1803B0AD8);
            v203 = (int *)sub_1803609A0(
                            (int)sub_180071F90,
                            qword_1803B0AE0,
                            qword_1803B0AE8,
                            qword_1803AFE48,
                            0,
                            0,
                            v202,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AE0, v203);
            if ( *v203 >= 0 )
            {
              v20 = (*(_QWORD *)v203)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v203 + 1) + 48LL))(v203);
            }
            v204 = sub_18036FB80(a1, qword_1803B0AF0);
            v205 = (int *)sub_1803609A0(
                            (int)sub_1800727B0,
                            qword_1803B0AF8,
                            qword_1803B0B00,
                            qword_1803AFDE8,
                            0,
                            0,
                            v204,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AF8, v205);
            if ( *v205 >= 0 )
            {
              v20 = (*(_QWORD *)v205)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v205 + 1) + 48LL))(v205);
            }
            v206 = (_DWORD *)qword_1803B09A0;
            v207 = sub_18036FB80(a1, qword_1803B0B08);
            if ( *v206 != -1 )
              ++*v206;
            v208 = (int *)sub_1803609A0(
                            (int)sub_180073DA0,
                            qword_1803B0B10,
                            qword_1803B0B18,
                            qword_1803AFC70,
                            (__int64)v206,
                            0,
                            v207,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B10, v208);
            if ( *v208 >= 0 )
            {
              v20 = (*(_QWORD *)v208)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v208 + 1) + 48LL))(v208);
            }
            v209 = sub_18036FB80(a1, qword_1803B0B20);
            v210 = (int *)sub_1803609A0(
                            (int)sub_180074F40,
                            qword_1803B0B28,
                            qword_1803B0B30,
                            qword_1803AFE28,
                            0,
                            0,
                            v209,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B28, v210);
            if ( *v210 >= 0 )
            {
              v20 = (*(_QWORD *)v210)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v210 + 1) + 48LL))(v210);
            }
            v211 = sub_18036FB80(a1, qword_1803B0B38);
            v212 = (int *)sub_1803609A0(
                            (int)sub_180075870,
                            qword_1803B0B40,
                            qword_1803B0B48,
                            qword_1803AFF28,
                            0,
                            0,
                            v211,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B40, v212);
            if ( *v212 >= 0 )
            {
              v20 = (*(_QWORD *)v212)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v212 + 1) + 48LL))(v212);
            }
            v213 = sub_18036FB80(a1, qword_1803B0A90);
            v214 = (int *)sub_1803609A0(
                            (int)sub_1800761F0,
                            qword_1803B0B50,
                            qword_1803B0B58,
                            qword_1803AFE58,
                            0,
                            0,
                            v213,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B50, v214);
            if ( *v214 >= 0 )
            {
              v20 = (*(_QWORD *)v214)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v214 + 1) + 48LL))(v214);
            }
            v215 = sub_18036FB80(a1, qword_1803B0AF0);
            v216 = (int *)sub_1803609A0(
                            (int)sub_180076750,
                            qword_1803B0B60,
                            qword_1803B0B68,
                            qword_1803AFEA0,
                            0,
                            0,
                            v215,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B60, v216);
            if ( *v216 >= 0 )
            {
              v20 = (*(_QWORD *)v216)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v216 + 1) + 48LL))(v216);
            }
            v217 = sub_18036FB80(a1, qword_1803B0A20);
            v218 = (int *)sub_1803609A0(
                            (int)sub_180077180,
                            qword_1803B0B70,
                            qword_1803B0B78,
                            qword_1803AFEE0,
                            0,
                            0,
                            v217,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B70, v218);
            if ( *v218 >= 0 )
            {
              v20 = (*(_QWORD *)v218)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v218 + 1) + 48LL))(v218);
            }
            v219 = sub_18036FB80(a1, qword_1803B0B80);
            v220 = (int *)sub_1803609A0(
                            (int)sub_180077890,
                            qword_1803B0B88,
                            qword_1803B0B90,
                            qword_1803AFD08,
                            0,
                            0,
                            v219,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B88, v220);
            if ( *v220 >= 0 )
            {
              v20 = (*(_QWORD *)v220)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v220 + 1) + 48LL))(v220);
            }
            v221 = sub_18036FB80(a1, qword_1803B0AF0);
            v222 = (int *)sub_1803609A0(
                            (int)sub_1800780B0,
                            qword_1803B0B98,
                            qword_1803B0BA0,
                            qword_1803AFDA8,
                            0,
                            0,
                            v221,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B98, v222);
            if ( *v222 >= 0 )
            {
              v20 = (*(_QWORD *)v222)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v222 + 1) + 48LL))(v222);
            }
            v223 = sub_18036FB80(a1, qword_1803B0B80);
            v224 = (int *)sub_1803609A0(
                            (int)sub_180079700,
                            qword_1803B0BA8,
                            qword_1803B0BB0,
                            qword_1803AFE80,
                            0,
                            0,
                            v223,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BA8, v224);
            if ( *v224 >= 0 )
            {
              v20 = (*(_QWORD *)v224)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v224 + 1) + 48LL))(v224);
            }
            v225 = sub_18036FB80(a1, qword_1803B0AF0);
            v226 = (int *)sub_1803609A0(
                            (int)sub_180079EF0,
                            qword_1803B0BB8,
                            qword_1803B0BC0,
                            qword_1803AFD18,
                            0,
                            0,
                            v225,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BB8, v226);
            if ( *v226 >= 0 )
            {
              v20 = (*(_QWORD *)v226)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v226 + 1) + 48LL))(v226);
            }
            v227 = sub_18036FB80(a1, qword_1803B0A20);
            v228 = (int *)sub_1803609A0(
                            (int)sub_18007B680,
                            qword_1803B0BC8,
                            qword_1803B0BD0,
                            qword_1803AFBE8,
                            0,
                            0,
                            v227,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BC8, v228);
            if ( *v228 >= 0 )
            {
              v20 = (*(_QWORD *)v228)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v228 + 1) + 48LL))(v228);
            }
            v229 = sub_18036FB80(a1, qword_1803B0AF0);
            v230 = (int *)sub_1803609A0(
                            (int)sub_18007BE90,
                            qword_1803B0BD8,
                            qword_1803B0BE0,
                            qword_1803AFDC0,
                            0,
                            0,
                            v229,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BD8, v230);
            if ( *v230 >= 0 )
            {
              v20 = (*(_QWORD *)v230)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v230 + 1) + 48LL))(v230);
            }
            v231 = sub_18036FB80(a1, qword_1803B0AF0);
            v232 = (int *)sub_1803609A0(
                            (int)sub_18007C840,
                            qword_1803B0BE8,
                            qword_1803B0BF0,
                            qword_1803AFB58,
                            0,
                            0,
                            v231,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BE8, v232);
            if ( *v232 >= 0 )
            {
              v20 = (*(_QWORD *)v232)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v232 + 1) + 48LL))(v232);
            }
            v233 = sub_18036FB80(a1, qword_1803B0AF0);
            v234 = (int *)sub_1803609A0(
                            (int)sub_18007F270,
                            qword_1803B0BF8,
                            qword_1803B0C00,
                            qword_1803AFBF0,
                            0,
                            0,
                            v233,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BF8, v234);
            if ( *v234 >= 0 )
            {
              v20 = (*(_QWORD *)v234)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v234 + 1) + 48LL))(v234);
            }
            v235 = sub_18036FB80(a1, qword_1803B0AF0);
            v236 = (int *)sub_1803609A0(
                            (int)sub_18007FC20,
                            qword_1803B0C08,
                            qword_1803B0C10,
                            qword_1803B0FD0,
                            0,
                            0,
                            v235,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C08, v236);
            if ( *v236 >= 0 )
            {
              v20 = (*(_QWORD *)v236)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v236 + 1) + 48LL))(v236);
            }
            v237 = sub_18036FB80(a1, qword_1803B0AF0);
            v238 = (int *)sub_1803609A0(
                            (int)sub_1800805D0,
                            qword_1803B0C18,
                            qword_1803B0C20,
                            qword_1803AFD28,
                            0,
                            0,
                            v237,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C18, v238);
            if ( *v238 >= 0 )
            {
              v20 = (*(_QWORD *)v238)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v238 + 1) + 48LL))(v238);
            }
            v239 = sub_18036FB80(a1, qword_1803B0A20);
            v240 = (int *)sub_1803609A0(
                            (int)sub_180081B90,
                            qword_1803B0C28,
                            qword_1803B0C30,
                            qword_1803AFBF8,
                            0,
                            0,
                            v239,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C28, v240);
            if ( *v240 >= 0 )
            {
              v20 = (*(_QWORD *)v240)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v240 + 1) + 48LL))(v240);
            }
            v241 = sub_18036FB80(a1, qword_1803B0AF0);
            v242 = (int *)sub_1803609A0(
                            (int)sub_1800823B0,
                            qword_1803B0C38,
                            qword_1803B0C40,
                            qword_1803AFF40,
                            0,
                            0,
                            v241,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C38, v242);
            if ( *v242 >= 0 )
            {
              v20 = (*(_QWORD *)v242)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v242 + 1) + 48LL))(v242);
            }
            v243 = sub_18036FB80(a1, qword_1803B0AF0);
            v244 = (int *)sub_1803609A0(
                            (int)sub_180082CB0,
                            qword_1803B0C48,
                            qword_1803B0C50,
                            qword_1803AFDF0,
                            0,
                            0,
                            v243,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C48, v244);
            if ( *v244 >= 0 )
            {
              v20 = (*(_QWORD *)v244)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v244 + 1) + 48LL))(v244);
            }
            v245 = sub_18036FB80(a1, qword_1803B0AF0);
            v246 = (int *)sub_1803609A0(
                            (int)sub_180084690,
                            qword_1803B0C58,
                            qword_1803B0C60,
                            qword_1803AFD30,
                            0,
                            0,
                            v245,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C58, v246);
            if ( *v246 >= 0 )
            {
              v20 = (*(_QWORD *)v246)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v246 + 1) + 48LL))(v246);
            }
            v247 = sub_18036FB80(a1, qword_1803B0C68);
            v248 = (int *)sub_1803609A0(
                            (int)sub_180085CA0,
                            qword_1803B0C70,
                            qword_1803B0C78,
                            qword_1803AFE18,
                            0,
                            0,
                            v247,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C70, v248);
            if ( *v248 >= 0 )
            {
              v20 = (*(_QWORD *)v248)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v248 + 1) + 48LL))(v248);
            }
            v249 = sub_18036FB80(a1, qword_1803B0C80);
            v250 = (int *)sub_1803609A0(
                            (int)sub_1800863B0,
                            qword_1803B0C88,
                            qword_1803B0C90,
                            qword_1803AFEC0,
                            0,
                            0,
                            v249,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C88, v250);
            if ( *v250 >= 0 )
            {
              v20 = (*(_QWORD *)v250)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v250 + 1) + 48LL))(v250);
            }
            v251 = (_DWORD *)qword_1803B09A0;
            v252 = sub_18036FB80(a1, qword_1803B0C98);
            if ( *v251 != -1 )
              ++*v251;
            v253 = (int *)sub_1803609A0(
                            (int)sub_180086D40,
                            qword_1803B0CA0,
                            qword_1803B0CA8,
                            qword_1803AFD98,
                            (__int64)v251,
                            0,
                            v252,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CA0, v253);
            if ( *v253 >= 0 )
            {
              v20 = (*(_QWORD *)v253)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v253 + 1) + 48LL))(v253);
            }
            v254 = sub_18036FB80(a1, qword_1803B0CB0);
            v255 = (int *)sub_1803609A0(
                            (int)sub_180087910,
                            qword_1803B0CB8,
                            qword_1803B0CC0,
                            qword_1803AFCA0,
                            0,
                            0,
                            v254,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CB8, v255);
            if ( *v255 >= 0 )
            {
              v20 = (*(_QWORD *)v255)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v255 + 1) + 48LL))(v255);
            }
            v256 = sub_18036FB80(a1, qword_1803B0CC8);
            v257 = (int *)sub_1803609A0(
                            (int)sub_1800880C0,
                            qword_1803B0CD0,
                            qword_1803B0CD8,
                            qword_1803AFC90,
                            0,
                            0,
                            v256,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CD0, v257);
            if ( *v257 >= 0 )
            {
              v20 = (*(_QWORD *)v257)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v257 + 1) + 48LL))(v257);
            }
            v258 = sub_18036FB80(a1, qword_1803B0A38);
            v259 = (int *)sub_1803609A0(
                            (int)sub_1800888E0,
                            qword_1803B0CE0,
                            qword_1803B0CE8,
                            qword_1803AFBE0,
                            0,
                            0,
                            v258,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CE0, v259);
            if ( *v259 >= 0 )
            {
              v20 = (*(_QWORD *)v259)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v259 + 1) + 48LL))(v259);
            }
            v260 = sub_18036FB80(a1, qword_1803B0CF0);
            v261 = (int *)sub_1803609A0(
                            (int)sub_180089120,
                            qword_1803B0CF8,
                            qword_1803B0D00,
                            qword_1803AFD80,
                            0,
                            0,
                            v260,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CF8, v261);
            if ( *v261 >= 0 )
            {
              v20 = (*(_QWORD *)v261)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v261 + 1) + 48LL))(v261);
            }
            v262 = sub_18036FB80(a1, qword_1803B0AF0);
            v263 = (int *)sub_1803609A0(
                            (int)sub_1800899F0,
                            qword_1803B0D08,
                            qword_1803B0D10,
                            qword_1803AFCC0,
                            0,
                            0,
                            v262,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D08, v263);
            if ( *v263 >= 0 )
            {
              v20 = (*(_QWORD *)v263)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v263 + 1) + 48LL))(v263);
            }
            v264 = sub_18036FB80(a1, qword_1803B0AF0);
            v265 = (int *)sub_1803609A0(
                            (int)sub_18008A790,
                            qword_1803B0D18,
                            qword_1803B0D20,
                            qword_1803AFEB0,
                            0,
                            0,
                            v264,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D18, v265);
            if ( *v265 >= 0 )
            {
              v20 = (*(_QWORD *)v265)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v265 + 1) + 48LL))(v265);
            }
            v266 = sub_18036FB80(a1, qword_1803B0D28);
            v267 = (int *)sub_1803609A0(
                            (int)sub_18008B840,
                            qword_1803B0770,
                            qword_1803B0D30,
                            qword_1803AFD10,
                            0,
                            0,
                            v266,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0770, v267);
            if ( *v267 >= 0 )
            {
              v20 = (*(_QWORD *)v267)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v267 + 1) + 48LL))(v267);
            }
            v268 = sub_18036FB80(a1, qword_1803B0AF0);
            v269 = (int *)sub_1803609A0(
                            (int)sub_18008C8A0,
                            qword_1803B0D38,
                            qword_1803B0D40,
                            qword_1803AFC28,
                            0,
                            0,
                            v268,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D38, v269);
            if ( *v269 >= 0 )
            {
              v20 = (*(_QWORD *)v269)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v269 + 1) + 48LL))(v269);
            }
            v270 = sub_18036FB80(a1, qword_1803B0AF0);
            v271 = (int *)sub_1803609A0(
                            (int)sub_18008DAD0,
                            qword_1803B0D48,
                            qword_1803B0D50,
                            qword_1803B0FA0,
                            0,
                            0,
                            v270,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D48, v271);
            if ( *v271 >= 0 )
            {
              v20 = (*(_QWORD *)v271)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v271 + 1) + 48LL))(v271);
            }
            v272 = sub_18036FB80(a1, qword_1803B0AF0);
            v273 = (int *)sub_1803609A0(
                            (int)sub_18008E1B0,
                            qword_1803B0D58,
                            qword_1803B0D60,
                            qword_1803AFD78,
                            0,
                            0,
                            v272,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D58, v273);
            if ( *v273 >= 0 )
            {
              v20 = (*(_QWORD *)v273)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v273 + 1) + 48LL))(v273);
            }
            v274 = sub_18036FB80(a1, qword_1803B0AF0);
            v275 = (int *)sub_1803609A0(
                            (int)sub_18008E9D0,
                            qword_1803B0D68,
                            qword_1803B0D70,
                            qword_1803AFBA8,
                            0,
                            0,
                            v274,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D68, v275);
            if ( *v275 >= 0 )
            {
              v20 = (*(_QWORD *)v275)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v275 + 1) + 48LL))(v275);
            }
            v276 = sub_18036FB80(a1, qword_1803B0AF0);
            v277 = (int *)sub_1803609A0(
                            (int)sub_18008EF80,
                            qword_1803B0D78,
                            qword_1803B0D80,
                            qword_1803AFC98,
                            0,
                            0,
                            v276,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D78, v277);
            if ( *v277 >= 0 )
            {
              v20 = (*(_QWORD *)v277)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v277 + 1) + 48LL))(v277);
            }
            v278 = sub_18036FB80(a1, qword_1803B0AF0);
            v279 = (int *)sub_1803609A0(
                            (int)sub_18008FB70,
                            qword_1803B0D88,
                            qword_1803B0D90,
                            qword_1803AFAE8,
                            0,
                            0,
                            v278,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D88, v279);
            if ( *v279 >= 0 )
            {
              v20 = (*(_QWORD *)v279)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v279 + 1) + 48LL))(v279);
            }
            v280 = (_DWORD *)qword_1803B0D98;
            v281 = sub_18036FB80(a1, qword_1803B0DA0);
            if ( *v280 != -1 )
              ++*v280;
            v282 = (int *)sub_1803609A0(
                            (int)sub_180090E60,
                            qword_1803B0DA8,
                            qword_1803B0DB0,
                            qword_1803AFCF8,
                            (__int64)v280,
                            0,
                            v281,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0DA8, v282);
            if ( *v282 >= 0 )
            {
              v20 = (*(_QWORD *)v282)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v282 + 1) + 48LL))(v282);
            }
            v283 = a1[2];
            v284 = (_DWORD *)qword_1803B0948;
            v285 = *(int **)(v283 + 269736);
            if ( v285 != nullptr )
            {
              *(_QWORD *)(v283 + 269736) = *((_QWORD *)v285 + 3);
              --*(_DWORD *)(v283 + 269888);
              *(_QWORD *)v285 = 1;
            }
            else
            {
              v285 = (int *)sub_18035EBD0(PyTuple_Type, 3);
            }
            *((_QWORD *)v285 + 3) = 0;
            *((_QWORD *)v285 + 4) = 0;
            *((_QWORD *)v285 + 5) = 0;
            sub_1801F33C0(v285);
            if ( *v284 != -1 )
              ++*v284;
            *((_QWORD *)v285 + 3) = v284;
            v286 = qword_1803AB240;
            if ( *(_DWORD *)qword_1803AB240 != -1 )
              ++*(_DWORD *)qword_1803AB240;
            *((_QWORD *)v285 + 4) = v286;
            if ( *(_DWORD *)v156 != -1 )
              ++*(_DWORD *)v156;
            *((_QWORD *)v285 + 5) = v156;
            *(_DWORD *)(v30 + 40) = 31;
            v288 = sub_180001790(a1, PyType_Type, v285, v155);
            if ( *v285 >= 0 )
            {
              v20 = (*(_QWORD *)v285)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v285 + 1) + 48LL))(v285);
            }
            if ( v288 == 0 )
            {
              v36 = a1[12];
              a1[12] = 0;
              if ( *(int *)v156 >= 0 )
              {
                v20 = (*(_QWORD *)v156)-- == 1;
                if ( v20 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v156 + 8) + 48LL))(v156);
              }
              if ( *v155 >= 0 )
              {
                v20 = (*(_QWORD *)v155)-- == 1;
                if ( v20 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v155 + 1) + 48LL))(v155);
              }
              if ( *(int *)v156 >= 0 )
              {
                v20 = (*(_QWORD *)v156)-- == 1;
                if ( v20 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v156 + 8) + 48LL))(v156);
              }
              v37 = 31;
LABEL_462:
              v289 = *(_QWORD *)(v36 + 40);
              if ( v289 != 0 )
              {
                if ( *(_QWORD *)(v289 + 24) == v30 )
                {
LABEL_481:
                  v293 = a1[7];
                  v294 = *(_QWORD *)(*(_QWORD *)v293 + 48LL);
                  *(_QWORD *)v293 = *(_QWORD *)(*(_QWORD *)v293 + 8LL);
                  *(_DWORD *)(v294 + 64) = -1;
                  v295 = *(int **)(v294 + 16);
                  if ( v295 != nullptr )
                  {
                    *(_QWORD *)(v294 + 16) = 0;
                    if ( *v295 >= 0 )
                    {
                      v20 = (*(_QWORD *)v295)-- == 1;
                      if ( v20 )
                        (*(void (**)(void))(*((_QWORD *)v295 + 1) + 48LL))();
                    }
                  }
                  if ( *(int *)v294 >= 0 )
                  {
                    v20 = (*(_QWORD *)v294)-- == 1;
                    if ( v20 )
                      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v294 + 8) + 48LL))(v294);
                  }
                  *(_QWORD *)(v294 + 80) = 0;
                  v296 = (int *)a1[12];
                  a1[12] = v36;
                  if ( v296 != nullptr && *v296 >= 0 )
                  {
                    v20 = (*(_QWORD *)v296)-- == 1;
                    if ( v20 )
                      (*(void (**)(void))(*((_QWORD *)v296 + 1) + 48LL))();
                  }
                  return 0;
                }
                v290 = sub_18038E2A0(v30, v37);
                v291 = (int *)v290;
                *(_QWORD *)(v290 + 16) = v289;
                if ( *(_DWORD *)v289 != -1 )
                  ++*(_DWORD *)v289;
                v292 = *(int **)(v36 + 40);
                if ( v290 == Py_NoneStruct )
                {
                  v290 = 0;
LABEL_473:
                  *(_QWORD *)(v36 + 40) = v290;
                  if ( v292 != nullptr && *v292 >= 0 )
                  {
                    v20 = (*(_QWORD *)v292)-- == 1;
                    if ( v20 )
                      (*(void (**)(void))(*((_QWORD *)v292 + 1) + 48LL))();
                  }
                  if ( v291 != nullptr && *v291 >= 0 )
                  {
                    v20 = (*(_QWORD *)v291)-- == 1;
                    if ( v20 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v291 + 1) + 48LL))(v291);
                  }
                  goto LABEL_481;
                }
              }
              else
              {
                v290 = sub_18038E2A0(v30, v37);
                v291 = (int *)v290;
                v292 = *(int **)(v36 + 40);
                if ( v290 == Py_NoneStruct )
                {
                  v290 = 0;
                  goto LABEL_473;
                }
              }
              if ( v291 != nullptr && *(_DWORD *)v290 != -1 )
                ++*(_DWORD *)v290;
              goto LABEL_473;
            }
            if ( *(_DWORD *)v288 != -1 )
              ++*(_DWORD *)v288;
            if ( *(int *)v156 >= 0 )
            {
              v20 = (*(_QWORD *)v156)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(__int64))(*(_QWORD *)(v156 + 8) + 48LL))(v156);
            }
            if ( *(int *)v288 >= 0 )
            {
              v20 = (*(_QWORD *)v288)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(__int64))(*(_QWORD *)(v288 + 8) + 48LL))(v288);
            }
            sub_180065D20(v287, qword_1803B0948, v288);
            sub_180001D80(a1, v298);
            if ( *v155 >= 0 )
            {
              v20 = (*(_QWORD *)v155)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v155 + 1) + 48LL))(v155);
            }
            if ( *(int *)v156 >= 0 )
            {
              v20 = (*(_QWORD *)v156)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(__int64))(*(_QWORD *)(v156 + 8) + 48LL))(v156);
            }
            result = qword_1803CA828;
            if ( *(_DWORD *)qword_1803CA828 != -1 )
              ++*(_DWORD *)qword_1803CA828;
            return result;
          }
        }
      }
    }
  }
  v37 = 16;
  goto LABEL_75;
}
