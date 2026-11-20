// Analyst annotations only; original export: source/ida/0x180091e60.c
// qword comments derive module constant slot + original file range; not runtime values.
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
// Original IDA line 330 | qword_1803AFF50 -> novels.api.novels_api#0 bytes [3924440, 3924450]: "host_api"
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
// Original IDA line 403 | qword_1803B0830 -> novels.api.novels_api#284 bytes [3929643, 3929652]: "__doc__"
  sub_180065C90(v26, qword_1803B0830, Py_NoneStruct);
// Original IDA line 404 | qword_1803B0838 -> novels.api.novels_api#285 bytes [3929652, 3929662]: "__file__"
  sub_180065C90(v27, qword_1803B0838, qword_1803AFF30);
  v28 = qword_1803CA828;
  v29 = *(_DWORD **)(qword_1803CA828 + 16);
  if ( *v29 != -1 )
    ++*v29;
  v30 = sub_18036CBB0(qword_1803AFE30, v28, v29, 0);
  sub_180001D00(a1, v30);
  v31 = qword_1803AFF30;
// Original IDA line 412 | qword_1803B0F98 -> novels.api.novels_api#521 bytes [3938075, 3938085]: "__spec__"
  v32 = (__int64 *)sub_1800015A0(qword_1803CA830, qword_1803B0F98);
  if ( v32 == nullptr || (v34 = *v32, *v32 == 0) )
  {
// Original IDA line 415 | qword_1803B0F98 -> novels.api.novels_api#521 bytes [3938075, 3938085]: "__spec__"
    v35 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B0F98);
    v34 = 0;
    if ( v35 != nullptr )
      v34 = *v35;
  }
// Original IDA line 420 | qword_1803B0840 -> novels.api.novels_api#286 bytes [3929662, 3929670]: "origin"
  if ( (unsigned __int8)sub_18037F3A0(v33, v34, qword_1803B0840, v31) == 0
// Original IDA line 421 | qword_1803B0848 -> novels.api.novels_api#287 bytes [3929670, 3929684]: "has_location"
    || (v38 = sub_180066560(), (unsigned __int8)sub_18037F3A0(v39, v38, qword_1803B0848, Py_TrueStruct) == 0) )
  {
    v36 = a1[12];
    v37 = 1;
    a1[12] = 0;
    goto LABEL_462;
  }
// Original IDA line 428 | qword_1803B0850 -> novels.api.novels_api#288 bytes [3929684, 3929696]: "__cached__"
  sub_180065C90(v40, qword_1803B0850, Py_NoneStruct);
  v41 = sub_180372070();
// Original IDA line 430 | qword_1803B0858 -> novels.api.novels_api#289 bytes [3929696, 3929709]: "annotations"
  v43 = sub_18037F0C0(v42, v41, qword_1803B0858);
// Original IDA line 431 | qword_1803B0858 -> novels.api.novels_api#289 bytes [3929696, 3929709]: "annotations"
  sub_180065D20(v44, qword_1803B0858, v43);
// Original IDA line 432 | qword_1803AFF58 -> novels.api.novels_api#1 bytes [3924450, 3924461]: "threading"
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
// Original IDA line 444 | qword_1803AFF58 -> novels.api.novels_api#1 bytes [3924450, 3924461]: "threading"
  sub_180065D20(v48, qword_1803AFF58, v47);
// Original IDA line 445 | qword_1803B0490 -> novels.api.novels_api#168 bytes [3927564, 3927570]: "time"
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
// Original IDA line 457 | qword_1803B0490 -> novels.api.novels_api#168 bytes [3927564, 3927570]: "time"
  sub_180065D20(v52, qword_1803B0490, v51);
// Original IDA line 458 | qword_1803B0860 -> novels.api.novels_api#290 bytes [3929709, 3929718]: "pathlib"
  v53 = qword_1803B0860;
  v54 = qword_1803CA830;
  v313 = qword_1803AB250;
// Original IDA line 461 | qword_1803B0868 -> novels.api.novels_api#291 bytes [3929718, 3929726]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3929720, 3929726], "value": "Path"}]}
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
// Original IDA line 473 | qword_1803AFFA8 -> novels.api.novels_api#11 bytes [3924565, 3924571]: "Path"
    v57 = sub_180371D80((_DWORD)a1, (_DWORD)v56, qword_1803CA830, qword_1803AFFA8, qword_1803AB250);
  else
// Original IDA line 475 | qword_1803AFFA8 -> novels.api.novels_api#11 bytes [3924565, 3924571]: "Path"
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
// Original IDA line 490 | qword_1803AFFA8 -> novels.api.novels_api#11 bytes [3924565, 3924571]: "Path"
  sub_180065D20(v58, qword_1803AFFA8, v59);
  v60 = sub_180372110();
  v61 = (int *)v60;
  if ( *(_DWORD *)v60 != -1 )
    ++*(_DWORD *)v60;
  if ( *(_QWORD *)(v60 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 496 | qword_1803B0870 -> novels.api.novels_api#292 bytes [3929726, 3929731]: "Any"
    v62 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0870, qword_1803AB250);
  else
// Original IDA line 498 | qword_1803B0870 -> novels.api.novels_api#292 bytes [3929726, 3929731]: "Any"
    v62 = sub_180371890(a1, v61, qword_1803B0870);
  if ( v62 == 0
// Original IDA line 500 | qword_1803B0870 -> novels.api.novels_api#292 bytes [3929726, 3929731]: "Any"
    || ((sub_180065D20(v63, qword_1803B0870, v62), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 502 | qword_1803B0878 -> novels.api.novels_api#293 bytes [3929731, 3929741]: "Callable"
      ? (v64 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0878, qword_1803AB250))
// Original IDA line 503 | qword_1803B0878 -> novels.api.novels_api#293 bytes [3929731, 3929741]: "Callable"
      : (v64 = sub_180371890(a1, v61, qword_1803B0878)),
        v64 == 0
// Original IDA line 505 | qword_1803B0878 -> novels.api.novels_api#293 bytes [3929731, 3929741]: "Callable"
     || ((sub_180065D20(v65, qword_1803B0878, v64), *((_QWORD *)v61 + 1) == PyModule_Type)
      || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 507 | qword_1803B0880 -> novels.api.novels_api#294 bytes [3929741, 3929747]: "Dict"
       ? (v66 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0880, qword_1803AB250))
// Original IDA line 508 | qword_1803B0880 -> novels.api.novels_api#294 bytes [3929741, 3929747]: "Dict"
       : (v66 = sub_180371890(a1, v61, qword_1803B0880)),
         v66 == 0
// Original IDA line 510 | qword_1803B0880 -> novels.api.novels_api#294 bytes [3929741, 3929747]: "Dict"
      || ((sub_180065D20(v67, qword_1803B0880, v66), *((_QWORD *)v61 + 1) == PyModule_Type)
       || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 512 | qword_1803B0888 -> novels.api.novels_api#295 bytes [3929747, 3929757]: "Optional"
        ? (v68 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0888, qword_1803AB250))
// Original IDA line 513 | qword_1803B0888 -> novels.api.novels_api#295 bytes [3929747, 3929757]: "Optional"
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
// Original IDA line 528 | qword_1803B0888 -> novels.api.novels_api#295 bytes [3929747, 3929757]: "Optional"
  sub_180065D20(v69, qword_1803B0888, v68);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
// Original IDA line 535 | qword_1803B0890 -> novels.api.novels_api#296 bytes [3929757, 3929773]: "models.schemas"
  v70 = qword_1803B0890;
  v71 = qword_1803CA830;
// Original IDA line 537 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v314 = qword_1803B0288;
// Original IDA line 538 | qword_1803B0898 -> novels.api.novels_api#297 bytes [3929773, 3929803]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3929775, 3929790], "value": "SessionRecord"}, {"tag": "a", "fileRange": [3929790, 3929803], "value": "utc_now_iso"}]}
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
// Original IDA line 550 | qword_1803B02E0 -> novels.api.novels_api#114 bytes [3926556, 3926571]: "SessionRecord"
    v73 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B02E0, qword_1803AB250);
  else
// Original IDA line 552 | qword_1803B02E0 -> novels.api.novels_api#114 bytes [3926556, 3926571]: "SessionRecord"
    v73 = sub_180371890(a1, v61, qword_1803B02E0);
  if ( v73 == 0
// Original IDA line 554 | qword_1803B02E0 -> novels.api.novels_api#114 bytes [3926556, 3926571]: "SessionRecord"
    || ((sub_180065D20(v74, qword_1803B02E0, v73), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 556 | qword_1803B0500 -> novels.api.novels_api#182 bytes [3927697, 3927710]: "utc_now_iso"
      ? (v75 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0500, qword_1803AB250))
// Original IDA line 557 | qword_1803B0500 -> novels.api.novels_api#182 bytes [3927697, 3927710]: "utc_now_iso"
      : (v75 = sub_180371890(a1, v61, qword_1803B0500)),
        v75 == 0) )
  {
    v37 = 8;
    goto LABEL_75;
  }
// Original IDA line 563 | qword_1803B0500 -> novels.api.novels_api#182 bytes [3927697, 3927710]: "utc_now_iso"
  sub_180065D20(v76, qword_1803B0500, v75);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
// Original IDA line 570 | qword_1803B08A0 -> novels.api.novels_api#298 bytes [3929803, 3929829]: "pipeline.pipeline_runner"
  v77 = qword_1803B08A0;
  v78 = qword_1803CA830;
// Original IDA line 572 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v315 = qword_1803B0288;
// Original IDA line 573 | qword_1803B08A8 -> novels.api.novels_api#299 bytes [3929829, 3929847]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3929831, 3929847], "value": "PipelineRunner"}]}
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
// Original IDA line 585 | qword_1803B00A0 -> novels.api.novels_api#42 bytes [3925079, 3925095]: "PipelineRunner"
    v81 = sub_180371D80((_DWORD)a1, (_DWORD)v80, qword_1803CA830, qword_1803B00A0, qword_1803AB250);
  else
// Original IDA line 587 | qword_1803B00A0 -> novels.api.novels_api#42 bytes [3925079, 3925095]: "PipelineRunner"
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
// Original IDA line 602 | qword_1803B00A0 -> novels.api.novels_api#42 bytes [3925079, 3925095]: "PipelineRunner"
  sub_180065D20(v82, qword_1803B00A0, v83);
// Original IDA line 603 | qword_1803B08B0 -> novels.api.novels_api#300 bytes [3929847, 3929871]: "services.agent_session"
  v84 = qword_1803B08B0;
  v85 = qword_1803CA830;
// Original IDA line 605 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v316 = qword_1803B0288;
// Original IDA line 606 | qword_1803B08B8 -> novels.api.novels_api#301 bytes [3929871, 3929894]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3929873, 3929894], "value": "AgentSessionService"}]}
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
// Original IDA line 618 | qword_1803B00C8 -> novels.api.novels_api#47 bytes [3925231, 3925252]: "AgentSessionService"
    v88 = sub_180371D80((_DWORD)a1, (_DWORD)v87, qword_1803CA830, qword_1803B00C8, qword_1803AB250);
  else
// Original IDA line 620 | qword_1803B00C8 -> novels.api.novels_api#47 bytes [3925231, 3925252]: "AgentSessionService"
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
// Original IDA line 635 | qword_1803B00C8 -> novels.api.novels_api#47 bytes [3925231, 3925252]: "AgentSessionService"
  sub_180065D20(v89, qword_1803B00C8, v90);
// Original IDA line 636 | qword_1803B08C0 -> novels.api.novels_api#302 bytes [3929894, 3929919]: "services.agent_settings"
  v91 = qword_1803B08C0;
  v92 = qword_1803CA830;
// Original IDA line 638 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v317 = qword_1803B0288;
// Original IDA line 639 | qword_1803B08C8 -> novels.api.novels_api#303 bytes [3929919, 3929978]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3929921, 3929934], "value": "AGENT_NODES"}, {"tag": "a", "fileRange": [3929934, 3929954], "value": "AgentSettingsStore"}, {"tag": "a", "fileRange": [3929954, 3929978], "value": "default... (truncated literal)
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
// Original IDA line 651 | qword_1803B05C8 -> novels.api.novels_api#207 bytes [3928141, 3928154]: "AGENT_NODES"
    v94 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B05C8, qword_1803AB250);
  else
// Original IDA line 653 | qword_1803B05C8 -> novels.api.novels_api#207 bytes [3928141, 3928154]: "AGENT_NODES"
    v94 = sub_180371890(a1, v61, qword_1803B05C8);
  if ( v94 == 0
// Original IDA line 655 | qword_1803B05C8 -> novels.api.novels_api#207 bytes [3928141, 3928154]: "AGENT_NODES"
    || ((sub_180065D20(v95, qword_1803B05C8, v94), *((_QWORD *)v61 + 1) == PyModule_Type)
     || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 657 | qword_1803B0050 -> novels.api.novels_api#32 bytes [3924911, 3924931]: "AgentSettingsStore"
      ? (v96 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0050, qword_1803AB250))
// Original IDA line 658 | qword_1803B0050 -> novels.api.novels_api#32 bytes [3924911, 3924931]: "AgentSettingsStore"
      : (v96 = sub_180371890(a1, v61, qword_1803B0050)),
        v96 == 0
// Original IDA line 660 | qword_1803B0050 -> novels.api.novels_api#32 bytes [3924911, 3924931]: "AgentSettingsStore"
     || ((sub_180065D20(v97, qword_1803B0050, v96), *((_QWORD *)v61 + 1) == PyModule_Type)
      || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 662 | qword_1803B05D8 -> novels.api.novels_api#209 bytes [3928160, 3928184]: "default_agent_settings"
       ? (v98 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B05D8, qword_1803AB250))
// Original IDA line 663 | qword_1803B05D8 -> novels.api.novels_api#209 bytes [3928160, 3928184]: "default_agent_settings"
       : (v98 = sub_180371890(a1, v61, qword_1803B05D8)),
         v98 == 0)) )
  {
    v37 = 11;
    goto LABEL_75;
  }
