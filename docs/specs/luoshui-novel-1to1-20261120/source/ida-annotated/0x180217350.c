// Analyst annotations only; original export: source/ida/0x180217350.c
// qword comments derive module constant slot + original file range; not runtime values.
__int64 __fastcall sub_180217350(__int64 a1, __int64 a2, __int64 a3)
{
  __int64 v4; // rsi
  __int64 v6; // rbp
  int *v7; // rbx
  int **v8; // rax
  int *v9; // r8
  __int64 *v10; // rax
  __int64 v11; // rsi
  __int64 v12; // rbx
  __int64 Length; // rax
  __int64 v14; // rax
  __int64 v15; // rax
  __int64 v16; // rcx
  _QWORD *v17; // rax
  __int64 Dict; // rax
  __int64 v19; // rcx
  __int64 v20; // rax
  __int64 v21; // rbp
  int *v22; // rbx
  __int64 v23; // r14
  int **v24; // rax
  int *v25; // r8
  __int64 v26; // rax
  int *AttrString; // rbx
  __int64 v28; // rcx
  __int64 v29; // rbp
  __int64 v30; // rcx
  __int64 v31; // rcx
  __int64 v32; // rcx
  __int64 v33; // rdx
  _DWORD *v34; // r8
  __int64 v35; // r13
  __int64 v36; // rbx
  __int64 *v37; // rax
  __int64 v38; // rcx
  __int64 v39; // rdx
  __int64 *v40; // rax
  __int64 v41; // rdx
  __int64 v42; // rbx
  unsigned int v43; // ebp
  __int64 v44; // rax
  __int64 v45; // rcx
  __int64 v46; // rcx
  __int64 v47; // rax
  __int64 v48; // rcx
  __int64 v49; // rax
  __int64 v50; // rcx
  int v51; // edx
  int v52; // r8d
  __int64 v53; // rax
  __int64 v54; // rcx
  int v55; // edx
  int v56; // r8d
  __int64 v57; // rax
  __int64 v58; // rcx
  int v59; // edx
  int v60; // r8d
  __int64 v61; // rax
  __int64 v62; // rcx
  __int64 v63; // rax
  int *v64; // r14
  __int64 v65; // rax
  __int64 v66; // rcx
  __int64 v67; // rax
  __int64 v68; // rcx
  __int64 v69; // rax
  __int64 v70; // rcx
  __int64 v71; // rax
  __int64 v72; // rcx
  __int64 v73; // rax
  __int64 v74; // rcx
  __int64 v75; // rax
  __int64 v76; // rcx
  int v77; // edx
  int v78; // r8d
  __int64 v79; // rax
  __int64 v80; // rax
  __int64 v81; // rcx
  __int64 v82; // rax
  __int64 v83; // rcx
  int v84; // edx
  int v85; // r8d
  __int64 v86; // rax
  int *v87; // rbx
  __int64 v88; // rax
  __int64 v89; // rcx
  __int64 v90; // rbp
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
  int *v103; // rbx
  __int64 v104; // rax
  __int64 v105; // rcx
  __int64 v106; // rbp
  int v107; // edx
  int v108; // r8d
  __int64 v109; // rax
  int *v110; // rbx
  __int64 v111; // rax
  __int64 v112; // rcx
  __int64 v113; // rbp
  int v114; // edx
  int v115; // r8d
  __int64 v116; // rax
  __int64 v117; // rax
  __int64 v118; // rcx
  __int64 v119; // rax
  __int64 v120; // rcx
  __int64 v121; // rax
  __int64 v122; // rcx
  int v123; // edx
  int v124; // r8d
  __int64 v125; // rax
  __int64 v126; // rax
  __int64 v127; // rcx
  __int64 v128; // rax
  __int64 v129; // rcx
  int v130; // edx
  int v131; // r8d
  __int64 v132; // rax
  int *v133; // rbx
  __int64 v134; // rax
  __int64 v135; // rcx
  __int64 v136; // rbp
  int v137; // edx
  int v138; // r8d
  __int64 v139; // rax
  int *v140; // rbx
  __int64 v141; // rax
  __int64 v142; // rcx
  __int64 v143; // rbp
  int v144; // edx
  int v145; // r8d
  __int64 v146; // rax
  int *v147; // rbx
  __int64 v148; // rax
  __int64 v149; // rcx
  __int64 v150; // rbp
  int v151; // edx
  int v152; // r8d
  __int64 v153; // rax
  int *v154; // rbx
  __int64 v155; // rax
  __int64 v156; // rcx
  __int64 v157; // rbp
  __int64 v158; // rax
  __int64 v159; // rcx
  int *v160; // r14
  __int64 v161; // rbp
  __int64 v162; // rax
  __int64 v163; // rcx
  __int64 v164; // rax
  int *v165; // rbx
  __int64 v166; // rcx
  __int64 v167; // rbp
  int *v168; // r12
  __int64 v169; // r14
  __int64 *v170; // rax
  __int64 v171; // r8
  __int64 *v172; // rax
  _DWORD *v173; // rbx
  __int64 v174; // rax
  int *v175; // rbx
  __int64 v176; // rax
  int *v177; // rbx
  _DWORD *v178; // rbx
  __int64 v179; // rax
  int *v180; // rbx
  _DWORD *v181; // rbx
  __int64 v182; // rbp
  __int64 v183; // rax
  int *v184; // rbx
  _DWORD *v185; // rbx
  __int64 v186; // rax
  int *v187; // rbx
  _DWORD *v188; // rbx
  __int64 v189; // rax
  int *v190; // rbx
  __int64 v191; // rax
  int *v192; // r15
  int *v193; // rbx
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
  __int64 v206; // rcx
  _DWORD *v207; // r15
  int *v208; // rbx
  __int64 v209; // rcx
  __int64 v210; // rcx
  __int64 v211; // r15
  __int64 v212; // r15
  __int64 v213; // rax
  bool v214; // zf
  __int64 v215; // r14
  int *v216; // rcx
  int v217; // eax
  int *v218; // rcx
  __int64 result; // rax
  __int64 v220; // rdx
  __int64 v221; // [rsp+20h] [rbp-98h]
  __int64 v222; // [rsp+20h] [rbp-98h]
  __int64 v223; // [rsp+20h] [rbp-98h]
  __int64 v224; // [rsp+20h] [rbp-98h]
  __int64 v225; // [rsp+20h] [rbp-98h]
  __int64 v226; // [rsp+20h] [rbp-98h]
  __int64 v227; // [rsp+20h] [rbp-98h]
  __int64 v228; // [rsp+20h] [rbp-98h]
  __int64 v229; // [rsp+20h] [rbp-98h]
  __int64 v230; // [rsp+20h] [rbp-98h]
  __int64 v231; // [rsp+20h] [rbp-98h]
  __int64 v232; // [rsp+28h] [rbp-90h]
  __int64 v233; // [rsp+28h] [rbp-90h]
  __int64 v234; // [rsp+28h] [rbp-90h]
  __int64 v235; // [rsp+28h] [rbp-90h]
  __int64 v236; // [rsp+28h] [rbp-90h]
  __int64 v237; // [rsp+28h] [rbp-90h]
  __int64 v238; // [rsp+28h] [rbp-90h]
  __int64 v239; // [rsp+28h] [rbp-90h]
  __int64 v240; // [rsp+28h] [rbp-90h]
  __int64 v241; // [rsp+28h] [rbp-90h]
  __int64 v242; // [rsp+28h] [rbp-90h]
  __int64 v243; // [rsp+28h] [rbp-90h]
  __int64 v244; // [rsp+28h] [rbp-90h]
  __int64 v245; // [rsp+28h] [rbp-90h]
  __int64 v246; // [rsp+60h] [rbp-58h]
  _QWORD v247[3]; // [rsp+68h] [rbp-50h] BYREF
  __int64 v248; // [rsp+C8h] [rbp+10h] BYREF

  v4 = *(_QWORD *)(a2 + 16);
  qword_1803CA5B8 = v4;
  qword_1803CA5C0 = a2;
  if ( byte_1803B6C2E == 0 )
  {
    if ( byte_1803B6C2F == 0 )
    {
// Original IDA line 255 | qword_1803BDF10 -> novels.services.agent_session#0 bytes [4306930, 4306937]: "state"
      sub_180388A30(a1, &qword_1803BDF10, "novels.services.agent_session");
      byte_1803B6C2F = 1;
    }
    sub_1801F4390();
    v4 = qword_1803CA5B8;
    byte_1803B6C2E = 1;
  }
  v6 = qword_1803AB330;
  v7 = (int *)qword_1803AB560;
  v8 = (int **)sub_1800015A0(v4, qword_1803AB330);
  if ( v8 == nullptr )
  {
LABEL_13:
    PyDict_SetItem(v4, v6, v7);
    goto LABEL_14;
  }
  v9 = *v8;
  if ( v7 != *v8 )
  {
    if ( v9 != nullptr )
    {
      if ( *v7 != -1 )
        ++*v7;
      *v8 = v7;
      if ( *v9 >= 0 )
      {
        v214 = (*(_QWORD *)v9)-- == 1;
        if ( v214 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v9 + 1) + 48LL))(v9);
      }
      goto LABEL_14;
    }
    goto LABEL_13;
  }
