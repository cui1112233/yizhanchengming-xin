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
  sub_1801F3400(v31, qword_1803BE9E0, Py_NoneStruct);
  sub_1801F3400(v32, qword_1803BE9E8, qword_1803BDF00);
  v33 = qword_1803CA5C0;
  v34 = *(_DWORD **)(qword_1803CA5C0 + 16);
  if ( *v34 != -1 )
    ++*v34;
  v246 = sub_18036CBB0(qword_1803BED38, v33, v34, 0);
  v35 = v246;
  sub_180001D00(a1, v246);
  v36 = qword_1803BDF00;
  v37 = (__int64 *)sub_1800015A0(qword_1803CA5B8, qword_1803BED20);
  if ( v37 == nullptr || (v39 = *v37, *v37 == 0) )
  {
    v40 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803BED20);
    v39 = 0;
    if ( v40 != nullptr )
      v39 = *v40;
  }
  if ( (unsigned __int8)sub_18037F3A0(v38, v39, qword_1803BE9F0, v36) == 0
    || (v44 = sub_1801F3740(), (unsigned __int8)sub_18037F3A0(v45, v44, qword_1803BE9F8, Py_TrueStruct) == 0) )
  {
    v42 = *(_QWORD *)(a1 + 96);
    v43 = 1;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_345;
  }
  sub_1801F3400(v46, qword_1803BEA00, Py_NoneStruct);
  v47 = sub_180372070();
  v49 = sub_18037F0C0(v48, v47, qword_1803BEA08);
  sub_1801F3490(v50, qword_1803BEA08, v49);
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
  sub_1801F3490(v54, qword_1803BE850, v53);
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
  sub_1801F3490(v58, qword_1803BEA10, v57);
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
  sub_1801F3490(v62, qword_1803BE0F0, v61);
  v63 = sub_180372110();
  v64 = (int *)v63;
  if ( *(_DWORD *)v63 != -1 )
    ++*(_DWORD *)v63;
  if ( *(_QWORD *)(v63 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
    v65 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA18, qword_1803AB250);
  else
    v65 = sub_180371890(a1, v64, qword_1803BEA18);
  if ( v65 != 0 )
  {
    sub_1801F3490(v66, qword_1803BEA18, v65);
    if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
      v67 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA20, qword_1803AB250);
    else
      v67 = sub_180371890(a1, v64, qword_1803BEA20);
    if ( v67 != 0 )
    {
      sub_1801F3490(v68, qword_1803BEA20, v67);
      if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
        v69 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA28, qword_1803AB250);
      else
        v69 = sub_180371890(a1, v64, qword_1803BEA28);
      if ( v69 != 0 )
      {
        sub_1801F3490(v70, qword_1803BEA28, v69);
        if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
          v71 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA30, qword_1803AB250);
        else
          v71 = sub_180371890(a1, v64, qword_1803BEA30);
        if ( v71 != 0 )
        {
          sub_1801F3490(v72, qword_1803BEA30, v71);
          if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
            v73 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA38, qword_1803AB250);
          else
            v73 = sub_180371890(a1, v64, qword_1803BEA38);
          if ( v73 != 0 )
          {
            sub_1801F3490(v74, qword_1803BEA38, v73);
            if ( *((_QWORD *)v64 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
              v75 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA40, qword_1803AB250);
            else
              v75 = sub_180371890(a1, v64, qword_1803BEA40);
            if ( v75 != 0 )
            {
              sub_1801F3490(v76, qword_1803BEA40, v75);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v77 = qword_1803BEA48;
              v78 = qword_1803CA5B8;
              v235 = qword_1803BE028;
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
                v80 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA58, qword_1803AB250);
              else
                v80 = sub_180371890(a1, v64, qword_1803BEA58);
              if ( v80 == 0
                || ((sub_1801F3490(v81, qword_1803BEA58, v80), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
                  ? (v82 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE0C8, qword_1803AB250))
                  : (v82 = sub_180371890(a1, v64, qword_1803BE0C8)),
                    v82 == 0) )
              {
                v43 = 8;
                goto LABEL_85;
              }
              sub_1801F3490(v83, qword_1803BE0C8, v82);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v84 = qword_1803BEA60;
              v85 = qword_1803CA5B8;
              v236 = qword_1803BE028;
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
                v88 = sub_180371D80(a1, (_DWORD)v87, qword_1803CA5B8, qword_1803BEA70, qword_1803AB250);
              else
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
              sub_1801F3490(v89, qword_1803BEA70, v90);
              v91 = qword_1803BEA78;
              v92 = qword_1803CA5B8;
              v237 = qword_1803BE028;
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
                v94 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BEA88, qword_1803AB250);
              else
                v94 = sub_180371890(a1, v64, qword_1803BEA88);
              if ( v94 == 0
                || ((sub_1801F3490(v95, qword_1803BEA88, v94), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
                  ? (v96 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE830, qword_1803AB250))
                  : (v96 = sub_180371890(a1, v64, qword_1803BE830)),
                    v96 == 0
                 || ((sub_1801F3490(v97, qword_1803BE830, v96), *((_QWORD *)v64 + 1) == PyModule_Type)
                  || (unsigned int)PyType_IsSubtype() != 0
                   ? (v98 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE138, qword_1803AB250))
                   : (v98 = sub_180371890(a1, v64, qword_1803BE138)),
                     v98 == 0)) )
              {
                v43 = 10;
                goto LABEL_85;
              }
              sub_1801F3490(v99, qword_1803BE138, v98);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v100 = qword_1803BEA90;
              v101 = qword_1803CA5B8;
              v238 = qword_1803BE028;
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
                v104 = sub_180371D80(a1, (_DWORD)v103, qword_1803CA5B8, qword_1803BE008, qword_1803AB250);
              else
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
              sub_1801F3490(v105, qword_1803BE008, v106);
              v107 = qword_1803BEAA0;
              v108 = qword_1803CA5B8;
              v239 = qword_1803BE028;
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
                v111 = sub_180371D80(a1, (_DWORD)v110, qword_1803CA5B8, qword_1803BDF60, qword_1803AB250);
              else
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
              sub_1801F3490(v112, qword_1803BDF60, v113);
              v114 = qword_1803BEAB0;
              v115 = qword_1803CA5B8;
              v240 = qword_1803BE028;
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
                v117 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE5A0, qword_1803AB250);
              else
                v117 = sub_180371890(a1, v64, qword_1803BE5A0);
              if ( v117 == 0
                || ((sub_1801F3490(v118, qword_1803BE5A0, v117), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
                  ? (v119 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE538, qword_1803AB250))
                  : (v119 = sub_180371890(a1, v64, qword_1803BE538)),
                    v119 == 0
                 || ((sub_1801F3490(v120, qword_1803BE538, v119), *((_QWORD *)v64 + 1) == PyModule_Type)
                  || (unsigned int)PyType_IsSubtype() != 0
                   ? (v121 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE548, qword_1803AB250))
                   : (v121 = sub_180371890(a1, v64, qword_1803BE548)),
                     v121 == 0)) )
              {
                v43 = 17;
                goto LABEL_85;
              }
              sub_1801F3490(v122, qword_1803BE548, v121);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v123 = qword_1803BEAC0;
              v124 = qword_1803CA5B8;
              v241 = qword_1803BE028;
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
                v126 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE1D8, qword_1803AB250);
              else
                v126 = sub_180371890(a1, v64, qword_1803BE1D8);
              if ( v126 == 0
                || ((sub_1801F3490(v127, qword_1803BE1D8, v126), *((_QWORD *)v64 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
                  ? (v128 = sub_180371D80(a1, (_DWORD)v64, qword_1803CA5B8, qword_1803BE1B0, qword_1803AB250))
                  : (v128 = sub_180371890(a1, v64, qword_1803BE1B0)),
                    v128 == 0) )
              {
                v43 = 18;
                goto LABEL_85;
              }
              sub_1801F3490(v129, qword_1803BE1B0, v128);
              if ( *v64 >= 0 )
              {
                v214 = (*(_QWORD *)v64)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v130 = qword_1803BEAD0;
              v131 = qword_1803CA5B8;
              v242 = qword_1803BE028;
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
                v134 = sub_180371D80(a1, (_DWORD)v133, qword_1803CA5B8, qword_1803BEAE0, qword_1803AB250);
              else
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
              sub_1801F3490(v135, qword_1803BEAE0, v136);
              v137 = qword_1803BEAE8;
              v138 = qword_1803CA5B8;
              v243 = qword_1803BE028;
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
                v141 = sub_180371D80(a1, (_DWORD)v140, qword_1803CA5B8, qword_1803BDF28, qword_1803AB250);
              else
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
              sub_1801F3490(v142, qword_1803BDF28, v143);
              v144 = qword_1803BEAF8;
              v145 = qword_1803CA5B8;
              v244 = qword_1803BE028;
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
                v148 = sub_180371D80(a1, (_DWORD)v147, qword_1803CA5B8, qword_1803BEB08, qword_1803AB250);
              else
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
              sub_1801F3490(v149, qword_1803BEB08, v150);
              v151 = qword_1803BE018;
              v152 = qword_1803CA5B8;
              v245 = qword_1803BE028;
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
                v155 = sub_180371D80(a1, (_DWORD)v154, qword_1803CA5B8, qword_1803BDFD0, qword_1803AB250);
              else
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
              sub_1801F3490(v156, qword_1803BDFD0, v157);
              v158 = sub_1801F4010();
              if ( v158 == 0 )
              {
                sub_180373340(a1, &v248, qword_1803BEA10);
                v42 = v248;
                v43 = 27;
                goto LABEL_345;
              }
              v160 = (int *)sub_18037F0C0(v159, v158, qword_1803BEB18);
              if ( v160 == nullptr )
              {
                v42 = *(_QWORD *)(a1 + 96);
                v43 = 27;
                *(_QWORD *)(a1 + 96) = 0;
                goto LABEL_345;
              }
              v161 = qword_1803BEB20;
              v162 = sub_1801F4010();
              if ( v162 == 0 )
              {
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
              sub_1801F3490(v166, qword_1803BE9B0, v167);
              v168 = (int *)sub_180370EC0(a1);
              v169 = sub_180370EC0(a1);
              if ( *(_DWORD *)v169 != -1 )
                ++*(_DWORD *)v169;
              v170 = (__int64 *)sub_1800015A0(qword_1803CA5B8, qword_1803BED18);
              if ( v170 == nullptr || (v171 = *v170, *v170 == 0) )
              {
                v172 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803BED18);
                v171 = 0;
                if ( v172 != nullptr )
                  v171 = *v172;
              }
              PyDict_SetItem(v169, qword_1803BEB30, v171);
              PyDict_SetItem(v169, qword_1803BEB40, qword_1803BEB38);
              v173 = (_DWORD *)qword_1803BEB48;
              v174 = sub_18036FB80(a1, qword_1803BEB50);
              if ( *v173 != -1 )
                ++*v173;
              v175 = (int *)sub_1803609A0(
                              (int)sub_1801F4B80,
                              qword_1803BEB58,
                              qword_1803BEB60,
                              qword_1803BDE28,
                              (__int64)v173,
                              0,
                              v174,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEB58, v175);
              if ( *v175 >= 0 )
              {
                v214 = (*(_QWORD *)v175)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v175 + 1) + 48LL))(v175);
              }
              v176 = sub_18036FB80(a1, qword_1803BEB68);
              v177 = (int *)sub_1803609A0(
                              (int)sub_1801F5210,
                              qword_1803BE0F8,
                              qword_1803BEB70,
                              qword_1803BDD70,
                              0,
                              0,
                              v176,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE0F8, v177);
              if ( *v177 >= 0 )
              {
                v214 = (*(_QWORD *)v177)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v177 + 1) + 48LL))(v177);
              }
              v178 = (_DWORD *)qword_1803BEB78;
              v179 = sub_18036FB80(a1, qword_1803BEB80);
              if ( *v178 != -1 )
                ++*v178;
              v180 = (int *)sub_1803609A0(
                              (int)sub_1801F5880,
                              qword_1803BDF88,
                              qword_1803BEB88,
                              qword_1803BDDE8,
                              (__int64)v178,
                              0,
                              v179,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BDF88, v180);
              if ( *v180 >= 0 )
              {
                v214 = (*(_QWORD *)v180)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v180 + 1) + 48LL))(v180);
              }
              v181 = (_DWORD *)qword_1803BEB78;
              v182 = sub_18036FB80(a1, qword_1803BEB90);
              v183 = sub_18036FB80(a1, qword_1803BEB98);
              if ( *v181 != -1 )
                ++*v181;
              v184 = (int *)sub_1803609A0(
                              (int)sub_1801F5CB0,
                              qword_1803BEBA0,
                              qword_1803BEBA8,
                              qword_1803BDDE0,
                              (__int64)v181,
                              v182,
                              v183,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBA0, v184);
              if ( *v184 >= 0 )
              {
                v214 = (*(_QWORD *)v184)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v184 + 1) + 48LL))(v184);
              }
              v185 = (_DWORD *)qword_1803BEB78;
              v186 = sub_18036FB80(a1, qword_1803BEBB0);
              if ( *v185 != -1 )
                ++*v185;
              v187 = (int *)sub_1803609A0(
                              (int)sub_1801FD690,
                              qword_1803BEBB8,
                              qword_1803BEBC0,
                              qword_1803BDE20,
                              (__int64)v185,
                              0,
                              v186,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBB8, v187);
              if ( *v187 >= 0 )
              {
                v214 = (*(_QWORD *)v187)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v187 + 1) + 48LL))(v187);
              }
              v188 = (_DWORD *)qword_1803BEB78;
              v189 = sub_18036FB80(a1, qword_1803BEBC8);
              if ( *v188 != -1 )
                ++*v188;
              v190 = (int *)sub_1803609A0(
                              (int)sub_1802095A0,
                              qword_1803BEBD0,
                              qword_1803BEBD8,
                              qword_1803BDE30,
                              (__int64)v188,
                              0,
                              v189,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BEBD0, v190);
              if ( *v190 >= 0 )
              {
                v214 = (*(_QWORD *)v190)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v190 + 1) + 48LL))(v190);
              }
              v191 = sub_18036FB80(a1, qword_1803BEBE0);
              v192 = (int *)sub_1803609A0(
                              (int)sub_18020DF00,
                              qword_1803BE130,
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
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE130, v193);
              if ( *v193 >= 0 )
              {
                v214 = (*(_QWORD *)v193)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v193 + 1) + 48LL))(v193);
              }
              v194 = sub_18036FB80(a1, qword_1803BEBF0);
              v195 = (int *)sub_1803609A0(
                              (int)sub_18020E2D0,
                              qword_1803BDFB8,
                              qword_1803BEBF8,
                              qword_1803BDEA8,
                              0,
                              0,
                              v194,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BDFB8, v195);
              if ( *v195 >= 0 )
              {
                v214 = (*(_QWORD *)v195)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v195 + 1) + 48LL))(v195);
              }
              v196 = sub_18036FB80(a1, qword_1803BEC00);
              v197 = (int *)sub_1803609A0(
                              (int)sub_180213700,
                              qword_1803BE2D8,
                              qword_1803BEC08,
                              qword_1803BDDF8,
                              0,
                              0,
                              v196,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE2D8, v197);
              if ( *v197 >= 0 )
              {
                v214 = (*(_QWORD *)v197)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v197 + 1) + 48LL))(v197);
              }
              v198 = sub_18036FB80(a1, qword_1803BEC10);
              v199 = (int *)sub_1803609A0(
                              (int)sub_180213E10,
                              qword_1803BE1A0,
                              qword_1803BEC18,
                              qword_1803BDE60,
                              0,
                              0,
                              v198,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE1A0, v199);
              if ( *v199 >= 0 )
              {
                v214 = (*(_QWORD *)v199)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v199 + 1) + 48LL))(v199);
              }
              v200 = sub_18036FB80(a1, qword_1803BEC20);
              v201 = (int *)sub_1803609A0(
                              (int)sub_180215D90,
                              qword_1803BE1E0,
                              qword_1803BEC28,
                              qword_1803BDE88,
                              0,
                              0,
                              v200,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE1E0, v201);
              if ( *v201 >= 0 )
              {
                v214 = (*(_QWORD *)v201)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v201 + 1) + 48LL))(v201);
              }
              v202 = sub_18036FB80(a1, qword_1803BEC30);
              v203 = (int *)sub_1803609A0(
                              (int)sub_1802160A0,
                              qword_1803BE228,
                              qword_1803BEC38,
                              qword_1803BDE58,
                              0,
                              0,
                              v202,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE228, v203);
              if ( *v203 >= 0 )
              {
                v214 = (*(_QWORD *)v203)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v203 + 1) + 48LL))(v203);
              }
              v204 = sub_18036FB80(a1, qword_1803BEC40);
              v205 = (int *)sub_1803609A0(
                              (int)sub_180216D20,
                              qword_1803BE2A8,
                              qword_1803BEC48,
                              qword_1803BDE50,
                              0,
                              0,
                              v204,
                              qword_1803CA5C0,
                              0,
                              nullptr,
                              0);
              PyDict_SetItem(v169, *(_QWORD *)&qword_1803BE2A8, v205);
              if ( *v205 >= 0 )
              {
                v214 = (*(_QWORD *)v205)-- == 1;
                if ( v214 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v205 + 1) + 48LL))(v205);
              }
              v206 = *(_QWORD *)(a1 + 16);
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