// Original IDA line 669 | qword_1803B05D8 -> novels.api.novels_api#209 bytes [3928160, 3928184]: "default_agent_settings"
  sub_180065D20(v99, qword_1803B05D8, v98);
  if ( *v61 >= 0 )
  {
    v20 = (*(_QWORD *)v61)-- == 1;
    if ( v20 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
  }
// Original IDA line 676 | qword_1803B08D0 -> novels.api.novels_api#304 bytes [3929978, 3930002]: "services.style_library"
  v100 = qword_1803B08D0;
  v101 = qword_1803CA830;
// Original IDA line 678 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v318 = qword_1803B0288;
// Original IDA line 679 | qword_1803B08D8 -> novels.api.novels_api#305 bytes [3930002, 3930104]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930004, 3930027], "value": "append_imported_files"}, {"tag": "a", "fileRange": [3930027, 3930049], "value": "delete_style_profile"}, {"tag": "a", "fileRange": [3930049, 3930068], "valu... (truncated literal)
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
// Original IDA line 691 | qword_1803B0408 -> novels.api.novels_api#151 bytes [3927158, 3927181]: "append_imported_files"
    v103 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0408, qword_1803AB250);
  else
// Original IDA line 693 | qword_1803B0408 -> novels.api.novels_api#151 bytes [3927158, 3927181]: "append_imported_files"
    v103 = sub_180371890(a1, v61, qword_1803B0408);
  if ( v103 != 0 )
  {
// Original IDA line 696 | qword_1803B0408 -> novels.api.novels_api#151 bytes [3927158, 3927181]: "append_imported_files"
    sub_180065D20(v104, qword_1803B0408, v103);
    if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 698 | qword_1803B03C0 -> novels.api.novels_api#142 bytes [3927046, 3927068]: "delete_style_profile"
      v105 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B03C0, qword_1803AB250);
    else
// Original IDA line 700 | qword_1803B03C0 -> novels.api.novels_api#142 bytes [3927046, 3927068]: "delete_style_profile"
      v105 = sub_180371890(a1, v61, qword_1803B03C0);
    if ( v105 != 0 )
    {
// Original IDA line 703 | qword_1803B03C0 -> novels.api.novels_api#142 bytes [3927046, 3927068]: "delete_style_profile"
      sub_180065D20(v106, qword_1803B03C0, v105);
      if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 705 | qword_1803B0340 -> novels.api.novels_api#126 bytes [3926767, 3926786]: "get_style_profile"
        v107 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0340, qword_1803AB250);
      else
// Original IDA line 707 | qword_1803B0340 -> novels.api.novels_api#126 bytes [3926767, 3926786]: "get_style_profile"
        v107 = sub_180371890(a1, v61, qword_1803B0340);
      if ( v107 != 0 )
      {
// Original IDA line 710 | qword_1803B0340 -> novels.api.novels_api#126 bytes [3926767, 3926786]: "get_style_profile"
        sub_180065D20(v108, qword_1803B0340, v107);
        if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 712 | qword_1803B03F8 -> novels.api.novels_api#149 bytes [3927136, 3927152]: "read_text_file"
          v109 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B03F8, qword_1803AB250);
        else
// Original IDA line 714 | qword_1803B03F8 -> novels.api.novels_api#149 bytes [3927136, 3927152]: "read_text_file"
          v109 = sub_180371890(a1, v61, qword_1803B03F8);
        if ( v109 != 0 )
        {
// Original IDA line 717 | qword_1803B03F8 -> novels.api.novels_api#149 bytes [3927136, 3927152]: "read_text_file"
          sub_180065D20(v110, qword_1803B03F8, v109);
          if ( *((_QWORD *)v61 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 719 | qword_1803B0390 -> novels.api.novels_api#136 bytes [3926901, 3926921]: "save_style_profile"
            v111 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0390, qword_1803AB250);
          else
// Original IDA line 721 | qword_1803B0390 -> novels.api.novels_api#136 bytes [3926901, 3926921]: "save_style_profile"
            v111 = sub_180371890(a1, v61, qword_1803B0390);
          if ( v111 != 0 )
          {
// Original IDA line 724 | qword_1803B0390 -> novels.api.novels_api#136 bytes [3926901, 3926921]: "save_style_profile"
            sub_180065D20(v112, qword_1803B0390, v111);
            if ( *v61 >= 0 )
            {
              v20 = (*(_QWORD *)v61)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
            }
// Original IDA line 731 | qword_1803B08E0 -> novels.api.novels_api#306 bytes [3930104, 3930128]: "agents.architect_agent"
            v113 = qword_1803B08E0;
            v114 = qword_1803CA830;
// Original IDA line 733 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v319 = qword_1803B0288;
// Original IDA line 734 | qword_1803B08E8 -> novels.api.novels_api#307 bytes [3930128, 3930146]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930130, 3930146], "value": "ArchitectAgent"}]}
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
// Original IDA line 746 | qword_1803B0368 -> novels.api.novels_api#131 bytes [3926830, 3926846]: "ArchitectAgent"
              v117 = sub_180371D80((_DWORD)a1, (_DWORD)v116, qword_1803CA830, qword_1803B0368, qword_1803AB250);
            else
// Original IDA line 748 | qword_1803B0368 -> novels.api.novels_api#131 bytes [3926830, 3926846]: "ArchitectAgent"
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
// Original IDA line 763 | qword_1803B0368 -> novels.api.novels_api#131 bytes [3926830, 3926846]: "ArchitectAgent"
            sub_180065D20(v118, qword_1803B0368, v119);
// Original IDA line 764 | qword_1803B08F0 -> novels.api.novels_api#308 bytes [3930146, 3930171]: "storage.file_repository"
            v120 = qword_1803B08F0;
            v121 = qword_1803CA830;
// Original IDA line 766 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v320 = qword_1803B0288;
// Original IDA line 767 | qword_1803B08F8 -> novels.api.novels_api#309 bytes [3930171, 3930189]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930173, 3930189], "value": "FileRepository"}]}
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
// Original IDA line 779 | qword_1803AFFB8 -> novels.api.novels_api#13 bytes [3924586, 3924602]: "FileRepository"
              v124 = sub_180371D80((_DWORD)a1, (_DWORD)v123, qword_1803CA830, qword_1803AFFB8, qword_1803AB250);
            else
// Original IDA line 781 | qword_1803AFFB8 -> novels.api.novels_api#13 bytes [3924586, 3924602]: "FileRepository"
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
// Original IDA line 796 | qword_1803AFFB8 -> novels.api.novels_api#13 bytes [3924586, 3924602]: "FileRepository"
            sub_180065D20(v125, qword_1803AFFB8, v126);
// Original IDA line 797 | qword_1803B0900 -> novels.api.novels_api#310 bytes [3930189, 3930212]: "storage.state_manager"
            v127 = qword_1803B0900;
            v128 = qword_1803CA830;
// Original IDA line 799 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v321 = qword_1803B0288;
// Original IDA line 800 | qword_1803B0908 -> novels.api.novels_api#311 bytes [3930212, 3930234]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930214, 3930234], "value": "NovelsStateManager"}]}
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
// Original IDA line 812 | qword_1803AFFE0 -> novels.api.novels_api#18 bytes [3924646, 3924666]: "NovelsStateManager"
              v131 = sub_180371D80((_DWORD)a1, (_DWORD)v130, qword_1803CA830, qword_1803AFFE0, qword_1803AB250);
            else
// Original IDA line 814 | qword_1803AFFE0 -> novels.api.novels_api#18 bytes [3924646, 3924666]: "NovelsStateManager"
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
// Original IDA line 829 | qword_1803AFFE0 -> novels.api.novels_api#18 bytes [3924646, 3924666]: "NovelsStateManager"
            sub_180065D20(v132, qword_1803AFFE0, v133);
// Original IDA line 830 | qword_1803B0910 -> novels.api.novels_api#312 bytes [3930234, 3930254]: "services.scheduler"
            v134 = qword_1803B0910;
            v135 = qword_1803CA830;
// Original IDA line 832 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v322 = qword_1803B0288;
// Original IDA line 833 | qword_1803B0918 -> novels.api.novels_api#313 bytes [3930254, 3930289]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930256, 3930272], "value": "NovelScheduler"}, {"tag": "a", "fileRange": [3930272, 3930289], "value": "SchedulerConfig"}]}
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
// Original IDA line 845 | qword_1803B0740 -> novels.api.novels_api#254 bytes [3929192, 3929208]: "NovelScheduler"
              v137 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0740, qword_1803AB250);
            else
// Original IDA line 847 | qword_1803B0740 -> novels.api.novels_api#254 bytes [3929192, 3929208]: "NovelScheduler"
              v137 = sub_180371890(a1, v61, qword_1803B0740);
            if ( v137 == 0
// Original IDA line 849 | qword_1803B0740 -> novels.api.novels_api#254 bytes [3929192, 3929208]: "NovelScheduler"
              || ((sub_180065D20(v138, qword_1803B0740, v137), *((_QWORD *)v61 + 1) == PyModule_Type)
               || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 851 | qword_1803B0748 -> novels.api.novels_api#255 bytes [3929208, 3929225]: "SchedulerConfig"
                ? (v139 = sub_180371D80((_DWORD)a1, (_DWORD)v61, qword_1803CA830, qword_1803B0748, qword_1803AB250))
// Original IDA line 852 | qword_1803B0748 -> novels.api.novels_api#255 bytes [3929208, 3929225]: "SchedulerConfig"
                : (v139 = sub_180371890(a1, v61, qword_1803B0748)),
                  v139 == 0) )
            {
              v37 = 26;
              goto LABEL_75;
            }
// Original IDA line 858 | qword_1803B0748 -> novels.api.novels_api#255 bytes [3929208, 3929225]: "SchedulerConfig"
            sub_180065D20(v140, qword_1803B0748, v139);
            if ( *v61 >= 0 )
            {
              v20 = (*(_QWORD *)v61)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
            }
// Original IDA line 865 | qword_1803B0920 -> novels.api.novels_api#314 bytes [3930289, 3930309]: "utils.cancellation"
            v141 = qword_1803B0920;
            v142 = qword_1803CA830;
// Original IDA line 867 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v323 = qword_1803B0288;
// Original IDA line 868 | qword_1803B0928 -> novels.api.novels_api#315 bytes [3930309, 3930336]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930311, 3930336], "value": "OperationCancelledError"}]}
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
// Original IDA line 880 | qword_1803B0268 -> novels.api.novels_api#99 bytes [3926135, 3926160]: "OperationCancelledError"
              v145 = sub_180371D80((_DWORD)a1, (_DWORD)v144, qword_1803CA830, qword_1803B0268, qword_1803AB250);
            else
// Original IDA line 882 | qword_1803B0268 -> novels.api.novels_api#99 bytes [3926135, 3926160]: "OperationCancelledError"
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
// Original IDA line 897 | qword_1803B0268 -> novels.api.novels_api#99 bytes [3926135, 3926160]: "OperationCancelledError"
            sub_180065D20(v146, qword_1803B0268, v147);
// Original IDA line 898 | qword_1803B0930 -> novels.api.novels_api#316 bytes [3930336, 3930360]: "utils.knowledge_bridge"
            v148 = qword_1803B0930;
            v149 = qword_1803CA830;
// Original IDA line 900 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v324 = qword_1803B0288;
// Original IDA line 901 | qword_1803B0938 -> novels.api.novels_api#317 bytes [3930360, 3930382]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3930362, 3930382], "value": "list_style_sources"}]}
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
// Original IDA line 913 | qword_1803B0338 -> novels.api.novels_api#125 bytes [3926747, 3926767]: "list_style_sources"
              v152 = sub_180371D80((_DWORD)a1, (_DWORD)v151, qword_1803CA830, qword_1803B0338, qword_1803AB250);
            else
// Original IDA line 915 | qword_1803B0338 -> novels.api.novels_api#125 bytes [3926747, 3926767]: "list_style_sources"
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
// Original IDA line 930 | qword_1803B0338 -> novels.api.novels_api#125 bytes [3926747, 3926767]: "list_style_sources"
            sub_180065D20(v153, qword_1803B0338, v154);
            v155 = (int *)sub_180370EC0(a1);
            v156 = sub_180370EC0(a1);
            if ( *(_DWORD *)v156 != -1 )
              ++*(_DWORD *)v156;
// Original IDA line 935 | qword_1803B0F90 -> novels.api.novels_api#520 bytes [3938065, 3938075]: "__name__"
            v157 = (__int64 *)sub_1800015A0(qword_1803CA830, qword_1803B0F90);
            if ( v157 == nullptr || (v158 = *v157, *v157 == 0) )
            {
// Original IDA line 938 | qword_1803B0F90 -> novels.api.novels_api#520 bytes [3938065, 3938075]: "__name__"
              v159 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B0F90);
              v158 = 0;
              if ( v159 != nullptr )
                v158 = *v159;
            }
// Original IDA line 943 | qword_1803B0940 -> novels.api.novels_api#318 bytes [3930382, 3930394]: "__module__"
            PyDict_SetItem(v156, qword_1803B0940, v158);
// Original IDA line 944 | qword_1803B0948 -> novels.api.novels_api#319 bytes [3930394, 3930405]: "NovelsApi"; qword_1803B0950 -> novels.api.novels_api#320 bytes [3930405, 3930419]: "__qualname__"
            PyDict_SetItem(v156, qword_1803B0950, qword_1803B0948);
            v160 = (int *)sub_1803609A0(
// Original IDA line 946 | sub_180067D50 -> NovelsApi.__init__
                            (int)sub_180067D50,
// Original IDA line 947 | qword_1803B0958 -> novels.api.novels_api#321 bytes [3930419, 3930429]: "__init__"
                            qword_1803B0958,
// Original IDA line 948 | qword_1803B0960 -> novels.api.novels_api#322 bytes [3930429, 3930449]: "NovelsApi.__init__"
                            qword_1803B0960,
                            qword_1803AFE90,
                            0,
                            0,
                            0,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 957 | qword_1803B0958 -> novels.api.novels_api#321 bytes [3930419, 3930429]: "__init__"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0958, v160);
            if ( *v160 >= 0 )
            {
              v20 = (*(_QWORD *)v160)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v160 + 1) + 48LL))(v160);
            }