LABEL_14:
  v10 = (__int64 *)sub_1800015A0(qword_1803CA5B8, qword_1803AB2A8);
  v11 = 0;
  v12 = 0;
  if ( v10 != nullptr )
    v12 = *v10;
  Length = PyUnicode_GetLength(v12);
  v14 = PyUnicode_Find(v12, qword_1803AB4A8, 0, Length, -1);
  if ( v14 != -1 )
  {
    v15 = PyUnicode_Substring(v12, 0, v14);
    sub_1801F3490(v16, qword_1803AB2B0, v15);
  }
  v17 = (_QWORD *)sub_1800015A0(qword_1803CA5B8, qword_1803AB300);
  if ( v17 == nullptr || *v17 == 0 )
  {
    Dict = PyModule_GetDict(qword_1803AB780);
    sub_1801F3400(v19, qword_1803AB300, Dict);
  }
  v20 = sub_18038D790(a3);
  v21 = qword_1803AB328;
  v22 = (int *)v20;
  v23 = qword_1803CA5B8;
  v24 = (int **)sub_1800015A0(qword_1803CA5B8, qword_1803AB328);
  if ( v24 == nullptr )
    goto LABEL_29;
  v25 = *v24;
  if ( v22 != *v24 )
  {
    if ( v25 != nullptr )
    {
      if ( *v22 != -1 )
        ++*v22;
      *v24 = v22;
      if ( *v25 >= 0 )
      {
        v214 = (*(_QWORD *)v25)-- == 1;
        if ( v214 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v25 + 1) + 48LL))(v25);
      }
    }
    else
    {
LABEL_29:
      PyDict_SetItem(v23, v21, v22);
    }
  }
  v26 = qword_1803C9FD0;
  if ( qword_1803C9FD0 == 0 )
  {
    v26 = PyImport_ImportModule("importlib._bootstrap");
    qword_1803C9FD0 = v26;
  }
  AttrString = (int *)PyObject_GetAttrString(v26, "_spec_from_module");
  v29 = sub_180376770(a1, AttrString, qword_1803CA5C0);
  if ( *AttrString >= 0 )
  {
    v214 = (*(_QWORD *)AttrString)-- == 1;
    if ( v214 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)AttrString + 1) + 48LL))(AttrString);
  }
  if ( v29 == 0 )
  {
    PyErr_PrintEx(0);
    abort();
  }
  sub_18037F3A0(v28, v29, qword_1803AB4D0, Py_TrueStruct);
  sub_1801F3490(v30, qword_1803AB4C8, v29);
  v248 = 0;
// Original IDA line 358 | qword_1803BE9E0 -> novels.services.agent_session#346 bytes [4312822, 4312831]: "__doc__"
  sub_1801F3400(v31, qword_1803BE9E0, Py_NoneStruct);
// Original IDA line 359 | qword_1803BE9E8 -> novels.services.agent_session#347 bytes [4312831, 4312841]: "__file__"
  sub_1801F3400(v32, qword_1803BE9E8, qword_1803BDF00);
  v33 = qword_1803CA5C0;
  v34 = *(_DWORD **)(qword_1803CA5C0 + 16);
  if ( *v34 != -1 )
    ++*v34;
  v246 = sub_18036CBB0(qword_1803BED38, v33, v34, 0);
  v35 = v246;
  sub_180001D00(a1, v246);
  v36 = qword_1803BDF00;
// Original IDA line 368 | qword_1803BED20 -> novels.services.agent_session#450 bytes [4317732, 4317742]: "__spec__"
  v37 = (__int64 *)sub_1800015A0(qword_1803CA5B8, qword_1803BED20);
  if ( v37 == nullptr || (v39 = *v37, *v37 == 0) )
  {
// Original IDA line 371 | qword_1803BED20 -> novels.services.agent_session#450 bytes [4317732, 4317742]: "__spec__"
    v40 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803BED20);
    v39 = 0;
    if ( v40 != nullptr )
      v39 = *v40;
  }
// Original IDA line 376 | qword_1803BE9F0 -> novels.services.agent_session#348 bytes [4312841, 4312849]: "origin"
  if ( (unsigned __int8)sub_18037F3A0(v38, v39, qword_1803BE9F0, v36) == 0
// Original IDA line 377 | qword_1803BE9F8 -> novels.services.agent_session#349 bytes [4312849, 4312863]: "has_location"
    || (v44 = sub_1801F3740(), (unsigned __int8)sub_18037F3A0(v45, v44, qword_1803BE9F8, Py_TrueStruct) == 0) )
  {
    v42 = *(_QWORD *)(a1 + 96);
    v43 = 1;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_345;
  }
// Original IDA line 384 | qword_1803BEA00 -> novels.services.agent_session#350 bytes [4312863, 4312875]: "__cached__"
  sub_1801F3400(v46, qword_1803BEA00, Py_NoneStruct);
  v47 = sub_180372070();
// Original IDA line 386 | qword_1803BEA08 -> novels.services.agent_session#351 bytes [4312875, 4312888]: "annotations"
  v49 = sub_18037F0C0(v48, v47, qword_1803BEA08);
// Original IDA line 387 | qword_1803BEA08 -> novels.services.agent_session#351 bytes [4312875, 4312888]: "annotations"
  sub_1801F3490(v50, qword_1803BEA08, v49);
// Original IDA line 388 | qword_1803BE850 -> novels.services.agent_session#296 bytes [4311941, 4311947]: "json"
  v51 = qword_1803BE850;
  v52 = qword_1803CA5B8;
  v232 = qword_1803AB250;
  *(_DWORD *)(v246 + 40) = 3;
  v53 = sub_1803717E0(a1, v51, v52, Py_NoneStruct, Py_NoneStruct, v232);
  if ( v53 == 0 )
  {
    v42 = *(_QWORD *)(a1 + 96);
    v43 = 3;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_345;
  }
// Original IDA line 400 | qword_1803BE850 -> novels.services.agent_session#296 bytes [4311941, 4311947]: "json"
  sub_1801F3490(v54, qword_1803BE850, v53);
// Original IDA line 401 | qword_1803BEA10 -> novels.services.agent_session#352 bytes [4312888, 4312892]: "re"
  v55 = qword_1803BEA10;
  v56 = qword_1803CA5B8;
  v233 = qword_1803AB250;
  *(_DWORD *)(v246 + 40) = 4;
  v57 = sub_1803717E0(a1, v55, v56, Py_NoneStruct, Py_NoneStruct, v233);
  if ( v57 == 0 )
  {
    v42 = *(_QWORD *)(a1 + 96);
    v43 = 4;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_345;
  }
// Original IDA line 413 | qword_1803BEA10 -> novels.services.agent_session#352 bytes [4312888, 4312892]: "re"
  sub_1801F3490(v58, qword_1803BEA10, v57);
// Original IDA line 414 | qword_1803BE0F0 -> novels.services.agent_session#60 bytes [4307715, 4307721]: "time"
  v59 = qword_1803BE0F0;
  v60 = qword_1803CA5B8;
  v234 = qword_1803AB250;
  *(_DWORD *)(v246 + 40) = 5;
  v61 = sub_1803717E0(a1, v59, v60, Py_NoneStruct, Py_NoneStruct, v234);
  if ( v61 == 0 )
  {
    v42 = *(_QWORD *)(a1 + 96);
    v43 = 5;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_345;
  }