// Original IDA line 964 | qword_1803B0968 -> novels.api.novels_api#323 bytes [3930449, 3930465]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930451, 3930459], "value": "return"}], "values": [{"tag": "a", "fileRange": [3930459, 3930465], "value": "Path"}]}
            v161 = sub_18036FB80(a1, qword_1803B0968);
            v162 = (int *)sub_1803609A0(
// Original IDA line 966 | sub_1800681B0 -> NovelsApi._data_root
                            (int)sub_1800681B0,
// Original IDA line 967 | qword_1803AFFE8 -> novels.api.novels_api#19 bytes [3924666, 3924678]: "_data_root"
                            qword_1803AFFE8,
// Original IDA line 968 | qword_1803B0970 -> novels.api.novels_api#324 bytes [3930465, 3930487]: "NovelsApi._data_root"
                            qword_1803B0970,
                            qword_1803AFC60,
                            0,
                            0,
                            v161,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 977 | qword_1803AFFE8 -> novels.api.novels_api#19 bytes [3924666, 3924678]: "_data_root"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803AFFE8, v162);
            if ( *v162 >= 0 )
            {
              v20 = (*(_QWORD *)v162)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v162 + 1) + 48LL))(v162);
            }
// Original IDA line 984 | qword_1803B0978 -> novels.api.novels_api#325 bytes [3930487, 3930502]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930489, 3930497], "value": "return"}], "values": [{"tag": "a", "fileRange": [3930497, 3930502], "value": "str"}]}
            v163 = sub_18036FB80(a1, qword_1803B0978);
            v164 = (int *)sub_1803609A0(
// Original IDA line 986 | sub_1800688D0 -> NovelsApi._data_dir
                            (int)sub_1800688D0,
// Original IDA line 987 | qword_1803AFFA0 -> novels.api.novels_api#10 bytes [3924554, 3924565]: "_data_dir"
                            qword_1803AFFA0,
// Original IDA line 988 | qword_1803B0980 -> novels.api.novels_api#326 bytes [3930502, 3930523]: "NovelsApi._data_dir"
                            qword_1803B0980,
                            qword_1803AFD50,
                            0,
                            0,
                            v163,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 997 | qword_1803AFFA0 -> novels.api.novels_api#10 bytes [3924554, 3924565]: "_data_dir"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803AFFA0, v164);
            if ( *v164 >= 0 )
            {
              v20 = (*(_QWORD *)v164)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v164 + 1) + 48LL))(v164);
            }
// Original IDA line 1004 | qword_1803B0988 -> novels.api.novels_api#327 bytes [3930523, 3930553]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930525, 3930533], "value": "return"}], "values": [{"tag": "a", "fileRange": [3930533, 3930553], "value": "NovelsStateManager"}]}
            v165 = sub_18036FB80(a1, qword_1803B0988);
            v166 = (int *)sub_1803609A0(
// Original IDA line 1006 | sub_180068CE0 -> NovelsApi._state
                            (int)sub_180068CE0,
// Original IDA line 1007 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
                            qword_1803B0058,
// Original IDA line 1008 | qword_1803B0990 -> novels.api.novels_api#328 bytes [3930553, 3930571]: "NovelsApi._state"
                            qword_1803B0990,
                            qword_1803AFC80,
                            0,
                            0,
                            v165,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1017 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0058, v166);
            if ( *v166 >= 0 )
            {
              v20 = (*(_QWORD *)v166)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v166 + 1) + 48LL))(v166);
            }
            v167 = (int *)sub_1803609A0(
// Original IDA line 1025 | sub_1800690A0 -> NovelsApi._settings_provider
                            (int)sub_1800690A0,
// Original IDA line 1026 | qword_1803B00A8 -> novels.api.novels_api#43 bytes [3925095, 3925115]: "_settings_provider"
                            qword_1803B00A8,
// Original IDA line 1027 | qword_1803B0998 -> novels.api.novels_api#329 bytes [3930571, 3930601]: "NovelsApi._settings_provider"
                            qword_1803B0998,
                            qword_1803AFCB0,
                            0,
                            0,
                            0,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1036 | qword_1803B00A8 -> novels.api.novels_api#43 bytes [3925095, 3925115]: "_settings_provider"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00A8, v167);
            if ( *v167 >= 0 )
            {
              v20 = (*(_QWORD *)v167)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v167 + 1) + 48LL))(v167);
            }
// Original IDA line 1043 | qword_1803B09A0 -> novels.api.novels_api#330 bytes [3930601, 3930604]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930603, 3930604], "value": null}]}
            v168 = (_DWORD *)qword_1803B09A0;
// Original IDA line 1044 | qword_1803B09A8 -> novels.api.novels_api#331 bytes [3930604, 3930676]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930606, 3930618], "value": "session_id"}, {"tag": "a", "fileRange": [3930618, 3930627], "value": "book_id"}, {"tag": "a", "fileRange": [3930627, 3930635], "value": "return"}], "values": ... (truncated literal)
            v169 = sub_18036FB80(a1, qword_1803B09A8);
            if ( *v168 != -1 )
              ++*v168;
            v170 = (int *)sub_1803609A0(
// Original IDA line 1048 | sub_180069340 -> NovelsApi._event_emitter
                            (int)sub_180069340,
// Original IDA line 1049 | qword_1803B00B8 -> novels.api.novels_api#45 bytes [3925130, 3925146]: "_event_emitter"
                            qword_1803B00B8,
// Original IDA line 1050 | qword_1803B09B0 -> novels.api.novels_api#332 bytes [3930676, 3930702]: "NovelsApi._event_emitter"
                            qword_1803B09B0,
                            qword_1803AFF38,
                            (__int64)v168,
                            0,
                            v169,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1059 | qword_1803B00B8 -> novels.api.novels_api#45 bytes [3925130, 3925146]: "_event_emitter"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00B8, v170);
            if ( *v170 >= 0 )
            {
              v20 = (*(_QWORD *)v170)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v170 + 1) + 48LL))(v170);
            }
// Original IDA line 1066 | qword_1803B09B8 -> novels.api.novels_api#333 bytes [3930702, 3930766]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930704, 3930712], "value": "prompt"}, {"tag": "a", "fileRange": [3930712, 3930719], "value": "title"}, {"tag": "a", "fileRange": [3930719, 3930731], "value": "output_dir"}, {"tag": "a", ... (truncated literal)
            v171 = sub_18036FB80(a1, qword_1803B09B8);
            v172 = (int *)sub_1803609A0(
// Original IDA line 1068 | sub_180069B90 -> NovelsApi._cover_image_generator
                            (int)sub_180069B90,
// Original IDA line 1069 | qword_1803B00E0 -> novels.api.novels_api#50 bytes [3925281, 3925305]: "_cover_image_generator"
                            qword_1803B00E0,
// Original IDA line 1070 | qword_1803B09C0 -> novels.api.novels_api#334 bytes [3930766, 3930800]: "NovelsApi._cover_image_generator"
                            qword_1803B09C0,
                            qword_1803AFE38,
                            0,
                            0,
                            v171,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1079 | qword_1803B00E0 -> novels.api.novels_api#50 bytes [3925281, 3925305]: "_cover_image_generator"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00E0, v172);
            if ( *v172 >= 0 )
            {
              v20 = (*(_QWORD *)v172)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v172 + 1) + 48LL))(v172);
            }
// Original IDA line 1086 | qword_1803B09C8 -> novels.api.novels_api#335 bytes [3930800, 3930804]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930802, 3930803], "value": null}, {"tag": "n", "fileRange": [3930803, 3930804], "value": null}]}
            v173 = (_DWORD *)qword_1803B09C8;
// Original IDA line 1087 | qword_1803B09D0 -> novels.api.novels_api#336 bytes [3930804, 3930902]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930806, 3930818], "value": "session_id"}, {"tag": "a", "fileRange": [3930818, 3930827], "value": "book_id"}, {"tag": "a", "fileRange": [3930827, 3930843], "value": "channel_effort"}, {"t... (truncated literal)
            v174 = sub_18036FB80(a1, qword_1803B09D0);
            if ( *v173 != -1 )
              ++*v173;
            v175 = (int *)sub_1803609A0(
// Original IDA line 1091 | sub_18006A620 -> NovelsApi._pipeline
                            (int)sub_18006A620,
// Original IDA line 1092 | qword_1803B00D0 -> novels.api.novels_api#48 bytes [3925252, 3925263]: "_pipeline"
                            qword_1803B00D0,
// Original IDA line 1093 | qword_1803B09D8 -> novels.api.novels_api#337 bytes [3930902, 3930923]: "NovelsApi._pipeline"
                            qword_1803B09D8,
                            qword_1803AFF00,
                            (__int64)v173,
                            0,
                            v174,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1102 | qword_1803B00D0 -> novels.api.novels_api#48 bytes [3925252, 3925263]: "_pipeline"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00D0, v175);
            if ( *v175 >= 0 )
            {
              v20 = (*(_QWORD *)v175)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v175 + 1) + 48LL))(v175);
            }
// Original IDA line 1109 | qword_1803B09C8 -> novels.api.novels_api#335 bytes [3930800, 3930804]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930802, 3930803], "value": null}, {"tag": "n", "fileRange": [3930803, 3930804], "value": null}]}
            v176 = (_DWORD *)qword_1803B09C8;
// Original IDA line 1110 | qword_1803B09E0 -> novels.api.novels_api#338 bytes [3930923, 3931026]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3930925, 3930937], "value": "session_id"}, {"tag": "a", "fileRange": [3930937, 3930946], "value": "book_id"}, {"tag": "a", "fileRange": [3930946, 3930962], "value": "channel_effort"}, {"t... (truncated literal)
            v177 = sub_18036FB80(a1, qword_1803B09E0);
            if ( *v176 != -1 )
              ++*v176;
            v178 = (int *)sub_1803609A0(
// Original IDA line 1114 | sub_18006AC60 -> NovelsApi._agent_service
                            (int)sub_18006AC60,
// Original IDA line 1115 | qword_1803B0248 -> novels.api.novels_api#95 bytes [3926073, 3926089]: "_agent_service"
                            qword_1803B0248,
// Original IDA line 1116 | qword_1803B09E8 -> novels.api.novels_api#339 bytes [3931026, 3931052]: "NovelsApi._agent_service"
                            qword_1803B09E8,
                            qword_1803AFD48,
                            (__int64)v176,
                            0,
                            v177,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1125 | qword_1803B0248 -> novels.api.novels_api#95 bytes [3926073, 3926089]: "_agent_service"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0248, v178);
            if ( *v178 >= 0 )
            {
              v20 = (*(_QWORD *)v178)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v178 + 1) + 48LL))(v178);
            }
// Original IDA line 1132 | qword_1803B09F0 -> novels.api.novels_api#340 bytes [3931052, 3931091]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931054, 3931059], "value": "key"}, {"tag": "a", "fileRange": [3931059, 3931066], "value": "value"}, {"tag": "a", "fileRange": [3931066, 3931074], "value": "return"}], "values": [{"tag": ... (truncated literal)
            v179 = sub_18036FB80(a1, qword_1803B09F0);
            v180 = (int *)sub_1803609A0(
// Original IDA line 1134 | sub_18006B5E0 -> NovelsApi._set_cancel
                            (int)sub_18006B5E0,
// Original IDA line 1135 | qword_1803B0628 -> novels.api.novels_api#219 bytes [3928485, 3928498]: "_set_cancel"
                            qword_1803B0628,
// Original IDA line 1136 | qword_1803B09F8 -> novels.api.novels_api#341 bytes [3931091, 3931114]: "NovelsApi._set_cancel"
                            qword_1803B09F8,
                            qword_1803AFB70,
                            0,
                            0,
                            v179,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1145 | qword_1803B0628 -> novels.api.novels_api#219 bytes [3928485, 3928498]: "_set_cancel"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0628, v180);
            if ( *v180 >= 0 )
            {
              v20 = (*(_QWORD *)v180)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v180 + 1) + 48LL))(v180);
            }
// Original IDA line 1152 | qword_1803B0A00 -> novels.api.novels_api#342 bytes [3931114, 3931140]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931116, 3931121], "value": "key"}, {"tag": "a", "fileRange": [3931121, 3931129], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931129, 3931134], "value": "str"}, {"tag": "a... (truncated literal)
            v181 = sub_18036FB80(a1, qword_1803B0A00);
            v182 = (int *)sub_1803609A0(
// Original IDA line 1154 | sub_18006BD80 -> NovelsApi._clear_cancel
                            (int)sub_18006BD80,
// Original IDA line 1155 | qword_1803B0220 -> novels.api.novels_api#90 bytes [3925956, 3925971]: "_clear_cancel"
                            qword_1803B0220,
// Original IDA line 1156 | qword_1803B0A08 -> novels.api.novels_api#343 bytes [3931140, 3931165]: "NovelsApi._clear_cancel"
                            qword_1803B0A08,
                            qword_1803AFB20,
                            0,
                            0,
                            v181,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1165 | qword_1803B0220 -> novels.api.novels_api#90 bytes [3925956, 3925971]: "_clear_cancel"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0220, v182);
            if ( *v182 >= 0 )
            {
              v20 = (*(_QWORD *)v182)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v182 + 1) + 48LL))(v182);
            }
// Original IDA line 1172 | qword_1803B0A10 -> novels.api.novels_api#344 bytes [3931165, 3931191]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931167, 3931172], "value": "key"}, {"tag": "a", "fileRange": [3931172, 3931180], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931180, 3931185], "value": "str"}, {"tag": "a... (truncated literal)
            v183 = sub_18036FB80(a1, qword_1803B0A10);
            v184 = (int *)sub_1803609A0(
// Original IDA line 1174 | sub_18006C590 -> NovelsApi._is_cancelled
                            (int)sub_18006C590,
// Original IDA line 1175 | qword_1803B00B0 -> novels.api.novels_api#44 bytes [3925115, 3925130]: "_is_cancelled"
                            qword_1803B00B0,
// Original IDA line 1176 | qword_1803B0A18 -> novels.api.novels_api#345 bytes [3931191, 3931216]: "NovelsApi._is_cancelled"
                            qword_1803B0A18,
                            qword_1803AFD58,
                            0,
                            0,
                            v183,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1185 | qword_1803B00B0 -> novels.api.novels_api#44 bytes [3925115, 3925130]: "_is_cancelled"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B00B0, v184);
            if ( *v184 >= 0 )
            {
              v20 = (*(_QWORD *)v184)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v184 + 1) + 48LL))(v184);
            }
// Original IDA line 1192 | qword_1803B0A20 -> novels.api.novels_api#346 bytes [3931216, 3931242]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931218, 3931226], "value": "return"}], "values": [{"tag": "u", "fileRange": [3931226, 3931242], "value": "Dict[str, Any]"}]}
            v185 = sub_18036FB80(a1, qword_1803B0A20);
            v186 = (int *)sub_1803609A0(
// Original IDA line 1194 | sub_18006CDE0 -> NovelsApi.novels_list_books
                            (int)sub_18006CDE0,
// Original IDA line 1195 | qword_1803B0A28 -> novels.api.novels_api#347 bytes [3931242, 3931261]: "novels_list_books"
                            qword_1803B0A28,
// Original IDA line 1196 | qword_1803B0A30 -> novels.api.novels_api#348 bytes [3931261, 3931290]: "NovelsApi.novels_list_books"
                            qword_1803B0A30,
                            qword_1803AFF18,
                            0,
                            0,
                            v185,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1205 | qword_1803B0A28 -> novels.api.novels_api#347 bytes [3931242, 3931261]: "novels_list_books"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A28, v186);
            if ( *v186 >= 0 )
            {
              v20 = (*(_QWORD *)v186)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v186 + 1) + 48LL))(v186);
            }
// Original IDA line 1212 | qword_1803B0A38 -> novels.api.novels_api#349 bytes [3931290, 3931330]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931292, 3931301], "value": "book_id"}, {"tag": "a", "fileRange": [3931301, 3931309], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931309, 3931314], "value": "str"}, {"tag"... (truncated literal)
            v187 = sub_18036FB80(a1, qword_1803B0A38);
            v188 = (int *)sub_1803609A0(
// Original IDA line 1214 | sub_18006D730 -> NovelsApi.novels_get_book
                            (int)sub_18006D730,
// Original IDA line 1215 | qword_1803B0A40 -> novels.api.novels_api#350 bytes [3931330, 3931347]: "novels_get_book"
                            qword_1803B0A40,
// Original IDA line 1216 | qword_1803B0A48 -> novels.api.novels_api#351 bytes [3931347, 3931374]: "NovelsApi.novels_get_book"
                            qword_1803B0A48,
                            qword_1803AFB50,
                            0,
                            0,
                            v187,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1225 | qword_1803B0A40 -> novels.api.novels_api#350 bytes [3931330, 3931347]: "novels_get_book"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A40, v188);
            if ( *v188 >= 0 )
            {
              v20 = (*(_QWORD *)v188)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v188 + 1) + 48LL))(v188);
            }
// Original IDA line 1232 | qword_1803B0A50 -> novels.api.novels_api#352 bytes [3931374, 3931422]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931376, 3931385], "value": "book_id"}, {"tag": "a", "fileRange": [3931385, 3931392], "value": "title"}, {"tag": "a", "fileRange": [3931392, 3931400], "value": "return"}], "values": [{"ta... (truncated literal)
            v189 = sub_18036FB80(a1, qword_1803B0A50);
            v190 = (int *)sub_1803609A0(
// Original IDA line 1234 | sub_18006DEE0 -> NovelsApi.novels_rename_book
                            (int)sub_18006DEE0,
// Original IDA line 1235 | qword_1803B0A58 -> novels.api.novels_api#353 bytes [3931422, 3931442]: "novels_rename_book"
                            qword_1803B0A58,
// Original IDA line 1236 | qword_1803B0A60 -> novels.api.novels_api#354 bytes [3931442, 3931472]: "NovelsApi.novels_rename_book"
                            qword_1803B0A60,
                            qword_1803B0FC8,
                            0,
                            0,
                            v189,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1245 | qword_1803B0A58 -> novels.api.novels_api#353 bytes [3931422, 3931442]: "novels_rename_book"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A58, v190);
            if ( *v190 >= 0 )
            {
              v20 = (*(_QWORD *)v190)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v190 + 1) + 48LL))(v190);
            }
// Original IDA line 1252 | qword_1803B0A38 -> novels.api.novels_api#349 bytes [3931290, 3931330]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931292, 3931301], "value": "book_id"}, {"tag": "a", "fileRange": [3931301, 3931309], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931309, 3931314], "value": "str"}, {"tag"... (truncated literal)
            v191 = sub_18036FB80(a1, qword_1803B0A38);
            v192 = (int *)sub_1803609A0(
// Original IDA line 1254 | sub_18006E800 -> NovelsApi.novels_delete_book
                            (int)sub_18006E800,
// Original IDA line 1255 | qword_1803B0A68 -> novels.api.novels_api#355 bytes [3931472, 3931492]: "novels_delete_book"
                            qword_1803B0A68,
// Original IDA line 1256 | qword_1803B0A70 -> novels.api.novels_api#356 bytes [3931492, 3931522]: "NovelsApi.novels_delete_book"
                            qword_1803B0A70,
                            qword_1803AFEE8,
                            0,
                            0,
                            v191,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1265 | qword_1803B0A68 -> novels.api.novels_api#355 bytes [3931472, 3931492]: "novels_delete_book"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A68, v192);
            if ( *v192 >= 0 )
            {
              v20 = (*(_QWORD *)v192)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v192 + 1) + 48LL))(v192);
            }
// Original IDA line 1272 | qword_1803B09C8 -> novels.api.novels_api#335 bytes [3930800, 3930804]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930802, 3930803], "value": null}, {"tag": "n", "fileRange": [3930803, 3930804], "value": null}]}
            v193 = (_DWORD *)qword_1803B09C8;
// Original IDA line 1273 | qword_1803B0A78 -> novels.api.novels_api#357 bytes [3931522, 3931597]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931524, 3931533], "value": "book_id"}, {"tag": "a", "fileRange": [3931533, 3931543], "value": "workflow"}, {"tag": "a", "fileRange": [3931543, 3931551], "value": "return"}], "values": [{... (truncated literal)
            v194 = sub_18036FB80(a1, qword_1803B0A78);
            if ( *v193 != -1 )
              ++*v193;
            v195 = (int *)sub_1803609A0(
// Original IDA line 1277 | sub_18006ED60 -> NovelsApi.novels_list_sessions
                            (int)sub_18006ED60,
// Original IDA line 1278 | qword_1803B0A80 -> novels.api.novels_api#358 bytes [3931597, 3931619]: "novels_list_sessions"
                            qword_1803B0A80,
// Original IDA line 1279 | qword_1803B0A88 -> novels.api.novels_api#359 bytes [3931619, 3931651]: "NovelsApi.novels_list_sessions"
                            qword_1803B0A88,
                            qword_1803AFF10,
                            (__int64)v193,
                            0,
                            v194,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1288 | qword_1803B0A80 -> novels.api.novels_api#358 bytes [3931597, 3931619]: "novels_list_sessions"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A80, v195);
            if ( *v195 >= 0 )
            {
              v20 = (*(_QWORD *)v195)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v195 + 1) + 48LL))(v195);
            }
// Original IDA line 1295 | qword_1803B0A90 -> novels.api.novels_api#360 bytes [3931651, 3931694]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931653, 3931665], "value": "session_id"}, {"tag": "a", "fileRange": [3931665, 3931673], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931673, 3931678], "value": "str"}, {"t... (truncated literal)
            v196 = sub_18036FB80(a1, qword_1803B0A90);
            v197 = (int *)sub_1803609A0(
// Original IDA line 1297 | sub_18006F6F0 -> NovelsApi.novels_get_session
                            (int)sub_18006F6F0,
// Original IDA line 1298 | qword_1803B0A98 -> novels.api.novels_api#361 bytes [3931694, 3931714]: "novels_get_session"
                            qword_1803B0A98,
// Original IDA line 1299 | qword_1803B0AA0 -> novels.api.novels_api#362 bytes [3931714, 3931744]: "NovelsApi.novels_get_session"
                            qword_1803B0AA0,
                            qword_1803AFEF0,
                            0,
                            0,
                            v196,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1308 | qword_1803B0A98 -> novels.api.novels_api#361 bytes [3931694, 3931714]: "novels_get_session"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0A98, v197);
            if ( *v197 >= 0 )
            {
              v20 = (*(_QWORD *)v197)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v197 + 1) + 48LL))(v197);
            }