// Original IDA line 426 | qword_1803BE0F0 -> novels.services.agent_session#60 bytes [4307715, 4307721]: "time"
  sub_1801F3490(v62, qword_1803BE0F0, v61);
  v63 = sub_180372110();
  v64 = (int *)v63;
  if ( *(_DWORD *)v63 != -1 )
    ++*(_DWORD *)v63;
  if ( *(_QWORD *)(v63 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 432 | qword_1803BEA18 -> novels.services.agent_session#353 bytes [4312892, 4312897]: "Any"
    v65 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA18, qword_1803AB250);
  else
// Original IDA line 434 | qword_1803BEA18 -> novels.services.agent_session#353 bytes [4312892, 4312897]: "Any"
    v65 = sub_180371890(a1, v64, qword_1803BEA18);
  if ( v65 != 0 )
  {
// Original IDA line 437 | qword_1803BEA18 -> novels.services.agent_session#353 bytes [4312892, 4312897]: "Any"
    sub_1801F3490(v66, qword_1803BEA18, v65);
    if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 439 | qword_1803BEA20 -> novels.services.agent_session#354 bytes [4312897, 4312907]: "Callable"
      v67 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA20, qword_1803AB250);
    else
// Original IDA line 441 | qword_1803BEA20 -> novels.services.agent_session#354 bytes [4312897, 4312907]: "Callable"
      v67 = sub_180371890(a1, v64, qword_1803BEA20);
    if ( v67 != 0 )
    {
// Original IDA line 444 | qword_1803BEA20 -> novels.services.agent_session#354 bytes [4312897, 4312907]: "Callable"
      sub_1801F3490(v68, qword_1803BEA20, v67);
      if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 446 | qword_1803BEA28 -> novels.services.agent_session#355 bytes [4312907, 4312913]: "Dict"
        v69 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA28, qword_1803AB250);
      else
// Original IDA line 448 | qword_1803BEA28 -> novels.services.agent_session#355 bytes [4312907, 4312913]: "Dict"
        v69 = sub_180371890(a1, v64, qword_1803BEA28);
      if ( v69 != 0 )
      {
// Original IDA line 451 | qword_1803BEA28 -> novels.services.agent_session#355 bytes [4312907, 4312913]: "Dict"
        sub_1801F3490(v70, qword_1803BEA28, v69);
        if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 453 | qword_1803BEA30 -> novels.services.agent_session#356 bytes [4312913, 4312919]: "List"
          v71 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA30, qword_1803AB250);
        else
// Original IDA line 455 | qword_1803BEA30 -> novels.services.agent_session#356 bytes [4312913, 4312919]: "List"
          v71 = sub_180371890(a1, v64, qword_1803BEA30);
        if ( v71 != 0 )
        {
// Original IDA line 458 | qword_1803BEA30 -> novels.services.agent_session#356 bytes [4312913, 4312919]: "List"
          sub_1801F3490(v72, qword_1803BEA30, v71);
          if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 460 | qword_1803BEA38 -> novels.services.agent_session#357 bytes [4312919, 4312929]: "Optional"
            v73 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA38, qword_1803AB250);
          else
// Original IDA line 462 | qword_1803BEA38 -> novels.services.agent_session#357 bytes [4312919, 4312929]: "Optional"
            v73 = sub_180371890(a1, v64, qword_1803BEA38);
          if ( v73 != 0 )
          {
// Original IDA line 465 | qword_1803BEA38 -> novels.services.agent_session#357 bytes [4312919, 4312929]: "Optional"
            sub_1801F3490(v74, qword_1803BEA38, v73);
            if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 467 | qword_1803BEA40 -> novels.services.agent_session#358 bytes [4312929, 4312936]: "Tuple"
              v75 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA40, qword_1803AB250);
            else
// Original IDA line 469 | qword_1803BEA40 -> novels.services.agent_session#358 bytes [4312929, 4312936]: "Tuple"
              v75 = sub_180371890(a1, v64, qword_1803BEA40);
            if ( v75 != 0 )
            {
// Original IDA line 472 | qword_1803BEA40 -> novels.services.agent_session#358 bytes [4312929, 4312936]: "Tuple"
              sub_1801F3490(v76, qword_1803BEA40, v75);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
// Original IDA line 479 | qword_1803BEA48 -> novels.services.agent_session#359 bytes [4312936, 4312952]: "models.schemas"
              v77 = qword_1803BEA48;
              v78 = qword_1803CA5B8;
// Original IDA line 481 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v235 = qword_1803BE028;
// Original IDA line 482 | qword_1803BEA50 -> novels.services.agent_session#360 bytes [4312952, 4312982]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4312954, 4312969], "value": "SessionRecord"}, {"tag": "a", "fileRange": [4312969, 4312982], "value": "utc_now_iso"}]}
              v221 = qword_1803BEA50;
              *(_DWORD *)(v246 + 40) = 8;
              v79 = sub_1803717E0(a1, v77, v78, Py_NoneStruct, v221, v235);
              v64 = (int *)v79;
              if ( v79 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 8;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v79 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 494 | qword_1803BEA58 -> novels.services.agent_session#361 bytes [4312982, 4312997]: "SessionRecord"
                v80 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA58, qword_1803AB250);
              else
// Original IDA line 496 | qword_1803BEA58 -> novels.services.agent_session#361 bytes [4312982, 4312997]: "SessionRecord"
                v80 = sub_180371890(a1, v64, qword_1803BEA58);
              if ( v80 == 0
// Original IDA line 498 | qword_1803BEA58 -> novels.services.agent_session#361 bytes [4312982, 4312997]: "SessionRecord"
                || ((sub_1801F3490(v81, qword_1803BEA58, v80), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 500 | qword_1803BE0C8 -> novels.services.agent_session#55 bytes [4307671, 4307684]: "utc_now_iso"
                  ? (v82 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE0C8, qword_1803AB250))
// Original IDA line 501 | qword_1803BE0C8 -> novels.services.agent_session#55 bytes [4307671, 4307684]: "utc_now_iso"
                  : (v82 = sub_180371890(a1, v64, qword_1803BE0C8)),
                    v82 == 0) )
              {
                v43 = 8;
                goto LABEL_85;
              }
// Original IDA line 507 | qword_1803BE0C8 -> novels.services.agent_session#55 bytes [4307671, 4307684]: "utc_now_iso"
              sub_1801F3490(v83, qword_1803BE0C8, v82);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
// Original IDA line 514 | qword_1803BEA60 -> novels.services.agent_session#362 bytes [4312997, 4313023]: "pipeline.pipeline_runner"
              v84 = qword_1803BEA60;
              v85 = qword_1803CA5B8;
// Original IDA line 516 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v236 = qword_1803BE028;
// Original IDA line 517 | qword_1803BEA68 -> novels.services.agent_session#363 bytes [4313023, 4313041]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313025, 4313041], "value": "PipelineRunner"}]}
              v222 = qword_1803BEA68;
              *(_DWORD *)(v246 + 40) = 9;
              v86 = sub_1803717E0(a1, v84, v85, Py_NoneStruct, v222, v236);
              v87 = (int *)v86;
              if ( v86 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 9;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v86 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 529 | qword_1803BEA70 -> novels.services.agent_session#364 bytes [4313041, 4313057]: "PipelineRunner"
                v88 = sub_180371D80(a1, (_DWORD)v87, qword_1803CA5B8, qword_1803BEA70, qword_1803AB250);
              else
// Original IDA line 531 | qword_1803BEA70 -> novels.services.agent_session#364 bytes [4313041, 4313057]: "PipelineRunner"
                v88 = sub_180371890(a1, v87, qword_1803BEA70);
              v90 = v88;
              if ( *v87 >= 0 )
              {
                v214 = (*(_QWORD *)v87)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v87 + 1) + 48LL))(v87);
              }
              if ( v90 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 9;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 546 | qword_1803BEA70 -> novels.services.agent_session#364 bytes [4313041, 4313057]: "PipelineRunner"
              sub_1801F3490(v89, qword_1803BEA70, v90);
// Original IDA line 547 | qword_1803BEA78 -> novels.services.agent_session#365 bytes [4313057, 4313078]: "prompts.agent_tools"
              v91 = qword_1803BEA78;
              v92 = qword_1803CA5B8;
// Original IDA line 549 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v237 = qword_1803BE028;
// Original IDA line 550 | qword_1803BEA80 -> novels.services.agent_session#366 bytes [4313078, 4313160]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313080, 4313101], "value": "PROPOSE_ACTION_TOOL"}, {"tag": "a", "fileRange": [4313101, 4313130], "value": "play_tools_for_active_world"}, {"tag": "a", "fileRange": [4313130, 4313160], ... (truncated literal)
              v223 = qword_1803BEA80;
              *(_DWORD *)(v246 + 40) = 10;
              v93 = sub_1803717E0(a1, v91, v92, Py_NoneStruct, v223, v237);
              v64 = (int *)v93;
              if ( v93 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 10;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v93 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 562 | qword_1803BEA88 -> novels.services.agent_session#367 bytes [4313160, 4313181]: "PROPOSE_ACTION_TOOL"
                v94 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA88, qword_1803AB250);
              else
// Original IDA line 564 | qword_1803BEA88 -> novels.services.agent_session#367 bytes [4313160, 4313181]: "PROPOSE_ACTION_TOOL"
                v94 = sub_180371890(a1, v64, qword_1803BEA88);
              if ( v94 == 0
// Original IDA line 566 | qword_1803BEA88 -> novels.services.agent_session#367 bytes [4313160, 4313181]: "PROPOSE_ACTION_TOOL"
                || ((sub_1801F3490(v95, qword_1803BEA88, v94), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 568 | qword_1803BE830 -> novels.services.agent_session#292 bytes [4311798, 4311827]: "play_tools_for_active_world"
                  ? (v96 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE830, qword_1803AB250))
// Original IDA line 569 | qword_1803BE830 -> novels.services.agent_session#292 bytes [4311798, 4311827]: "play_tools_for_active_world"
                  : (v96 = sub_180371890(a1, v64, qword_1803BE830)),
                    v96 == 0
// Original IDA line 571 | qword_1803BE830 -> novels.services.agent_session#292 bytes [4311798, 4311827]: "play_tools_for_active_world"
                 || ((sub_1801F3490(v97, qword_1803BE830, v96), *((_QWORD *)v64 + 1) == PyModule_Type)
                  || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 573 | qword_1803BE138 -> novels.services.agent_session#69 bytes [4307920, 4307950]: "propose_action_tool_for_kind"
                   ? (v98 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE138, qword_1803AB250))
// Original IDA line 574 | qword_1803BE138 -> novels.services.agent_session#69 bytes [4307920, 4307950]: "propose_action_tool_for_kind"
                   : (v98 = sub_180371890(a1, v64, qword_1803BE138)),
                     v98 == 0)) )
              {
                v43 = 10;
                goto LABEL_85;
              }