// Original IDA line 1315 | qword_1803B0AA8 -> novels.api.novels_api#363 bytes [3931744, 3931785]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931746, 3931756], "value": "story_id"}, {"tag": "a", "fileRange": [3931756, 3931764], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931764, 3931769], "value": "str"}, {"tag... (truncated literal)
            v198 = sub_18036FB80(a1, qword_1803B0AA8);
            v199 = (int *)sub_1803609A0(
// Original IDA line 1317 | sub_180071030 -> NovelsApi.novels_get_short_story
                            (int)sub_180071030,
// Original IDA line 1318 | qword_1803B0AB0 -> novels.api.novels_api#364 bytes [3931785, 3931809]: "novels_get_short_story"
                            qword_1803B0AB0,
// Original IDA line 1319 | qword_1803B0AB8 -> novels.api.novels_api#365 bytes [3931809, 3931843]: "NovelsApi.novels_get_short_story"
                            qword_1803B0AB8,
                            qword_1803AFD38,
                            0,
                            0,
                            v198,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1328 | qword_1803B0AB0 -> novels.api.novels_api#364 bytes [3931785, 3931809]: "novels_get_short_story"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AB0, v199);
            if ( *v199 >= 0 )
            {
              v20 = (*(_QWORD *)v199)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v199 + 1) + 48LL))(v199);
            }
// Original IDA line 1335 | qword_1803B0AC0 -> novels.api.novels_api#366 bytes [3931843, 3931890]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931845, 3931855], "value": "story_id"}, {"tag": "a", "fileRange": [3931855, 3931860], "value": "key"}, {"tag": "a", "fileRange": [3931860, 3931868], "value": "return"}], "values": [{"tag... (truncated literal)
            v200 = sub_18036FB80(a1, qword_1803B0AC0);
            v201 = (int *)sub_1803609A0(
// Original IDA line 1337 | sub_1800717E0 -> NovelsApi.novels_read_short_story_file
                            (int)sub_1800717E0,
// Original IDA line 1338 | qword_1803B0AC8 -> novels.api.novels_api#367 bytes [3931890, 3931920]: "novels_read_short_story_file"
                            qword_1803B0AC8,
// Original IDA line 1339 | qword_1803B0AD0 -> novels.api.novels_api#368 bytes [3931920, 3931960]: "NovelsApi.novels_read_short_story_file"
                            qword_1803B0AD0,
                            qword_1803AFB10,
                            0,
                            0,
                            v200,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1348 | qword_1803B0AC8 -> novels.api.novels_api#367 bytes [3931890, 3931920]: "novels_read_short_story_file"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AC8, v201);
            if ( *v201 >= 0 )
            {
              v20 = (*(_QWORD *)v201)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v201 + 1) + 48LL))(v201);
            }
// Original IDA line 1355 | qword_1803B0AD8 -> novels.api.novels_api#369 bytes [3931960, 3932017]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931962, 3931972], "value": "story_id"}, {"tag": "a", "fileRange": [3931972, 3931977], "value": "key"}, {"tag": "a", "fileRange": [3931977, 3931986], "value": "content"}, {"tag": "a", "fi... (truncated literal)
            v202 = sub_18036FB80(a1, qword_1803B0AD8);
            v203 = (int *)sub_1803609A0(
// Original IDA line 1357 | sub_180071F90 -> NovelsApi.novels_save_short_story_file
                            (int)sub_180071F90,
// Original IDA line 1358 | qword_1803B0AE0 -> novels.api.novels_api#370 bytes [3932017, 3932047]: "novels_save_short_story_file"
                            qword_1803B0AE0,
// Original IDA line 1359 | qword_1803B0AE8 -> novels.api.novels_api#371 bytes [3932047, 3932087]: "NovelsApi.novels_save_short_story_file"
                            qword_1803B0AE8,
                            qword_1803AFE48,
                            0,
                            0,
                            v202,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1368 | qword_1803B0AE0 -> novels.api.novels_api#370 bytes [3932017, 3932047]: "novels_save_short_story_file"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AE0, v203);
            if ( *v203 >= 0 )
            {
              v20 = (*(_QWORD *)v203)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v203 + 1) + 48LL))(v203);
            }
// Original IDA line 1375 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v204 = sub_18036FB80(a1, qword_1803B0AF0);
            v205 = (int *)sub_1803609A0(
// Original IDA line 1377 | sub_1800727B0 -> NovelsApi.novels_confirm_pipeline_checkpoint
                            (int)sub_1800727B0,
// Original IDA line 1378 | qword_1803B0AF8 -> novels.api.novels_api#373 bytes [3932138, 3932174]: "novels_confirm_pipeline_checkpoint"
                            qword_1803B0AF8,
// Original IDA line 1379 | qword_1803B0B00 -> novels.api.novels_api#374 bytes [3932174, 3932220]: "NovelsApi.novels_confirm_pipeline_checkpoint"
                            qword_1803B0B00,
                            qword_1803AFDE8,
                            0,
                            0,
                            v204,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1388 | qword_1803B0AF8 -> novels.api.novels_api#373 bytes [3932138, 3932174]: "novels_confirm_pipeline_checkpoint"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0AF8, v205);
            if ( *v205 >= 0 )
            {
              v20 = (*(_QWORD *)v205)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v205 + 1) + 48LL))(v205);
            }
// Original IDA line 1395 | qword_1803B09A0 -> novels.api.novels_api#330 bytes [3930601, 3930604]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930603, 3930604], "value": null}]}
            v206 = (_DWORD *)qword_1803B09A0;
// Original IDA line 1396 | qword_1803B0B08 -> novels.api.novels_api#375 bytes [3932220, 3932281]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932222, 3932231], "value": "payload"}, {"tag": "a", "fileRange": [3932231, 3932239], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932239, 3932265], "value": "Optional[Dict... (truncated literal)
            v207 = sub_18036FB80(a1, qword_1803B0B08);
            if ( *v206 != -1 )
              ++*v206;
            v208 = (int *)sub_1803609A0(
// Original IDA line 1400 | sub_180073DA0 -> NovelsApi.novels_create_session
                            (int)sub_180073DA0,
// Original IDA line 1401 | qword_1803B0B10 -> novels.api.novels_api#376 bytes [3932281, 3932304]: "novels_create_session"
                            qword_1803B0B10,
// Original IDA line 1402 | qword_1803B0B18 -> novels.api.novels_api#377 bytes [3932304, 3932337]: "NovelsApi.novels_create_session"
                            qword_1803B0B18,
                            qword_1803AFC70,
                            (__int64)v206,
                            0,
                            v207,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1411 | qword_1803B0B10 -> novels.api.novels_api#376 bytes [3932281, 3932304]: "novels_create_session"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B10, v208);
            if ( *v208 >= 0 )
            {
              v20 = (*(_QWORD *)v208)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v208 + 1) + 48LL))(v208);
            }
// Original IDA line 1418 | qword_1803B0B20 -> novels.api.novels_api#378 bytes [3932337, 3932388]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932339, 3932351], "value": "session_id"}, {"tag": "a", "fileRange": [3932351, 3932358], "value": "title"}, {"tag": "a", "fileRange": [3932358, 3932366], "value": "return"}], "values": [{... (truncated literal)
            v209 = sub_18036FB80(a1, qword_1803B0B20);
            v210 = (int *)sub_1803609A0(
// Original IDA line 1420 | sub_180074F40 -> NovelsApi.novels_rename_session
                            (int)sub_180074F40,
// Original IDA line 1421 | qword_1803B0B28 -> novels.api.novels_api#379 bytes [3932388, 3932411]: "novels_rename_session"
                            qword_1803B0B28,
// Original IDA line 1422 | qword_1803B0B30 -> novels.api.novels_api#380 bytes [3932411, 3932444]: "NovelsApi.novels_rename_session"
                            qword_1803B0B30,
                            qword_1803AFE28,
                            0,
                            0,
                            v209,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1431 | qword_1803B0B28 -> novels.api.novels_api#379 bytes [3932388, 3932411]: "novels_rename_session"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B28, v210);
            if ( *v210 >= 0 )
            {
              v20 = (*(_QWORD *)v210)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v210 + 1) + 48LL))(v210);
            }
// Original IDA line 1438 | qword_1803B0B38 -> novels.api.novels_api#381 bytes [3932444, 3932497]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932446, 3932458], "value": "session_id"}, {"tag": "a", "fileRange": [3932458, 3932467], "value": "book_id"}, {"tag": "a", "fileRange": [3932467, 3932475], "value": "return"}], "values": ... (truncated literal)
            v211 = sub_18036FB80(a1, qword_1803B0B38);
            v212 = (int *)sub_1803609A0(
// Original IDA line 1440 | sub_180075870 -> NovelsApi.novels_bind_session_book
                            (int)sub_180075870,
// Original IDA line 1441 | qword_1803B0B40 -> novels.api.novels_api#382 bytes [3932497, 3932523]: "novels_bind_session_book"
                            qword_1803B0B40,
// Original IDA line 1442 | qword_1803B0B48 -> novels.api.novels_api#383 bytes [3932523, 3932559]: "NovelsApi.novels_bind_session_book"
                            qword_1803B0B48,
                            qword_1803AFF28,
                            0,
                            0,
                            v211,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1451 | qword_1803B0B40 -> novels.api.novels_api#382 bytes [3932497, 3932523]: "novels_bind_session_book"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B40, v212);
            if ( *v212 >= 0 )
            {
              v20 = (*(_QWORD *)v212)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v212 + 1) + 48LL))(v212);
            }
// Original IDA line 1458 | qword_1803B0A90 -> novels.api.novels_api#360 bytes [3931651, 3931694]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931653, 3931665], "value": "session_id"}, {"tag": "a", "fileRange": [3931665, 3931673], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931673, 3931678], "value": "str"}, {"t... (truncated literal)
            v213 = sub_18036FB80(a1, qword_1803B0A90);
            v214 = (int *)sub_1803609A0(
// Original IDA line 1460 | sub_1800761F0 -> NovelsApi.novels_delete_session
                            (int)sub_1800761F0,
// Original IDA line 1461 | qword_1803B0B50 -> novels.api.novels_api#384 bytes [3932559, 3932582]: "novels_delete_session"
                            qword_1803B0B50,
// Original IDA line 1462 | qword_1803B0B58 -> novels.api.novels_api#385 bytes [3932582, 3932615]: "NovelsApi.novels_delete_session"
                            qword_1803B0B58,
                            qword_1803AFE58,
                            0,
                            0,
                            v213,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1471 | qword_1803B0B50 -> novels.api.novels_api#384 bytes [3932559, 3932582]: "novels_delete_session"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B50, v214);
            if ( *v214 >= 0 )
            {
              v20 = (*(_QWORD *)v214)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v214 + 1) + 48LL))(v214);
            }
// Original IDA line 1478 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v215 = sub_18036FB80(a1, qword_1803B0AF0);
            v216 = (int *)sub_1803609A0(
// Original IDA line 1480 | sub_180076750 -> NovelsApi.novels_update_session_messages
                            (int)sub_180076750,
// Original IDA line 1481 | qword_1803B0B60 -> novels.api.novels_api#386 bytes [3932615, 3932647]: "novels_update_session_messages"
                            qword_1803B0B60,
// Original IDA line 1482 | qword_1803B0B68 -> novels.api.novels_api#387 bytes [3932647, 3932689]: "NovelsApi.novels_update_session_messages"
                            qword_1803B0B68,
                            qword_1803AFEA0,
                            0,
                            0,
                            v215,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1491 | qword_1803B0B60 -> novels.api.novels_api#386 bytes [3932615, 3932647]: "novels_update_session_messages"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B60, v216);
            if ( *v216 >= 0 )
            {
              v20 = (*(_QWORD *)v216)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v216 + 1) + 48LL))(v216);
            }
// Original IDA line 1498 | qword_1803B0A20 -> novels.api.novels_api#346 bytes [3931216, 3931242]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931218, 3931226], "value": "return"}], "values": [{"tag": "u", "fileRange": [3931226, 3931242], "value": "Dict[str, Any]"}]}
            v217 = sub_18036FB80(a1, qword_1803B0A20);
            v218 = (int *)sub_1803609A0(
// Original IDA line 1500 | sub_180077180 -> NovelsApi.novels_list_style_sources
                            (int)sub_180077180,
// Original IDA line 1501 | qword_1803B0B70 -> novels.api.novels_api#388 bytes [3932689, 3932716]: "novels_list_style_sources"
                            qword_1803B0B70,
// Original IDA line 1502 | qword_1803B0B78 -> novels.api.novels_api#389 bytes [3932716, 3932753]: "NovelsApi.novels_list_style_sources"
                            qword_1803B0B78,
                            qword_1803AFEE0,
                            0,
                            0,
                            v217,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1511 | qword_1803B0B70 -> novels.api.novels_api#388 bytes [3932689, 3932716]: "novels_list_style_sources"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B70, v218);
            if ( *v218 >= 0 )
            {
              v20 = (*(_QWORD *)v218)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v218 + 1) + 48LL))(v218);
            }
// Original IDA line 1518 | qword_1803B0B80 -> novels.api.novels_api#390 bytes [3932753, 3932796]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932755, 3932767], "value": "profile_id"}, {"tag": "a", "fileRange": [3932767, 3932775], "value": "return"}], "values": [{"tag": "a", "fileRange": [3932775, 3932780], "value": "str"}, {"t... (truncated literal)
            v219 = sub_18036FB80(a1, qword_1803B0B80);
            v220 = (int *)sub_1803609A0(
// Original IDA line 1520 | sub_180077890 -> NovelsApi.novels_get_style_profile
                            (int)sub_180077890,
// Original IDA line 1521 | qword_1803B0B88 -> novels.api.novels_api#391 bytes [3932796, 3932822]: "novels_get_style_profile"
                            qword_1803B0B88,
// Original IDA line 1522 | qword_1803B0B90 -> novels.api.novels_api#392 bytes [3932822, 3932858]: "NovelsApi.novels_get_style_profile"
                            qword_1803B0B90,
                            qword_1803AFD08,
                            0,
                            0,
                            v219,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1531 | qword_1803B0B88 -> novels.api.novels_api#391 bytes [3932796, 3932822]: "novels_get_style_profile"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B88, v220);
            if ( *v220 >= 0 )
            {
              v20 = (*(_QWORD *)v220)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v220 + 1) + 48LL))(v220);
            }
// Original IDA line 1538 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v221 = sub_18036FB80(a1, qword_1803B0AF0);
            v222 = (int *)sub_1803609A0(
// Original IDA line 1540 | sub_1800780B0 -> NovelsApi.novels_save_style_profile
                            (int)sub_1800780B0,
// Original IDA line 1541 | qword_1803B0B98 -> novels.api.novels_api#393 bytes [3932858, 3932885]: "novels_save_style_profile"
                            qword_1803B0B98,
// Original IDA line 1542 | qword_1803B0BA0 -> novels.api.novels_api#394 bytes [3932885, 3932922]: "NovelsApi.novels_save_style_profile"
                            qword_1803B0BA0,
                            qword_1803AFDA8,
                            0,
                            0,
                            v221,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1551 | qword_1803B0B98 -> novels.api.novels_api#393 bytes [3932858, 3932885]: "novels_save_style_profile"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0B98, v222);
            if ( *v222 >= 0 )
            {
              v20 = (*(_QWORD *)v222)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v222 + 1) + 48LL))(v222);
            }
// Original IDA line 1558 | qword_1803B0B80 -> novels.api.novels_api#390 bytes [3932753, 3932796]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932755, 3932767], "value": "profile_id"}, {"tag": "a", "fileRange": [3932767, 3932775], "value": "return"}], "values": [{"tag": "a", "fileRange": [3932775, 3932780], "value": "str"}, {"t... (truncated literal)
            v223 = sub_18036FB80(a1, qword_1803B0B80);
            v224 = (int *)sub_1803609A0(
// Original IDA line 1560 | sub_180079700 -> NovelsApi.novels_delete_style_profile
                            (int)sub_180079700,
// Original IDA line 1561 | qword_1803B0BA8 -> novels.api.novels_api#395 bytes [3932922, 3932951]: "novels_delete_style_profile"
                            qword_1803B0BA8,
// Original IDA line 1562 | qword_1803B0BB0 -> novels.api.novels_api#396 bytes [3932951, 3932990]: "NovelsApi.novels_delete_style_profile"
                            qword_1803B0BB0,
                            qword_1803AFE80,
                            0,
                            0,
                            v223,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1571 | qword_1803B0BA8 -> novels.api.novels_api#395 bytes [3932922, 3932951]: "novels_delete_style_profile"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BA8, v224);
            if ( *v224 >= 0 )
            {
              v20 = (*(_QWORD *)v224)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v224 + 1) + 48LL))(v224);
            }
// Original IDA line 1578 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v225 = sub_18036FB80(a1, qword_1803B0AF0);
            v226 = (int *)sub_1803609A0(
// Original IDA line 1580 | sub_180079EF0 -> NovelsApi.novels_import_style_files
                            (int)sub_180079EF0,
// Original IDA line 1581 | qword_1803B0BB8 -> novels.api.novels_api#397 bytes [3932990, 3933017]: "novels_import_style_files"
                            qword_1803B0BB8,
// Original IDA line 1582 | qword_1803B0BC0 -> novels.api.novels_api#398 bytes [3933017, 3933054]: "NovelsApi.novels_import_style_files"
                            qword_1803B0BC0,
                            qword_1803AFD18,
                            0,
                            0,
                            v225,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1591 | qword_1803B0BB8 -> novels.api.novels_api#397 bytes [3932990, 3933017]: "novels_import_style_files"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BB8, v226);
            if ( *v226 >= 0 )
            {
              v20 = (*(_QWORD *)v226)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v226 + 1) + 48LL))(v226);
            }
// Original IDA line 1598 | qword_1803B0A20 -> novels.api.novels_api#346 bytes [3931216, 3931242]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931218, 3931226], "value": "return"}], "values": [{"tag": "u", "fileRange": [3931226, 3931242], "value": "Dict[str, Any]"}]}
            v227 = sub_18036FB80(a1, qword_1803B0A20);
            v228 = (int *)sub_1803609A0(
// Original IDA line 1600 | sub_18007B680 -> NovelsApi.novels_pick_style_files
                            (int)sub_18007B680,
// Original IDA line 1601 | qword_1803B0BC8 -> novels.api.novels_api#399 bytes [3933054, 3933079]: "novels_pick_style_files"
                            qword_1803B0BC8,
// Original IDA line 1602 | qword_1803B0BD0 -> novels.api.novels_api#400 bytes [3933079, 3933114]: "NovelsApi.novels_pick_style_files"
                            qword_1803B0BD0,
                            qword_1803AFBE8,
                            0,
                            0,
                            v227,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1611 | qword_1803B0BC8 -> novels.api.novels_api#399 bytes [3933054, 3933079]: "novels_pick_style_files"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BC8, v228);
            if ( *v228 >= 0 )
            {
              v20 = (*(_QWORD *)v228)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v228 + 1) + 48LL))(v228);
            }
// Original IDA line 1618 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v229 = sub_18036FB80(a1, qword_1803B0AF0);
            v230 = (int *)sub_1803609A0(
// Original IDA line 1620 | sub_18007BE90 -> NovelsApi.novels_init_imitation
                            (int)sub_18007BE90,
// Original IDA line 1621 | qword_1803B0BD8 -> novels.api.novels_api#401 bytes [3933114, 3933137]: "novels_init_imitation"
                            qword_1803B0BD8,
// Original IDA line 1622 | qword_1803B0BE0 -> novels.api.novels_api#402 bytes [3933137, 3933170]: "NovelsApi.novels_init_imitation"
                            qword_1803B0BE0,
                            qword_1803AFDC0,
                            0,
                            0,
                            v229,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1631 | qword_1803B0BD8 -> novels.api.novels_api#401 bytes [3933114, 3933137]: "novels_init_imitation"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BD8, v230);
            if ( *v230 >= 0 )
            {
              v20 = (*(_QWORD *)v230)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v230 + 1) + 48LL))(v230);
            }
// Original IDA line 1638 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v231 = sub_18036FB80(a1, qword_1803B0AF0);
            v232 = (int *)sub_1803609A0(
// Original IDA line 1640 | sub_18007C840 -> NovelsApi.novels_run_imitation_short
                            (int)sub_18007C840,
// Original IDA line 1641 | qword_1803B0BE8 -> novels.api.novels_api#403 bytes [3933170, 3933198]: "novels_run_imitation_short"
                            qword_1803B0BE8,
// Original IDA line 1642 | qword_1803B0BF0 -> novels.api.novels_api#404 bytes [3933198, 3933236]: "NovelsApi.novels_run_imitation_short"
                            qword_1803B0BF0,
                            qword_1803AFB58,
                            0,
                            0,
                            v231,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1651 | qword_1803B0BE8 -> novels.api.novels_api#403 bytes [3933170, 3933198]: "novels_run_imitation_short"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BE8, v232);
            if ( *v232 >= 0 )
            {
              v20 = (*(_QWORD *)v232)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v232 + 1) + 48LL))(v232);
            }
// Original IDA line 1658 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v233 = sub_18036FB80(a1, qword_1803B0AF0);
            v234 = (int *)sub_1803609A0(
// Original IDA line 1660 | sub_18007F270 -> NovelsApi.novels_init_fanfic
                            (int)sub_18007F270,
// Original IDA line 1661 | qword_1803B0BF8 -> novels.api.novels_api#405 bytes [3933236, 3933256]: "novels_init_fanfic"
                            qword_1803B0BF8,
// Original IDA line 1662 | qword_1803B0C00 -> novels.api.novels_api#406 bytes [3933256, 3933286]: "NovelsApi.novels_init_fanfic"
                            qword_1803B0C00,
                            qword_1803AFBF0,
                            0,
                            0,
                            v233,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1671 | qword_1803B0BF8 -> novels.api.novels_api#405 bytes [3933236, 3933256]: "novels_init_fanfic"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0BF8, v234);
            if ( *v234 >= 0 )
            {
              v20 = (*(_QWORD *)v234)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v234 + 1) + 48LL))(v234);
            }
// Original IDA line 1678 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v235 = sub_18036FB80(a1, qword_1803B0AF0);
            v236 = (int *)sub_1803609A0(
// Original IDA line 1680 | sub_18007FC20 -> NovelsApi.novels_init_spinoff
                            (int)sub_18007FC20,
// Original IDA line 1681 | qword_1803B0C08 -> novels.api.novels_api#407 bytes [3933286, 3933307]: "novels_init_spinoff"
                            qword_1803B0C08,
// Original IDA line 1682 | qword_1803B0C10 -> novels.api.novels_api#408 bytes [3933307, 3933338]: "NovelsApi.novels_init_spinoff"
                            qword_1803B0C10,
                            qword_1803B0FD0,
                            0,
                            0,
                            v235,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1691 | qword_1803B0C08 -> novels.api.novels_api#407 bytes [3933286, 3933307]: "novels_init_spinoff"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C08, v236);
            if ( *v236 >= 0 )
            {
              v20 = (*(_QWORD *)v236)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v236 + 1) + 48LL))(v236);
            }
// Original IDA line 1698 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v237 = sub_18036FB80(a1, qword_1803B0AF0);
            v238 = (int *)sub_1803609A0(
// Original IDA line 1700 | sub_1800805D0 -> NovelsApi.novels_import_chapters
                            (int)sub_1800805D0,
// Original IDA line 1701 | qword_1803B0C18 -> novels.api.novels_api#409 bytes [3933338, 3933362]: "novels_import_chapters"
                            qword_1803B0C18,
// Original IDA line 1702 | qword_1803B0C20 -> novels.api.novels_api#410 bytes [3933362, 3933396]: "NovelsApi.novels_import_chapters"
                            qword_1803B0C20,
                            qword_1803AFD28,
                            0,
                            0,
                            v237,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1711 | qword_1803B0C18 -> novels.api.novels_api#409 bytes [3933338, 3933362]: "novels_import_chapters"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C18, v238);
            if ( *v238 >= 0 )
            {
              v20 = (*(_QWORD *)v238)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v238 + 1) + 48LL))(v238);
            }
// Original IDA line 1718 | qword_1803B0A20 -> novels.api.novels_api#346 bytes [3931216, 3931242]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931218, 3931226], "value": "return"}], "values": [{"tag": "u", "fileRange": [3931226, 3931242], "value": "Dict[str, Any]"}]}
            v239 = sub_18036FB80(a1, qword_1803B0A20);
            v240 = (int *)sub_1803609A0(
// Original IDA line 1720 | sub_180081B90 -> NovelsApi.novels_get_agent_settings
                            (int)sub_180081B90,
// Original IDA line 1721 | qword_1803B0C28 -> novels.api.novels_api#411 bytes [3933396, 3933423]: "novels_get_agent_settings"
                            qword_1803B0C28,
// Original IDA line 1722 | qword_1803B0C30 -> novels.api.novels_api#412 bytes [3933423, 3933460]: "NovelsApi.novels_get_agent_settings"
                            qword_1803B0C30,
                            qword_1803AFBF8,
                            0,
                            0,
                            v239,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1731 | qword_1803B0C28 -> novels.api.novels_api#411 bytes [3933396, 3933423]: "novels_get_agent_settings"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C28, v240);
            if ( *v240 >= 0 )
            {
              v20 = (*(_QWORD *)v240)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v240 + 1) + 48LL))(v240);
            }