// Original IDA line 580 | qword_1803BE138 -> novels.services.agent_session#69 bytes [4307920, 4307950]: "propose_action_tool_for_kind"
              sub_1801F3490(v99, qword_1803BE138, v98);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
// Original IDA line 587 | qword_1803BEA90 -> novels.services.agent_session#368 bytes [4313181, 4313205]: "prompts.system_prompts"
              v100 = qword_1803BEA90;
              v101 = qword_1803CA5B8;
// Original IDA line 589 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v238 = qword_1803BE028;
// Original IDA line 590 | qword_1803BEA98 -> novels.services.agent_session#369 bytes [4313205, 4313228]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313207, 4313228], "value": "build_system_prompt"}]}
              v224 = qword_1803BEA98;
              *(_DWORD *)(v246 + 40) = 15;
              v102 = sub_1803717E0(a1, v100, v101, Py_NoneStruct, v224, v238);
              v103 = (int *)v102;
              if ( v102 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 15;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v102 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 602 | qword_1803BE008 -> novels.services.agent_session#31 bytes [4307331, 4307352]: "build_system_prompt"
                v104 = sub_180371D80(a1, (_DWORD)v103, qword_1803CA5B8, qword_1803BE008, qword_1803AB250);
              else
// Original IDA line 604 | qword_1803BE008 -> novels.services.agent_session#31 bytes [4307331, 4307352]: "build_system_prompt"
                v104 = sub_180371890(a1, v103, qword_1803BE008);
              v106 = v104;
              if ( *v103 >= 0 )
              {
                v214 = (*(_QWORD *)v103)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v103 + 1) + 48LL))(v103);
              }
              if ( v106 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 15;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 619 | qword_1803BE008 -> novels.services.agent_session#31 bytes [4307331, 4307352]: "build_system_prompt"
              sub_1801F3490(v105, qword_1803BE008, v106);
// Original IDA line 620 | qword_1803BEAA0 -> novels.services.agent_session#370 bytes [4313228, 4313253]: "services.agent_settings"
              v107 = qword_1803BEAA0;
              v108 = qword_1803CA5B8;
// Original IDA line 622 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v239 = qword_1803BE028;
// Original IDA line 623 | qword_1803BEAA8 -> novels.services.agent_session#371 bytes [4313253, 4313275]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313255, 4313275], "value": "AgentSettingsStore"}]}
              v225 = qword_1803BEAA8;
              *(_DWORD *)(v246 + 40) = 16;
              v109 = sub_1803717E0(a1, v107, v108, Py_NoneStruct, v225, v239);
              v110 = (int *)v109;
              if ( v109 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 16;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v109 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 635 | qword_1803BDF60 -> novels.services.agent_session#10 bytes [4307066, 4307086]: "AgentSettingsStore"
                v111 = sub_180371D80(a1, (_DWORD)v110, qword_1803CA5B8, qword_1803BDF60, qword_1803AB250);
              else
// Original IDA line 637 | qword_1803BDF60 -> novels.services.agent_session#10 bytes [4307066, 4307086]: "AgentSettingsStore"
                v111 = sub_180371890(a1, v110, qword_1803BDF60);
              v113 = v111;
              if ( *v110 >= 0 )
              {
                v214 = (*(_QWORD *)v110)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v110 + 1) + 48LL))(v110);
              }
              if ( v113 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 16;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 652 | qword_1803BDF60 -> novels.services.agent_session#10 bytes [4307066, 4307086]: "AgentSettingsStore"
              sub_1801F3490(v112, qword_1803BDF60, v113);
// Original IDA line 653 | qword_1803BEAB0 -> novels.services.agent_session#372 bytes [4313275, 4313300]: "services.chapter_status"
              v114 = qword_1803BEAB0;
              v115 = qword_1803CA5B8;
// Original IDA line 655 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v240 = qword_1803BE028;
// Original IDA line 656 | qword_1803BEAB8 -> novels.services.agent_session#373 bytes [4313300, 4313374]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313302, 4313323], "value": "build_audit_summary"}, {"tag": "a", "fileRange": [4313323, 4313352], "value": "build_chapter_write_summary"}, {"tag": "a", "fileRange": [4313352, 4313374], ... (truncated literal)
              v226 = qword_1803BEAB8;
              *(_DWORD *)(v246 + 40) = 17;
              v116 = sub_1803717E0(a1, v114, v115, Py_NoneStruct, v226, v240);
              v64 = (int *)v116;
              if ( v116 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 17;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v116 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 668 | qword_1803BE5A0 -> novels.services.agent_session#210 bytes [4310268, 4310289]: "build_audit_summary"
                v117 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE5A0, qword_1803AB250);
              else
// Original IDA line 670 | qword_1803BE5A0 -> novels.services.agent_session#210 bytes [4310268, 4310289]: "build_audit_summary"
                v117 = sub_180371890(a1, v64, qword_1803BE5A0);
              if ( v117 == 0
// Original IDA line 672 | qword_1803BE5A0 -> novels.services.agent_session#210 bytes [4310268, 4310289]: "build_audit_summary"
                || ((sub_1801F3490(v118, qword_1803BE5A0, v117), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 674 | qword_1803BE538 -> novels.services.agent_session#197 bytes [4310064, 4310093]: "build_chapter_write_summary"
                  ? (v119 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE538, qword_1803AB250))
// Original IDA line 675 | qword_1803BE538 -> novels.services.agent_session#197 bytes [4310064, 4310093]: "build_chapter_write_summary"
                  : (v119 = sub_180371890(a1, v64, qword_1803BE538)),
                    v119 == 0
// Original IDA line 677 | qword_1803BE538 -> novels.services.agent_session#197 bytes [4310064, 4310093]: "build_chapter_write_summary"
                 || ((sub_1801F3490(v120, qword_1803BE538, v119), *((_QWORD *)v64 + 1) == PyModule_Type)
                  || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 679 | qword_1803BE548 -> novels.services.agent_session#199 bytes [4310101, 4310123]: "chapter_needs_review"
                   ? (v121 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE548, qword_1803AB250))
// Original IDA line 680 | qword_1803BE548 -> novels.services.agent_session#199 bytes [4310101, 4310123]: "chapter_needs_review"
                   : (v121 = sub_180371890(a1, v64, qword_1803BE548)),
                     v121 == 0)) )
              {
                v43 = 17;
                goto LABEL_85;
              }
// Original IDA line 686 | qword_1803BE548 -> novels.services.agent_session#199 bytes [4310101, 4310123]: "chapter_needs_review"
              sub_1801F3490(v122, qword_1803BE548, v121);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
// Original IDA line 693 | qword_1803BEAC0 -> novels.services.agent_session#374 bytes [4313374, 4313396]: "services.tool_policy"
              v123 = qword_1803BEAC0;
              v124 = qword_1803CA5B8;
// Original IDA line 695 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v241 = qword_1803BE028;
// Original IDA line 696 | qword_1803BEAC8 -> novels.services.agent_session#375 bytes [4313396, 4313442]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313398, 4313416], "value": "ASSISTED_ACTIONS"}, {"tag": "a", "fileRange": [4313416, 4313442], "value": "validate_proposed_action"}]}
              v227 = qword_1803BEAC8;
              *(_DWORD *)(v246 + 40) = 18;
              v125 = sub_1803717E0(a1, v123, v124, Py_NoneStruct, v227, v241);
              v64 = (int *)v125;
              if ( v125 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 18;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v125 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 708 | qword_1803BE1D8 -> novels.services.agent_session#89 bytes [4308243, 4308261]: "ASSISTED_ACTIONS"
                v126 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE1D8, qword_1803AB250);
              else
// Original IDA line 710 | qword_1803BE1D8 -> novels.services.agent_session#89 bytes [4308243, 4308261]: "ASSISTED_ACTIONS"
                v126 = sub_180371890(a1, v64, qword_1803BE1D8);
              if ( v126 == 0
// Original IDA line 712 | qword_1803BE1D8 -> novels.services.agent_session#89 bytes [4308243, 4308261]: "ASSISTED_ACTIONS"
                || ((sub_1801F3490(v127, qword_1803BE1D8, v126), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 714 | qword_1803BE1B0 -> novels.services.agent_session#84 bytes [4308170, 4308196]: "validate_proposed_action"
                  ? (v128 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE1B0, qword_1803AB250))
// Original IDA line 715 | qword_1803BE1B0 -> novels.services.agent_session#84 bytes [4308170, 4308196]: "validate_proposed_action"
                  : (v128 = sub_180371890(a1, v64, qword_1803BE1B0)),
                    v128 == 0) )
              {
                v43 = 18;
                goto LABEL_85;
              }
// Original IDA line 721 | qword_1803BE1B0 -> novels.services.agent_session#84 bytes [4308170, 4308196]: "validate_proposed_action"
              sub_1801F3490(v129, qword_1803BE1B0, v128);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
// Original IDA line 728 | qword_1803BEAD0 -> novels.services.agent_session#376 bytes [4313442, 4313465]: "storage.state_manager"
              v130 = qword_1803BEAD0;
              v131 = qword_1803CA5B8;
// Original IDA line 730 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v242 = qword_1803BE028;
// Original IDA line 731 | qword_1803BEAD8 -> novels.services.agent_session#377 bytes [4313465, 4313487]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313467, 4313487], "value": "NovelsStateManager"}]}
              v228 = qword_1803BEAD8;
              *(_DWORD *)(v246 + 40) = 22;
              v132 = sub_1803717E0(a1, v130, v131, Py_NoneStruct, v228, v242);
              v133 = (int *)v132;
              if ( v132 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 22;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v132 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 743 | qword_1803BEAE0 -> novels.services.agent_session#378 bytes [4313487, 4313507]: "NovelsStateManager"
                v134 = sub_180371D80(a1, (_DWORD)v133, qword_1803CA5B8, qword_1803BEAE0, qword_1803AB250);
              else
// Original IDA line 745 | qword_1803BEAE0 -> novels.services.agent_session#378 bytes [4313487, 4313507]: "NovelsStateManager"
                v134 = sub_180371890(a1, v133, qword_1803BEAE0);
              v136 = v134;
              if ( *v133 >= 0 )
              {
                v214 = (*(_QWORD *)v133)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v133 + 1) + 48LL))(v133);
              }
              if ( v136 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 22;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 760 | qword_1803BEAE0 -> novels.services.agent_session#378 bytes [4313487, 4313507]: "NovelsStateManager"
              sub_1801F3490(v135, qword_1803BEAE0, v136);
// Original IDA line 761 | qword_1803BEAE8 -> novels.services.agent_session#379 bytes [4313507, 4313524]: "utils.ai_client"
              v137 = qword_1803BEAE8;
              v138 = qword_1803CA5B8;
// Original IDA line 763 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v243 = qword_1803BE028;
// Original IDA line 764 | qword_1803BEAF0 -> novels.services.agent_session#380 bytes [4313524, 4313542]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313526, 4313542], "value": "NovelsAIClient"}]}
              v229 = qword_1803BEAF0;
              *(_DWORD *)(v246 + 40) = 23;
              v139 = sub_1803717E0(a1, v137, v138, Py_NoneStruct, v229, v243);
              v140 = (int *)v139;
              if ( v139 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 23;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v139 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 776 | qword_1803BDF28 -> novels.services.agent_session#3 bytes [4306966, 4306982]: "NovelsAIClient"
                v141 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA5B8, qword_1803BDF28, qword_1803AB250);
              else
// Original IDA line 778 | qword_1803BDF28 -> novels.services.agent_session#3 bytes [4306966, 4306982]: "NovelsAIClient"
                v141 = sub_180371890(a1, v140, qword_1803BDF28);
              v143 = v141;
              if ( *v140 >= 0 )
              {
                v214 = (*(_QWORD *)v140)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v140 + 1) + 48LL))(v140);
              }
              if ( v143 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 23;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 793 | qword_1803BDF28 -> novels.services.agent_session#3 bytes [4306966, 4306982]: "NovelsAIClient"
              sub_1801F3490(v142, qword_1803BDF28, v143);
// Original IDA line 794 | qword_1803BEAF8 -> novels.services.agent_session#381 bytes [4313542, 4313562]: "utils.cancellation"
              v144 = qword_1803BEAF8;
              v145 = qword_1803CA5B8;
// Original IDA line 796 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v244 = qword_1803BE028;
// Original IDA line 797 | qword_1803BEB00 -> novels.services.agent_session#382 bytes [4313562, 4313589]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313564, 4313589], "value": "OperationCancelledError"}]}
              v230 = qword_1803BEB00;
              *(_DWORD *)(v246 + 40) = 24;
              v146 = sub_1803717E0(a1, v144, v145, Py_NoneStruct, v230, v244);
              v147 = (int *)v146;
              if ( v146 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 24;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v146 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 809 | qword_1803BEB08 -> novels.services.agent_session#383 bytes [4313589, 4313614]: "OperationCancelledError"
                v148 = sub_180371D80(a1, (_DWORD)v147, qword_1803CA5B8, qword_1803BEB08, qword_1803AB250);
              else
// Original IDA line 811 | qword_1803BEB08 -> novels.services.agent_session#383 bytes [4313589, 4313614]: "OperationCancelledError"
                v148 = sub_180371890(a1, v147, qword_1803BEB08);
              v150 = v148;
              if ( *v147 >= 0 )
              {
                v214 = (*(_QWORD *)v147)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v147 + 1) + 48LL))(v147);
              }
              if ( v150 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 24;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 826 | qword_1803BEB08 -> novels.services.agent_session#383 bytes [4313589, 4313614]: "OperationCancelledError"
              sub_1801F3490(v149, qword_1803BEB08, v150);
// Original IDA line 827 | qword_1803BE018 -> novels.services.agent_session#33 bytes [4307388, 4307412]: "utils.knowledge_bridge"
              v151 = qword_1803BE018;
              v152 = qword_1803CA5B8;
// Original IDA line 829 | qword_1803BE028 -> novels.services.agent_session#35 bytes [4307439, 4307441]: 2
              v245 = qword_1803BE028;
// Original IDA line 830 | qword_1803BEB10 -> novels.services.agent_session#384 bytes [4313614, 4313636]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [4313616, 4313636], "value": "list_style_sources"}]}
              v231 = qword_1803BEB10;
              *(_DWORD *)(v246 + 40) = 25;
              v153 = sub_1803717E0(a1, v151, v152, Py_NoneStruct, v231, v245);
              v154 = (int *)v153;
              if ( v153 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 25;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              if ( *(_QWORD *)(v153 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 842 | qword_1803BDFD0 -> novels.services.agent_session#24 bytes [4307268, 4307288]: "list_style_sources"
                v155 = sub_180371D80(a1, (_DWORD)v154, qword_1803CA5B8, qword_1803BDFD0, qword_1803AB250);
              else
// Original IDA line 844 | qword_1803BDFD0 -> novels.services.agent_session#24 bytes [4307268, 4307288]: "list_style_sources"
                v155 = sub_180371890(a1, v154, qword_1803BDFD0);
              v157 = v155;
              if ( *v154 >= 0 )
              {
                v214 = (*(_QWORD *)v154)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v154 + 1) + 48LL))(v154);
              }
              if ( v157 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 25;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 859 | qword_1803BDFD0 -> novels.services.agent_session#24 bytes [4307268, 4307288]: "list_style_sources"
              sub_1801F3490(v156, qword_1803BDFD0, v157);
              v158 = sub_1801F4010();
              if ( v158 == 0 )
              {
// Original IDA line 863 | qword_1803BEA10 -> novels.services.agent_session#352 bytes [4312888, 4312892]: "re"
                sub_180373340(a1, &v248, qword_1803BEA10);
                v42 = v248;
                v43 = 27;
                goto LABEL_345;
              }
// Original IDA line 868 | qword_1803BEB18 -> novels.services.agent_session#385 bytes [4313636, 4313645]: "compile"
              v160 = (int *)sub_18037F0C0(v159, v158, qword_1803BEB18);
              if ( v160 == nullptr )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 27;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 876 | qword_1803BEB20 -> novels.services.agent_session#386 bytes [4313645, 4313683]: "<<<ACTION>>>\\s*(\\{.*?\\})\\s*<<<END>>>"
              v161 = qword_1803BEB20;
              v162 = sub_1801F4010();
              if ( v162 == 0 )
              {
// Original IDA line 880 | qword_1803BEA10 -> novels.services.agent_session#352 bytes [4312888, 4312892]: "re"
                sub_180373340(a1, &v248, qword_1803BEA10);
                if ( *v160 >= 0 )
                {
                  v214 = (*(_QWORD *)v160)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v160 + 1) + 48LL))(v160);
                }
                v42 = v248;
                v43 = 27;
                goto LABEL_345;
              }
// Original IDA line 891 | qword_1803BEB28 -> novels.services.agent_session#387 bytes [4313683, 4313691]: "DOTALL"
              v164 = sub_18037F0C0(v163, v162, qword_1803BEB28);
              v165 = (int *)v164;
              if ( v164 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v160 >= 0 )
                {
                  v214 = (*(_QWORD *)v160)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v160 + 1) + 48LL))(v160);
                }
                v43 = 27;
                goto LABEL_345;
              }
              *(_DWORD *)(v246 + 40) = 27;
              v247[0] = v161;
              v247[1] = v164;
              v167 = sub_180377750(a1, v160, v247);
              if ( *v160 >= 0 )
              {
                v214 = (*(_QWORD *)v160)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v160 + 1) + 48LL))(v160);
              }
              if ( *v165 >= 0 )
              {
                v214 = (*(_QWORD *)v165)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v165 + 1) + 48LL))(v165);
              }
              if ( v167 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 27;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
// Original IDA line 929 | qword_1803BE9B0 -> novels.services.agent_session#340 bytes [4312774, 4312790]: "ACTION_PATTERN"
              sub_1801F3490(v166, qword_1803BE9B0, v167);
              v168 = (int *)sub_180370EC0(a1);
              v169 = sub_180370EC0(a1);
              if ( *(_DWORD *)v169 != -1 )
                ++*(_DWORD *)v169;
// Original IDA line 934 | qword_1803BED18 -> novels.services.agent_session#449 bytes [4317722, 4317732]: "__name__"
              v170 = (__int64 *)sub_1800015A0(qword_1803CA5B8, qword_1803BED18);
              if ( v170 == nullptr || (v171 = *v170, *v170 == 0) )
              {
// Original IDA line 937 | qword_1803BED18 -> novels.services.agent_session#449 bytes [4317722, 4317732]: "__name__"
                v172 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803BED18);
                v171 = 0;
                if ( v172 != nullptr )
                  v171 = *v172;
              }