// Original IDA line 1738 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v241 = sub_18036FB80(a1, qword_1803B0AF0);
            v242 = (int *)sub_1803609A0(
// Original IDA line 1740 | sub_1800823B0 -> NovelsApi.novels_save_agent_settings
                            (int)sub_1800823B0,
// Original IDA line 1741 | qword_1803B0C38 -> novels.api.novels_api#413 bytes [3933460, 3933488]: "novels_save_agent_settings"
                            qword_1803B0C38,
// Original IDA line 1742 | qword_1803B0C40 -> novels.api.novels_api#414 bytes [3933488, 3933526]: "NovelsApi.novels_save_agent_settings"
                            qword_1803B0C40,
                            qword_1803AFF40,
                            0,
                            0,
                            v241,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1751 | qword_1803B0C38 -> novels.api.novels_api#413 bytes [3933460, 3933488]: "novels_save_agent_settings"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C38, v242);
            if ( *v242 >= 0 )
            {
              v20 = (*(_QWORD *)v242)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v242 + 1) + 48LL))(v242);
            }
// Original IDA line 1758 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v243 = sub_18036FB80(a1, qword_1803B0AF0);
            v244 = (int *)sub_1803609A0(
// Original IDA line 1760 | sub_180082CB0 -> NovelsApi.novels_send_message
                            (int)sub_180082CB0,
// Original IDA line 1761 | qword_1803B0C48 -> novels.api.novels_api#415 bytes [3933526, 3933547]: "novels_send_message"
                            qword_1803B0C48,
// Original IDA line 1762 | qword_1803B0C50 -> novels.api.novels_api#416 bytes [3933547, 3933578]: "NovelsApi.novels_send_message"
                            qword_1803B0C50,
                            qword_1803AFDF0,
                            0,
                            0,
                            v243,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1771 | qword_1803B0C48 -> novels.api.novels_api#415 bytes [3933526, 3933547]: "novels_send_message"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C48, v244);
            if ( *v244 >= 0 )
            {
              v20 = (*(_QWORD *)v244)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v244 + 1) + 48LL))(v244);
            }
// Original IDA line 1778 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v245 = sub_18036FB80(a1, qword_1803B0AF0);
            v246 = (int *)sub_1803609A0(
// Original IDA line 1780 | sub_180084690 -> NovelsApi.novels_confirm_action
                            (int)sub_180084690,
// Original IDA line 1781 | qword_1803B0C58 -> novels.api.novels_api#417 bytes [3933578, 3933601]: "novels_confirm_action"
                            qword_1803B0C58,
// Original IDA line 1782 | qword_1803B0C60 -> novels.api.novels_api#418 bytes [3933601, 3933634]: "NovelsApi.novels_confirm_action"
                            qword_1803B0C60,
                            qword_1803AFD30,
                            0,
                            0,
                            v245,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1791 | qword_1803B0C58 -> novels.api.novels_api#417 bytes [3933578, 3933601]: "novels_confirm_action"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C58, v246);
            if ( *v246 >= 0 )
            {
              v20 = (*(_QWORD *)v246)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v246 + 1) + 48LL))(v246);
            }
// Original IDA line 1798 | qword_1803B0C68 -> novels.api.novels_api#419 bytes [3933634, 3933670]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3933636, 3933641], "value": "key"}, {"tag": "a", "fileRange": [3933641, 3933649], "value": "return"}], "values": [{"tag": "a", "fileRange": [3933649, 3933654], "value": "str"}, {"tag": "u... (truncated literal)
            v247 = sub_18036FB80(a1, qword_1803B0C68);
            v248 = (int *)sub_1803609A0(
// Original IDA line 1800 | sub_180085CA0 -> NovelsApi.novels_cancel_operation
                            (int)sub_180085CA0,
// Original IDA line 1801 | qword_1803B0C70 -> novels.api.novels_api#420 bytes [3933670, 3933695]: "novels_cancel_operation"
                            qword_1803B0C70,
// Original IDA line 1802 | qword_1803B0C78 -> novels.api.novels_api#421 bytes [3933695, 3933730]: "NovelsApi.novels_cancel_operation"
                            qword_1803B0C78,
                            qword_1803AFE18,
                            0,
                            0,
                            v247,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1811 | qword_1803B0C70 -> novels.api.novels_api#420 bytes [3933670, 3933695]: "novels_cancel_operation"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C70, v248);
            if ( *v248 >= 0 )
            {
              v20 = (*(_QWORD *)v248)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v248 + 1) + 48LL))(v248);
            }
// Original IDA line 1818 | qword_1803B0C80 -> novels.api.novels_api#422 bytes [3933730, 3933791]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3933732, 3933741], "value": "book_id"}, {"tag": "a", "fileRange": [3933741, 3933757], "value": "chapter_number"}, {"tag": "a", "fileRange": [3933757, 3933765], "value": "return"}], "value... (truncated literal)
            v249 = sub_18036FB80(a1, qword_1803B0C80);
            v250 = (int *)sub_1803609A0(
// Original IDA line 1820 | sub_1800863B0 -> NovelsApi.novels_get_chapter
                            (int)sub_1800863B0,
// Original IDA line 1821 | qword_1803B0C88 -> novels.api.novels_api#423 bytes [3933791, 3933811]: "novels_get_chapter"
                            qword_1803B0C88,
// Original IDA line 1822 | qword_1803B0C90 -> novels.api.novels_api#424 bytes [3933811, 3933841]: "NovelsApi.novels_get_chapter"
                            qword_1803B0C90,
                            qword_1803AFEC0,
                            0,
                            0,
                            v249,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1831 | qword_1803B0C88 -> novels.api.novels_api#423 bytes [3933791, 3933811]: "novels_get_chapter"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0C88, v250);
            if ( *v250 >= 0 )
            {
              v20 = (*(_QWORD *)v250)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v250 + 1) + 48LL))(v250);
            }
// Original IDA line 1838 | qword_1803B09A0 -> novels.api.novels_api#330 bytes [3930601, 3930604]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [3930603, 3930604], "value": null}]}
            v251 = (_DWORD *)qword_1803B09A0;
// Original IDA line 1839 | qword_1803B0C98 -> novels.api.novels_api#425 bytes [3933841, 3933951]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3933843, 3933852], "value": "book_id"}, {"tag": "a", "fileRange": [3933852, 3933868], "value": "chapter_number"}, {"tag": "a", "fileRange": [3933868, 3933877], "value": "content"}, {"tag"... (truncated literal)
            v252 = sub_18036FB80(a1, qword_1803B0C98);
            if ( *v251 != -1 )
              ++*v251;
            v253 = (int *)sub_1803609A0(
// Original IDA line 1843 | sub_180086D40 -> NovelsApi.novels_save_chapter
                            (int)sub_180086D40,
// Original IDA line 1844 | qword_1803B0CA0 -> novels.api.novels_api#426 bytes [3933951, 3933972]: "novels_save_chapter"
                            qword_1803B0CA0,
// Original IDA line 1845 | qword_1803B0CA8 -> novels.api.novels_api#427 bytes [3933972, 3934003]: "NovelsApi.novels_save_chapter"
                            qword_1803B0CA8,
                            qword_1803AFD98,
                            (__int64)v251,
                            0,
                            v252,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1854 | qword_1803B0CA0 -> novels.api.novels_api#426 bytes [3933951, 3933972]: "novels_save_chapter"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CA0, v253);
            if ( *v253 >= 0 )
            {
              v20 = (*(_QWORD *)v253)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v253 + 1) + 48LL))(v253);
            }
// Original IDA line 1861 | qword_1803B0CB0 -> novels.api.novels_api#428 bytes [3934003, 3934049]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3934005, 3934014], "value": "book_id"}, {"tag": "a", "fileRange": [3934014, 3934019], "value": "key"}, {"tag": "a", "fileRange": [3934019, 3934027], "value": "return"}], "values": [{"tag"... (truncated literal)
            v254 = sub_18036FB80(a1, qword_1803B0CB0);
            v255 = (int *)sub_1803609A0(
// Original IDA line 1863 | sub_180087910 -> NovelsApi.novels_read_story_file
                            (int)sub_180087910,
// Original IDA line 1864 | qword_1803B0CB8 -> novels.api.novels_api#429 bytes [3934049, 3934073]: "novels_read_story_file"
                            qword_1803B0CB8,
// Original IDA line 1865 | qword_1803B0CC0 -> novels.api.novels_api#430 bytes [3934073, 3934107]: "NovelsApi.novels_read_story_file"
                            qword_1803B0CC0,
                            qword_1803AFCA0,
                            0,
                            0,
                            v254,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1874 | qword_1803B0CB8 -> novels.api.novels_api#429 bytes [3934049, 3934073]: "novels_read_story_file"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CB8, v255);
            if ( *v255 >= 0 )
            {
              v20 = (*(_QWORD *)v255)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v255 + 1) + 48LL))(v255);
            }
// Original IDA line 1881 | qword_1803B0CC8 -> novels.api.novels_api#431 bytes [3934107, 3934163]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3934109, 3934118], "value": "book_id"}, {"tag": "a", "fileRange": [3934118, 3934123], "value": "key"}, {"tag": "a", "fileRange": [3934123, 3934132], "value": "content"}, {"tag": "a", "fil... (truncated literal)
            v256 = sub_18036FB80(a1, qword_1803B0CC8);
            v257 = (int *)sub_1803609A0(
// Original IDA line 1883 | sub_1800880C0 -> NovelsApi.novels_save_story_file
                            (int)sub_1800880C0,
// Original IDA line 1884 | qword_1803B0CD0 -> novels.api.novels_api#432 bytes [3934163, 3934187]: "novels_save_story_file"
                            qword_1803B0CD0,
// Original IDA line 1885 | qword_1803B0CD8 -> novels.api.novels_api#433 bytes [3934187, 3934221]: "NovelsApi.novels_save_story_file"
                            qword_1803B0CD8,
                            qword_1803AFC90,
                            0,
                            0,
                            v256,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1894 | qword_1803B0CD0 -> novels.api.novels_api#432 bytes [3934163, 3934187]: "novels_save_story_file"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CD0, v257);
            if ( *v257 >= 0 )
            {
              v20 = (*(_QWORD *)v257)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v257 + 1) + 48LL))(v257);
            }
// Original IDA line 1901 | qword_1803B0A38 -> novels.api.novels_api#349 bytes [3931290, 3931330]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3931292, 3931301], "value": "book_id"}, {"tag": "a", "fileRange": [3931301, 3931309], "value": "return"}], "values": [{"tag": "a", "fileRange": [3931309, 3931314], "value": "str"}, {"tag"... (truncated literal)
            v258 = sub_18036FB80(a1, qword_1803B0A38);
            v259 = (int *)sub_1803609A0(
// Original IDA line 1903 | sub_1800888E0 -> NovelsApi.novels_get_book_rules
                            (int)sub_1800888E0,
// Original IDA line 1904 | qword_1803B0CE0 -> novels.api.novels_api#434 bytes [3934221, 3934244]: "novels_get_book_rules"
                            qword_1803B0CE0,
// Original IDA line 1905 | qword_1803B0CE8 -> novels.api.novels_api#435 bytes [3934244, 3934277]: "NovelsApi.novels_get_book_rules"
                            qword_1803B0CE8,
                            qword_1803AFBE0,
                            0,
                            0,
                            v258,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1914 | qword_1803B0CE0 -> novels.api.novels_api#434 bytes [3934221, 3934244]: "novels_get_book_rules"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CE0, v259);
            if ( *v259 >= 0 )
            {
              v20 = (*(_QWORD *)v259)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v259 + 1) + 48LL))(v259);
            }
// Original IDA line 1921 | qword_1803B0CF0 -> novels.api.novels_api#436 bytes [3934277, 3934342]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3934279, 3934288], "value": "book_id"}, {"tag": "a", "fileRange": [3934288, 3934297], "value": "payload"}, {"tag": "a", "fileRange": [3934297, 3934305], "value": "return"}], "values": [{"... (truncated literal)
            v260 = sub_18036FB80(a1, qword_1803B0CF0);
            v261 = (int *)sub_1803609A0(
// Original IDA line 1923 | sub_180089120 -> NovelsApi.novels_save_book_rules
                            (int)sub_180089120,
// Original IDA line 1924 | qword_1803B0CF8 -> novels.api.novels_api#437 bytes [3934342, 3934366]: "novels_save_book_rules"
                            qword_1803B0CF8,
// Original IDA line 1925 | qword_1803B0D00 -> novels.api.novels_api#438 bytes [3934366, 3934400]: "NovelsApi.novels_save_book_rules"
                            qword_1803B0D00,
                            qword_1803AFD80,
                            0,
                            0,
                            v260,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1934 | qword_1803B0CF8 -> novels.api.novels_api#437 bytes [3934342, 3934366]: "novels_save_book_rules"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0CF8, v261);
            if ( *v261 >= 0 )
            {
              v20 = (*(_QWORD *)v261)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v261 + 1) + 48LL))(v261);
            }