// Original IDA line 942 | qword_1803BEB30 -> novels.services.agent_session#388 bytes [4313691, 4313703]: "__module__"
              PyDict_SetItem(v169, qword_1803BEB30, v171);
// Original IDA line 943 | qword_1803BEB38 -> novels.services.agent_session#389 bytes [4313703, 4313724]: "AgentSessionService"; qword_1803BEB40 -> novels.services.agent_session#390 bytes [4313724, 4313738]: "__qualname__"
              PyDict_SetItem(v169, qword_1803BEB40, qword_1803BEB38);
// Original IDA line 944 | qword_1803BEB48 -> novels.services.agent_session#391 bytes [4313738, 4313744]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [4313740, 4313741], "value": null}, {"tag": "n", "fileRange": [4313741, 4313742], "value": null}, {"tag": "n", "fileRange": [4313742, 4313743], "value": null}, {"tag": "s", "fileRange": ... (truncated literal)
              v173 = (_DWORD *)qword_1803BEB48;
// Original IDA line 945 | qword_1803BEB50 -> novels.services.agent_session#392 bytes [4313744, 4313984]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4313746, 4313753], "value": "state"}, {"tag": "a", "fileRange": [4313753, 4313763], "value": "pipeline"}, {"tag": "a", "fileRange": [4313763, 4313779], "value": "cancel_checker"}, {"tag":... (truncated literal)
              v174 = sub_18036FB80(a1, qword_1803BEB50);
              if ( *v173 != -1 )
                ++*v173;
              v175 = (int *)sub_1803609A0(
// Original IDA line 949 | sub_1801F4B80 -> AgentSessionService.__init__
                              (int)sub_1801F4B80,
// Original IDA line 950 | qword_1803BEB58 -> novels.services.agent_session#393 bytes [4313984, 4313994]: "__init__"
                              qword_1803BEB58,
// Original IDA line 951 | qword_1803BEB60 -> novels.services.agent_session#394 bytes [4313994, 4314024]: "AgentSessionService.__init__"
                              qword_1803BEB60,
                              qword_1803BDE28,
                              (__int64)v173,
                              0,
                              v174,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 960 | qword_1803BEB58 -> novels.services.agent_session#393 bytes [4313984, 4313994]: "__init__"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEB58, v175);
              if ( *v175 >= 0 )
              {
                v214 = (*(_QWORD *)v175)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v175 + 1) + 48LL))(v175);
              }
// Original IDA line 967 | qword_1803BEB68 -> novels.services.agent_session#395 bytes [4314024, 4314071]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314026, 4314038], "value": "event_type"}, {"tag": "a", "fileRange": [4314038, 4314047], "value": "payload"}, {"tag": "a", "fileRange": [4314047, 4314055], "value": "return"}], "values": ... (truncated literal)
              v176 = sub_18036FB80(a1, qword_1803BEB68);
              v177 = (int *)sub_1803609A0(
// Original IDA line 969 | sub_1801F5210 -> AgentSessionService._emit
                              (int)sub_1801F5210,
// Original IDA line 970 | qword_1803BE0F8 -> novels.services.agent_session#61 bytes [4307721, 4307728]: "_emit"
                              qword_1803BE0F8,
// Original IDA line 971 | qword_1803BEB70 -> novels.services.agent_session#396 bytes [4314071, 4314098]: "AgentSessionService._emit"
                              qword_1803BEB70,
                              qword_1803BDD70,
                              0,
                              0,
                              v176,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 980 | qword_1803BE0F8 -> novels.services.agent_session#61 bytes [4307721, 4307728]: "_emit"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE0F8, v177);
              if ( *v177 >= 0 )
              {
                v214 = (*(_QWORD *)v177)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v177 + 1) + 48LL))(v177);
              }
// Original IDA line 987 | qword_1803BEB78 -> novels.services.agent_session#397 bytes [4314098, 4314101]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [4314100, 4314101], "value": null}]}
              v178 = (_DWORD *)qword_1803BEB78;
// Original IDA line 988 | qword_1803BEB80 -> novels.services.agent_session#398 bytes [4314101, 4314148]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314103, 4314110], "value": "model"}, {"tag": "a", "fileRange": [4314110, 4314118], "value": "return"}], "values": [{"tag": "u", "fileRange": [4314118, 4314133], "value": "Optional[str]"}... (truncated literal)
              v179 = sub_18036FB80(a1, qword_1803BEB80);
              if ( *v178 != -1 )
                ++*v178;
              v180 = (int *)sub_1803609A0(
// Original IDA line 992 | sub_1801F5880 -> AgentSessionService._resolve_chat_model
                              (int)sub_1801F5880,
// Original IDA line 993 | qword_1803BDF88 -> novels.services.agent_session#15 bytes [4307129, 4307150]: "_resolve_chat_model"
                              qword_1803BDF88,
// Original IDA line 994 | qword_1803BEB88 -> novels.services.agent_session#399 bytes [4314148, 4314189]: "AgentSessionService._resolve_chat_model"
                              qword_1803BEB88,
                              qword_1803BDDE8,
                              (__int64)v178,
                              0,
                              v179,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1003 | qword_1803BDF88 -> novels.services.agent_session#15 bytes [4307129, 4307150]: "_resolve_chat_model"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BDF88, v180);
              if ( *v180 >= 0 )
              {
                v214 = (*(_QWORD *)v180)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v180 + 1) + 48LL))(v180);
              }
// Original IDA line 1010 | qword_1803BEB78 -> novels.services.agent_session#397 bytes [4314098, 4314101]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [4314100, 4314101], "value": null}]}
              v181 = (_DWORD *)qword_1803BEB78;
// Original IDA line 1011 | qword_1803BEB90 -> novels.services.agent_session#400 bytes [4314189, 4314296]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314191, 4314211], "value": "knowledge_book_ids"}, {"tag": "a", "fileRange": [4314211, 4314226], "value": "thinking_mode"}, {"tag": "a", "fileRange": [4314226, 4314242], "value": "context... (truncated literal)
              v182 = sub_18036FB80(a1, qword_1803BEB90);
// Original IDA line 1012 | qword_1803BEB98 -> novels.services.agent_session#401 bytes [4314296, 4314548]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314298, 4314310], "value": "session_id"}, {"tag": "a", "fileRange": [4314310, 4314324], "value": "user_message"}, {"tag": "a", "fileRange": [4314324, 4314331], "value": "model"}, {"tag":... (truncated literal)
              v183 = sub_18036FB80(a1, qword_1803BEB98);
              if ( *v181 != -1 )
                ++*v181;
              v184 = (int *)sub_1803609A0(
// Original IDA line 1016 | sub_1801F5CB0 -> AgentSessionService.send_message
                              (int)sub_1801F5CB0,
// Original IDA line 1017 | qword_1803BEBA0 -> novels.services.agent_session#402 bytes [4314548, 4314562]: "send_message"
                              qword_1803BEBA0,
// Original IDA line 1018 | qword_1803BEBA8 -> novels.services.agent_session#403 bytes [4314562, 4314596]: "AgentSessionService.send_message"
                              qword_1803BEBA8,
                              qword_1803BDDE0,
                              (__int64)v181,
                              v182,
                              v183,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1027 | qword_1803BEBA0 -> novels.services.agent_session#402 bytes [4314548, 4314562]: "send_message"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBA0, v184);
              if ( *v184 >= 0 )
              {
                v214 = (*(_QWORD *)v184)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v184 + 1) + 48LL))(v184);
              }
// Original IDA line 1034 | qword_1803BEB78 -> novels.services.agent_session#397 bytes [4314098, 4314101]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [4314100, 4314101], "value": null}]}
              v185 = (_DWORD *)qword_1803BEB78;