// Original IDA line 1941 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v262 = sub_18036FB80(a1, qword_1803B0AF0);
            v263 = (int *)sub_1803609A0(
// Original IDA line 1943 | sub_1800899F0 -> NovelsApi.novels_write_next
                            (int)sub_1800899F0,
// Original IDA line 1944 | qword_1803B0D08 -> novels.api.novels_api#439 bytes [3934400, 3934419]: "novels_write_next"
                            qword_1803B0D08,
// Original IDA line 1945 | qword_1803B0D10 -> novels.api.novels_api#440 bytes [3934419, 3934448]: "NovelsApi.novels_write_next"
                            qword_1803B0D10,
                            qword_1803AFCC0,
                            0,
                            0,
                            v262,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1954 | qword_1803B0D08 -> novels.api.novels_api#439 bytes [3934400, 3934419]: "novels_write_next"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D08, v263);
            if ( *v263 >= 0 )
            {
              v20 = (*(_QWORD *)v263)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v263 + 1) + 48LL))(v263);
            }
// Original IDA line 1961 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v264 = sub_18036FB80(a1, qword_1803B0AF0);
            v265 = (int *)sub_1803609A0(
// Original IDA line 1963 | sub_18008A790 -> NovelsApi.novels_write_batch
                            (int)sub_18008A790,
// Original IDA line 1964 | qword_1803B0D18 -> novels.api.novels_api#441 bytes [3934448, 3934468]: "novels_write_batch"
                            qword_1803B0D18,
// Original IDA line 1965 | qword_1803B0D20 -> novels.api.novels_api#442 bytes [3934468, 3934498]: "NovelsApi.novels_write_batch"
                            qword_1803B0D20,
                            qword_1803AFEB0,
                            0,
                            0,
                            v264,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1974 | qword_1803B0D18 -> novels.api.novels_api#441 bytes [3934448, 3934468]: "novels_write_batch"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D18, v265);
            if ( *v265 >= 0 )
            {
              v20 = (*(_QWORD *)v265)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v265 + 1) + 48LL))(v265);
            }
// Original IDA line 1981 | qword_1803B0D28 -> novels.api.novels_api#443 bytes [3934498, 3934524]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3934500, 3934508], "value": "return"}], "values": [{"tag": "a", "fileRange": [3934508, 3934524], "value": "NovelScheduler"}]}
            v266 = sub_18036FB80(a1, qword_1803B0D28);
            v267 = (int *)sub_1803609A0(
// Original IDA line 1983 | sub_18008B840 -> NovelsApi._get_scheduler
                            (int)sub_18008B840,
// Original IDA line 1984 | qword_1803B0770 -> novels.api.novels_api#260 bytes [3929363, 3929379]: "_get_scheduler"
                            qword_1803B0770,
// Original IDA line 1985 | qword_1803B0D30 -> novels.api.novels_api#444 bytes [3934524, 3934550]: "NovelsApi._get_scheduler"
                            qword_1803B0D30,
                            qword_1803AFD10,
                            0,
                            0,
                            v266,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 1994 | qword_1803B0770 -> novels.api.novels_api#260 bytes [3929363, 3929379]: "_get_scheduler"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0770, v267);
            if ( *v267 >= 0 )
            {
              v20 = (*(_QWORD *)v267)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v267 + 1) + 48LL))(v267);
            }
// Original IDA line 2001 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v268 = sub_18036FB80(a1, qword_1803B0AF0);
            v269 = (int *)sub_1803609A0(
// Original IDA line 2003 | sub_18008C8A0 -> NovelsApi.novels_scheduler_start
                            (int)sub_18008C8A0,
// Original IDA line 2004 | qword_1803B0D38 -> novels.api.novels_api#445 bytes [3934550, 3934574]: "novels_scheduler_start"
                            qword_1803B0D38,
// Original IDA line 2005 | qword_1803B0D40 -> novels.api.novels_api#446 bytes [3934574, 3934608]: "NovelsApi.novels_scheduler_start"
                            qword_1803B0D40,
                            qword_1803AFC28,
                            0,
                            0,
                            v268,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2014 | qword_1803B0D38 -> novels.api.novels_api#445 bytes [3934550, 3934574]: "novels_scheduler_start"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D38, v269);
            if ( *v269 >= 0 )
            {
              v20 = (*(_QWORD *)v269)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v269 + 1) + 48LL))(v269);
            }
// Original IDA line 2021 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v270 = sub_18036FB80(a1, qword_1803B0AF0);
            v271 = (int *)sub_1803609A0(
// Original IDA line 2023 | sub_18008DAD0 -> NovelsApi.novels_scheduler_stop
                            (int)sub_18008DAD0,
// Original IDA line 2024 | qword_1803B0D48 -> novels.api.novels_api#447 bytes [3934608, 3934631]: "novels_scheduler_stop"
                            qword_1803B0D48,
// Original IDA line 2025 | qword_1803B0D50 -> novels.api.novels_api#448 bytes [3934631, 3934664]: "NovelsApi.novels_scheduler_stop"
                            qword_1803B0D50,
                            qword_1803B0FA0,
                            0,
                            0,
                            v270,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2034 | qword_1803B0D48 -> novels.api.novels_api#447 bytes [3934608, 3934631]: "novels_scheduler_stop"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D48, v271);
            if ( *v271 >= 0 )
            {
              v20 = (*(_QWORD *)v271)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v271 + 1) + 48LL))(v271);
            }
// Original IDA line 2041 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v272 = sub_18036FB80(a1, qword_1803B0AF0);
            v273 = (int *)sub_1803609A0(
// Original IDA line 2043 | sub_18008E1B0 -> NovelsApi.novels_scheduler_pause
                            (int)sub_18008E1B0,
// Original IDA line 2044 | qword_1803B0D58 -> novels.api.novels_api#449 bytes [3934664, 3934688]: "novels_scheduler_pause"
                            qword_1803B0D58,
// Original IDA line 2045 | qword_1803B0D60 -> novels.api.novels_api#450 bytes [3934688, 3934722]: "NovelsApi.novels_scheduler_pause"
                            qword_1803B0D60,
                            qword_1803AFD78,
                            0,
                            0,
                            v272,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2054 | qword_1803B0D58 -> novels.api.novels_api#449 bytes [3934664, 3934688]: "novels_scheduler_pause"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D58, v273);
            if ( *v273 >= 0 )
            {
              v20 = (*(_QWORD *)v273)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v273 + 1) + 48LL))(v273);
            }
// Original IDA line 2061 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v274 = sub_18036FB80(a1, qword_1803B0AF0);
            v275 = (int *)sub_1803609A0(
// Original IDA line 2063 | sub_18008E9D0 -> NovelsApi.novels_scheduler_status
                            (int)sub_18008E9D0,
// Original IDA line 2064 | qword_1803B0D68 -> novels.api.novels_api#451 bytes [3934722, 3934747]: "novels_scheduler_status"
                            qword_1803B0D68,
// Original IDA line 2065 | qword_1803B0D70 -> novels.api.novels_api#452 bytes [3934747, 3934782]: "NovelsApi.novels_scheduler_status"
                            qword_1803B0D70,
                            qword_1803AFBA8,
                            0,
                            0,
                            v274,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2074 | qword_1803B0D68 -> novels.api.novels_api#451 bytes [3934722, 3934747]: "novels_scheduler_status"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D68, v275);
            if ( *v275 >= 0 )
            {
              v20 = (*(_QWORD *)v275)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v275 + 1) + 48LL))(v275);
            }
// Original IDA line 2081 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v276 = sub_18036FB80(a1, qword_1803B0AF0);
            v277 = (int *)sub_1803609A0(
// Original IDA line 2083 | sub_18008EF80 -> NovelsApi.novels_audit_chapter
                            (int)sub_18008EF80,
// Original IDA line 2084 | qword_1803B0D78 -> novels.api.novels_api#453 bytes [3934782, 3934804]: "novels_audit_chapter"
                            qword_1803B0D78,
// Original IDA line 2085 | qword_1803B0D80 -> novels.api.novels_api#454 bytes [3934804, 3934836]: "NovelsApi.novels_audit_chapter"
                            qword_1803B0D80,
                            qword_1803AFC98,
                            0,
                            0,
                            v276,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2094 | qword_1803B0D78 -> novels.api.novels_api#453 bytes [3934782, 3934804]: "novels_audit_chapter"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D78, v277);
            if ( *v277 >= 0 )
            {
              v20 = (*(_QWORD *)v277)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v277 + 1) + 48LL))(v277);
            }
// Original IDA line 2101 | qword_1803B0AF0 -> novels.api.novels_api#372 bytes [3932087, 3932138]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3932089, 3932098], "value": "payload"}, {"tag": "a", "fileRange": [3932098, 3932106], "value": "return"}], "values": [{"tag": "u", "fileRange": [3932106, 3932122], "value": "Dict[str, Any... (truncated literal)
            v278 = sub_18036FB80(a1, qword_1803B0AF0);
            v279 = (int *)sub_1803609A0(
// Original IDA line 2103 | sub_18008FB70 -> NovelsApi.novels_revise_chapter
                            (int)sub_18008FB70,
// Original IDA line 2104 | qword_1803B0D88 -> novels.api.novels_api#455 bytes [3934836, 3934859]: "novels_revise_chapter"
                            qword_1803B0D88,
// Original IDA line 2105 | qword_1803B0D90 -> novels.api.novels_api#456 bytes [3934859, 3934892]: "NovelsApi.novels_revise_chapter"
                            qword_1803B0D90,
                            qword_1803AFAE8,
                            0,
                            0,
                            v278,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2114 | qword_1803B0D88 -> novels.api.novels_api#455 bytes [3934836, 3934859]: "novels_revise_chapter"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0D88, v279);
            if ( *v279 >= 0 )
            {
              v20 = (*(_QWORD *)v279)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v279 + 1) + 48LL))(v279);
            }
// Original IDA line 2121 | qword_1803B0D98 -> novels.api.novels_api#457 bytes [3934892, 3934898]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3934894, 3934898], "value": "md"}]}
            v280 = (_DWORD *)qword_1803B0D98;
// Original IDA line 2122 | qword_1803B0DA0 -> novels.api.novels_api#458 bytes [3934898, 3934944]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3934900, 3934909], "value": "book_id"}, {"tag": "a", "fileRange": [3934909, 3934914], "value": "fmt"}, {"tag": "a", "fileRange": [3934914, 3934922], "value": "return"}], "values": [{"tag"... (truncated literal)
            v281 = sub_18036FB80(a1, qword_1803B0DA0);
            if ( *v280 != -1 )
              ++*v280;
            v282 = (int *)sub_1803609A0(
// Original IDA line 2126 | sub_180090E60 -> NovelsApi.novels_export_book
                            (int)sub_180090E60,
// Original IDA line 2127 | qword_1803B0DA8 -> novels.api.novels_api#459 bytes [3934944, 3934964]: "novels_export_book"
                            qword_1803B0DA8,
// Original IDA line 2128 | qword_1803B0DB0 -> novels.api.novels_api#460 bytes [3934964, 3934994]: "NovelsApi.novels_export_book"
                            qword_1803B0DB0,
                            qword_1803AFCF8,
                            (__int64)v280,
                            0,
                            v281,
                            qword_1803CA828,
                            0,
                            nullptr,
                            0);
// Original IDA line 2137 | qword_1803B0DA8 -> novels.api.novels_api#459 bytes [3934944, 3934964]: "novels_export_book"
            PyDict_SetItem(v156, *(_QWORD *)&qword_1803B0DA8, v282);
            if ( *v282 >= 0 )
            {
              v20 = (*(_QWORD *)v282)-- == 1;
              if ( v20 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v282 + 1) + 48LL))(v282);
            }
            v283 = a1[2];
// Original IDA line 2145 | qword_1803B0948 -> novels.api.novels_api#319 bytes [3930394, 3930405]: "NovelsApi"
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
// Original IDA line 2296 | qword_1803B0948 -> novels.api.novels_api#319 bytes [3930394, 3930405]: "NovelsApi"
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