// Original IDA line 1035 | qword_1803BEBB0 -> novels.services.agent_session#404 bytes [4314596, 4314661]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314598, 4314610], "value": "session_id"}, {"tag": "a", "fileRange": [4314610, 4314617], "value": "model"}, {"tag": "a", "fileRange": [4314617, 4314625], "value": "return"}], "values": [{... (truncated literal)
              v186 = sub_18036FB80(a1, qword_1803BEBB0);
              if ( *v185 != -1 )
                ++*v185;
              v187 = (int *)sub_1803609A0(
// Original IDA line 1039 | sub_1801FD690 -> AgentSessionService.confirm_action
                              (int)sub_1801FD690,
// Original IDA line 1040 | qword_1803BEBB8 -> novels.services.agent_session#405 bytes [4314661, 4314677]: "confirm_action"
                              qword_1803BEBB8,
// Original IDA line 1041 | qword_1803BEBC0 -> novels.services.agent_session#406 bytes [4314677, 4314713]: "AgentSessionService.confirm_action"
                              qword_1803BEBC0,
                              qword_1803BDE20,
                              (__int64)v185,
                              0,
                              v186,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1050 | qword_1803BEBB8 -> novels.services.agent_session#405 bytes [4314661, 4314677]: "confirm_action"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBB8, v187);
              if ( *v187 >= 0 )
              {
                v214 = (*(_QWORD *)v187)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v187 + 1) + 48LL))(v187);
              }
// Original IDA line 1057 | qword_1803BEB78 -> novels.services.agent_session#397 bytes [4314098, 4314101]: {"type": "tuple", "items": [{"tag": "n", "fileRange": [4314100, 4314101], "value": null}]}
              v188 = (_DWORD *)qword_1803BEB78;
// Original IDA line 1058 | qword_1803BEBC8 -> novels.services.agent_session#407 bytes [4314713, 4314803]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314715, 4314727], "value": "session_id"}, {"tag": "a", "fileRange": [4314727, 4314736], "value": "payload"}, {"tag": "a", "fileRange": [4314736, 4314743], "value": "model"}, {"tag": "a",... (truncated literal)
              v189 = sub_18036FB80(a1, qword_1803BEBC8);
              if ( *v188 != -1 )
                ++*v188;
              v190 = (int *)sub_1803609A0(
// Original IDA line 1062 | sub_1802095A0 -> AgentSessionService.confirm_pipeline_checkpoint
                              (int)sub_1802095A0,
// Original IDA line 1063 | qword_1803BEBD0 -> novels.services.agent_session#408 bytes [4314803, 4314832]: "confirm_pipeline_checkpoint"
                              qword_1803BEBD0,
// Original IDA line 1064 | qword_1803BEBD8 -> novels.services.agent_session#409 bytes [4314832, 4314881]: "AgentSessionService.confirm_pipeline_checkpoint"
                              qword_1803BEBD8,
                              qword_1803BDE30,
                              (__int64)v188,
                              0,
                              v189,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1073 | qword_1803BEBD0 -> novels.services.agent_session#408 bytes [4314803, 4314832]: "confirm_pipeline_checkpoint"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBD0, v190);
              if ( *v190 >= 0 )
              {
                v214 = (*(_QWORD *)v190)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v190 + 1) + 48LL))(v190);
              }
// Original IDA line 1080 | qword_1803BEBE0 -> novels.services.agent_session#410 bytes [4314881, 4314921]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314883, 4314892], "value": "session"}, {"tag": "a", "fileRange": [4314892, 4314900], "value": "return"}], "values": [{"tag": "a", "fileRange": [4314900, 4314915], "value": "SessionRecord... (truncated literal)
              v191 = sub_18036FB80(a1, qword_1803BEBE0);
              v192 = (int *)sub_1803609A0(
// Original IDA line 1082 | sub_18020DF00 -> AgentSessionService._should_expose_propose_tool
                              (int)sub_18020DF00,
// Original IDA line 1083 | qword_1803BE130 -> novels.services.agent_session#68 bytes [4307891, 4307920]: "_should_expose_propose_tool"
                              qword_1803BE130,
// Original IDA line 1084 | qword_1803BEBE8 -> novels.services.agent_session#411 bytes [4314921, 4314970]: "AgentSessionService._should_expose_propose_tool"
                              qword_1803BEBE8,
                              qword_1803BDEF0,
                              0,
                              0,
                              v191,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              v193 = (int *)sub_18036F0F0(a1, v192);
              if ( *v192 >= 0 )
              {
                v214 = (*(_QWORD *)v192)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v192 + 1) + 48LL))(v192);
              }
// Original IDA line 1100 | qword_1803BE130 -> novels.services.agent_session#68 bytes [4307891, 4307920]: "_should_expose_propose_tool"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE130, v193);
              if ( *v193 >= 0 )
              {
                v214 = (*(_QWORD *)v193)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v193 + 1) + 48LL))(v193);
              }
// Original IDA line 1107 | qword_1803BEBF0 -> novels.services.agent_session#412 bytes [4314970, 4315061]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4314972, 4314981], "value": "session"}, {"tag": "a", "fileRange": [4314981, 4314995], "value": "user_message"}, {"tag": "a", "fileRange": [4314995, 4315002], "value": "model"}, {"tag": "a... (truncated literal)
              v194 = sub_18036FB80(a1, qword_1803BEBF0);
              v195 = (int *)sub_1803609A0(
// Original IDA line 1109 | sub_18020E2D0 -> AgentSessionService._handle_play_message
                              (int)sub_18020E2D0,
// Original IDA line 1110 | qword_1803BDFB8 -> novels.services.agent_session#21 bytes [4307220, 4307242]: "_handle_play_message"
                              qword_1803BDFB8,
// Original IDA line 1111 | qword_1803BEBF8 -> novels.services.agent_session#413 bytes [4315061, 4315103]: "AgentSessionService._handle_play_message"
                              qword_1803BEBF8,
                              qword_1803BDEA8,
                              0,
                              0,
                              v194,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1120 | qword_1803BDFB8 -> novels.services.agent_session#21 bytes [4307220, 4307242]: "_handle_play_message"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BDFB8, v195);
              if ( *v195 >= 0 )
              {
                v214 = (*(_QWORD *)v195)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v195 + 1) + 48LL))(v195);
              }
// Original IDA line 1127 | qword_1803BEC00 -> novels.services.agent_session#414 bytes [4315103, 4315167]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4315105, 4315116], "value": "tool_name"}, {"tag": "a", "fileRange": [4315116, 4315125], "value": "summary"}, {"tag": "a", "fileRange": [4315125, 4315132], "value": "extra"}, {"tag": "a", ... (truncated literal)
              v196 = sub_18036FB80(a1, qword_1803BEC00);
              v197 = (int *)sub_1803609A0(
// Original IDA line 1129 | sub_180213700 -> AgentSessionService._tool_message
                              (int)sub_180213700,
// Original IDA line 1130 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
                              qword_1803BE2D8,
// Original IDA line 1131 | qword_1803BEC08 -> novels.services.agent_session#415 bytes [4315167, 4315202]: "AgentSessionService._tool_message"
                              qword_1803BEC08,
                              qword_1803BDDF8,
                              0,
                              0,
                              v196,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1140 | qword_1803BE2D8 -> novels.services.agent_session#121 bytes [4308756, 4308771]: "_tool_message"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE2D8, v197);
              if ( *v197 >= 0 )
              {
                v214 = (*(_QWORD *)v197)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v197 + 1) + 48LL))(v197);
              }
// Original IDA line 1147 | qword_1803BEC10 -> novels.services.agent_session#416 bytes [4315202, 4315238]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4315204, 4315209], "value": "raw"}, {"tag": "a", "fileRange": [4315209, 4315217], "value": "return"}], "values": [{"tag": "a", "fileRange": [4315217, 4315222], "value": "str"}, {"tag": "u... (truncated literal)
              v198 = sub_18036FB80(a1, qword_1803BEC10);
              v199 = (int *)sub_1803609A0(
// Original IDA line 1149 | sub_180213E10 -> AgentSessionService._normalize_propose_action
                              (int)sub_180213E10,
// Original IDA line 1150 | qword_1803BE1A0 -> novels.services.agent_session#82 bytes [4308126, 4308153]: "_normalize_propose_action"
                              qword_1803BE1A0,
// Original IDA line 1151 | qword_1803BEC18 -> novels.services.agent_session#417 bytes [4315238, 4315285]: "AgentSessionService._normalize_propose_action"
                              qword_1803BEC18,
                              qword_1803BDE60,
                              0,
                              0,
                              v198,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1160 | qword_1803BE1A0 -> novels.services.agent_session#82 bytes [4308126, 4308153]: "_normalize_propose_action"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE1A0, v199);
              if ( *v199 >= 0 )
              {
                v214 = (*(_QWORD *)v199)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v199 + 1) + 48LL))(v199);
              }
// Original IDA line 1167 | qword_1803BEC20 -> novels.services.agent_session#418 bytes [4315285, 4315309]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4315287, 4315295], "value": "action"}, {"tag": "a", "fileRange": [4315295, 4315303], "value": "return"}], "values": [{"tag": "a", "fileRange": [4315303, 4315308], "value": "str"}, {"tag":... (truncated literal)
              v200 = sub_18036FB80(a1, qword_1803BEC20);
              v201 = (int *)sub_1803609A0(
// Original IDA line 1169 | sub_180215D90 -> AgentSessionService._route_for_action
                              (int)sub_180215D90,
// Original IDA line 1170 | qword_1803BE1E0 -> novels.services.agent_session#90 bytes [4308261, 4308280]: "_route_for_action"
                              qword_1803BE1E0,
// Original IDA line 1171 | qword_1803BEC28 -> novels.services.agent_session#419 bytes [4315309, 4315348]: "AgentSessionService._route_for_action"
                              qword_1803BEC28,
                              qword_1803BDE88,
                              0,
                              0,
                              v200,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1180 | qword_1803BE1E0 -> novels.services.agent_session#90 bytes [4308261, 4308280]: "_route_for_action"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE1E0, v201);
              if ( *v201 >= 0 )
              {
                v214 = (*(_QWORD *)v201)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v201 + 1) + 48LL))(v201);
              }
// Original IDA line 1187 | qword_1803BEC30 -> novels.services.agent_session#420 bytes [4315348, 4315410]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4315350, 4315359], "value": "content"}, {"tag": "a", "fileRange": [4315359, 4315367], "value": "return"}], "values": [{"tag": "a", "fileRange": [4315367, 4315372], "value": "str"}, {"tag"... (truncated literal)
              v202 = sub_18036FB80(a1, qword_1803BEC30);
              v203 = (int *)sub_1803609A0(
// Original IDA line 1189 | sub_1802160A0 -> AgentSessionService._split_action
                              (int)sub_1802160A0,
// Original IDA line 1190 | qword_1803BE228 -> novels.services.agent_session#99 bytes [4308380, 4308395]: "_split_action"
                              qword_1803BE228,
// Original IDA line 1191 | qword_1803BEC38 -> novels.services.agent_session#421 bytes [4315410, 4315445]: "AgentSessionService._split_action"
                              qword_1803BEC38,
                              qword_1803BDE58,
                              0,
                              0,
                              v202,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1200 | qword_1803BE228 -> novels.services.agent_session#99 bytes [4308380, 4308395]: "_split_action"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE228, v203);
              if ( *v203 >= 0 )
              {
                v214 = (*(_QWORD *)v203)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v203 + 1) + 48LL))(v203);
              }
// Original IDA line 1207 | qword_1803BEC40 -> novels.services.agent_session#422 bytes [4315445, 4315471]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [4315447, 4315452], "value": "key"}, {"tag": "a", "fileRange": [4315452, 4315460], "value": "return"}], "values": [{"tag": "a", "fileRange": [4315460, 4315465], "value": "str"}, {"tag": "a... (truncated literal)
              v204 = sub_18036FB80(a1, qword_1803BEC40);
              v205 = (int *)sub_1803609A0(
// Original IDA line 1209 | sub_180216D20 -> AgentSessionService._is_cancelled
                              (int)sub_180216D20,
// Original IDA line 1210 | qword_1803BE2A8 -> novels.services.agent_session#115 bytes [4308651, 4308666]: "_is_cancelled"
                              qword_1803BE2A8,
// Original IDA line 1211 | qword_1803BEC48 -> novels.services.agent_session#423 bytes [4315471, 4315506]: "AgentSessionService._is_cancelled"
                              qword_1803BEC48,
                              qword_1803BDE50,
                              0,
                              0,
                              v204,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
// Original IDA line 1220 | qword_1803BE2A8 -> novels.services.agent_session#115 bytes [4308651, 4308666]: "_is_cancelled"
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE2A8, v205);
              if ( *v205 >= 0 )
              {
                v214 = (*(_QWORD *)v205)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v205 + 1) + 48LL))(v205);
              }
              v206 = *(_QWORD *)(a1 + 16);
// Original IDA line 1228 | qword_1803BEB38 -> novels.services.agent_session#389 bytes [4313703, 4313724]: "AgentSessionService"
              v207 = (_DWORD *)qword_1803BEB38;
              v208 = *(int **)(v206 + 269736);
              if ( v208 != nullptr )
              {
                *(_QWORD *)(v206 + 269736) = *((_QWORD *)v208 + 3);
                --*(_DWORD *)(v206 + 269888);
                *(_QWORD *)v208 = 1;
              }
              else
              {
                v208 = (int *)sub_18035EBD0(PyTuple_Type, 3);
              }
              *((_QWORD *)v208 + 3) = 0;
              *((_QWORD *)v208 + 4) = 0;
              *((_QWORD *)v208 + 5) = 0;
              sub_1801F33C0(v208);
              if ( *v207 != -1 )
                ++*v207;
              *((_QWORD *)v208 + 3) = v207;
              v209 = qword_1803AB240;
              if ( *(_DWORD *)qword_1803AB240 != -1 )
                ++*(_DWORD *)qword_1803AB240;
              *((_QWORD *)v208 + 4) = v209;
              if ( *(_DWORD *)v169 != -1 )
                ++*(_DWORD *)v169;
              *((_QWORD *)v208 + 5) = v169;
              *(_DWORD *)(v246 + 40) = 30;
              v211 = sub_180001790(a1, PyType_Type, v208, v168);
              if ( *v208 >= 0 )
              {
                v214 = (*(_QWORD *)v208)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v208 + 1) + 48LL))(v208);
              }
              if ( v211 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *(int *)v169 >= 0 )
                {
                  v214 = (*(_QWORD *)v169)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(__int64))(*(_QWORD *)(v169 + 8) + 48LL))(v169);
                }
                if ( *v168 >= 0 )
                {
                  v214 = (*(_QWORD *)v168)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v168 + 1) + 48LL))(v168);
                }
                if ( *(int *)v169 >= 0 )
                {
                  v214 = (*(_QWORD *)v169)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(__int64))(*(_QWORD *)(v169 + 8) + 48LL))(v169);
                }
                v35 = v246;
                v43 = 30;
LABEL_345:
                v212 = *(_QWORD *)(v42 + 40);
                if ( v212 != 0 )
                {
                  if ( *(_QWORD *)(v212 + 24) == v35 )
                  {
LABEL_363:
                    sub_180001D80(a1, v41);
                    v218 = *(int **)(a1 + 96);
                    *(_QWORD *)(a1 + 96) = v42;
                    if ( v218 != nullptr && *v218 >= 0 )
                    {
                      v214 = (*(_QWORD *)v218)-- == 1;
                      if ( v214 )
                        (*(void (**)(void))(*((_QWORD *)v218 + 1) + 48LL))();
                    }
                    return 0;
                  }
                  v213 = sub_18038E2A0(v35, v43);
                  v215 = v213;
                  *(_QWORD *)(v213 + 16) = v212;
                  if ( *(_DWORD *)v212 != -1 )
                    ++*(_DWORD *)v212;
                  v214 = v213 == Py_NoneStruct;
                }
                else
                {
                  v213 = sub_18038E2A0(v35, v43);
                  v214 = v213 == Py_NoneStruct;
                  v215 = v213;
                }
                v216 = *(int **)(v42 + 40);
                if ( !v214 )
                {
                  v11 = v215;
                  if ( v215 != 0 )
                  {
                    v217 = *(_DWORD *)v213 + 1;
                    if ( v217 != 0 )
                      *(_DWORD *)v215 = v217;
                  }
                }
                *(_QWORD *)(v42 + 40) = v11;
                if ( v216 != nullptr && *v216 >= 0 )
                {
                  v214 = (*(_QWORD *)v216)-- == 1;
                  if ( v214 )
                    (*(void (**)(void))(*((_QWORD *)v216 + 1) + 48LL))();
                }
                if ( v215 != 0 && *(int *)v215 >= 0 )
                {
                  v214 = (*(_QWORD *)v215)-- == 1;
                  if ( v214 )
                    (*(void (__fastcall **)(__int64))(*(_QWORD *)(v215 + 8) + 48LL))(v215);
                }
                goto LABEL_363;
              }
              if ( *(_DWORD *)v211 != -1 )
                ++*(_DWORD *)v211;
              if ( *(int *)v169 >= 0 )
              {
                v214 = (*(_QWORD *)v169)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v169 + 8) + 48LL))(v169);
              }
              if ( *(int *)v211 >= 0 )
              {
                v214 = (*(_QWORD *)v211)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v211 + 8) + 48LL))(v211);
              }
// Original IDA line 1357 | qword_1803BEB38 -> novels.services.agent_session#389 bytes [4313703, 4313724]: "AgentSessionService"
              sub_1801F3490(v210, qword_1803BEB38, v211);
              sub_180001D80(a1, v220);
              if ( *v168 >= 0 )
              {
                v214 = (*(_QWORD *)v168)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v168 + 1) + 48LL))(v168);
              }
              if ( *(int *)v169 >= 0 )
              {
                v214 = (*(_QWORD *)v169)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v169 + 8) + 48LL))(v169);
              }
              result = qword_1803CA5C0;
              if ( *(_DWORD *)qword_1803CA5C0 != -1 )
                ++*(_DWORD *)qword_1803CA5C0;
              return result;
            }
          }
        }
      }
    }
  }
  v43 = 6;
LABEL_85:
  v42 = *(_QWORD *)(a1 + 96);
  *(_QWORD *)(a1 + 96) = 0;
  if ( *v64 >= 0 )
  {
    v214 = (*(_QWORD *)v64)-- == 1;
    if ( v214 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
  }
  goto LABEL_345;
}
