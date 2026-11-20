// Analyst annotations only; original export: source/ida/0x180082cb0.c
// qword comments derive module constant slot + original file range; not runtime values.
// write access to const memory has been detected, the output may be wrong!
// Original IDA line 2 | sub_180082CB0 -> NovelsApi.novels_send_message
int *__fastcall sub_180082CB0(__int64 a1, __int64 a2, int **a3)
{
  __int64 v3; // r12
  int *v4; // r13
  int *v6; // rbx
  bool v7; // zf
  __int64 v8; // rcx
  __int64 v9; // rax
  int *v10; // rbx
  unsigned int v11; // edi
  int *v12; // r14
  __int64 v13; // rcx
  int *v14; // rbx
  int *v15; // r14
  __int64 v16; // rcx
  __int64 v17; // r15
  __int64 v18; // rax
  int *v19; // rbx
  int *v20; // r14
  __int64 v21; // rcx
  int *v22; // rbx
  int *v23; // r14
  __int64 v24; // rcx
  __int64 v25; // r12
  __int64 v26; // rax
  int *v27; // rbx
  __int64 v28; // rax
  int v29; // eax
  bool v30; // cl
  int v31; // eax
  __int64 v32; // rdx
  int *v33; // r12
  __int64 v34; // rax
  int *v35; // rbx
  __int64 v36; // rcx
  __int64 v37; // r12
  __int64 v38; // rcx
  int *v39; // r15
  __int64 v40; // rax
  int *v41; // r14
  __int64 v42; // rbx
  int v43; // eax
  int *v44; // rcx
  __int64 v45; // rcx
  int *v46; // rbx
  __int64 v47; // rcx
  int *v48; // r14
  int *v49; // rbx
  __int64 v50; // rcx
  __int64 v51; // r15
  __int64 v52; // rax
  int *v53; // rbx
  __int64 v54; // rax
  __int64 v55; // rcx
  __int64 v56; // rax
  int *v57; // rbx
  int *v58; // r14
  int v59; // eax
  __int64 v60; // rcx
  __int64 v61; // rax
  int *v62; // rbx
  __int64 v63; // rax
  int *v64; // rbx
  int v65; // eax
  __int64 v66; // rcx
  __int64 v67; // rax
  int *v68; // rbx
  __int64 v69; // rcx
  __int64 v70; // rax
  int *v71; // rbx
  __int64 v72; // rcx
  int *v73; // r15
  __int64 v74; // rax
  int *v75; // rbx
  int *v76; // r12
  int v77; // ebx
  __int64 v78; // rax
  int *v79; // r12
  int v80; // eax
  __int64 v81; // rax
  __int64 v82; // rcx
  int *v83; // rbx
  __int64 v84; // rax
  __int64 v85; // rbx
  int *v86; // rdi
  int v87; // eax
  __int64 v88; // rax
  __int64 v89; // rcx
  int *v90; // rbx
  int *v91; // rbx
  __int64 v92; // rdx
  __int64 v93; // rdx
  _DWORD *v94; // rcx
  _DWORD *v95; // r14
  __int64 v96; // rax
  int *v97; // rdi
  int *v98; // rcx
  int *v99; // rcx
  __int64 v100; // rbx
  __int64 v101; // rax
  __int64 v102; // rbx
  unsigned int v103; // edi
  int v104; // eax
  __int64 v105; // r14
  __int64 v106; // rdi
  __int64 v107; // rdi
  __int64 v108; // rax
  int *v109; // rbx
  __int64 v110; // rax
  __int64 v111; // rcx
  __int64 v112; // r15
  int v113; // edx
  int v114; // r8d
  __int64 v115; // rax
  int *v116; // r14
  __int64 v117; // rax
  __int64 v118; // r12
  __int64 v119; // rax
  __int64 v120; // rbx
  __int64 v121; // rax
  int v122; // ebx
  __int64 v123; // rax
  __int64 v124; // r14
  __int64 v125; // rdi
  int *v126; // rcx
  int *v127; // rbx
  int *v128; // r15
  int *v129; // rdi
  int *v130; // r14
  int *v131; // rcx
  int *v132; // rsi
  char v133; // al
  __int64 v134; // rax
  int *v135; // rcx
  __int64 v136; // r14
  __int64 v137; // rax
  int *v138; // rdi
  int *v139; // rcx
  __int64 v140; // rdx
  __int64 v141; // r8
  __int64 v142; // r9
  __int64 v143; // rcx
  int *v144; // rcx
  __int64 v146; // rdi
  __int64 v147; // [rsp+20h] [rbp-F0h]
  __int64 v148; // [rsp+28h] [rbp-E8h]
  __int64 v149; // [rsp+90h] [rbp-80h]
  __int64 v150; // [rsp+90h] [rbp-80h]
  __int64 v151; // [rsp+98h] [rbp-78h]
  int *v152; // [rsp+98h] [rbp-78h]
  __int64 v153; // [rsp+98h] [rbp-78h]
  int *v154; // [rsp+98h] [rbp-78h]
  int *v155; // [rsp+A0h] [rbp-70h]
  int *v156; // [rsp+A8h] [rbp-68h]
  int *v157; // [rsp+B0h] [rbp-60h]
  int *v158; // [rsp+B8h] [rbp-58h]
  int *v159; // [rsp+C0h] [rbp-50h]
  int *v160; // [rsp+C8h] [rbp-48h]
  int *v161; // [rsp+D0h] [rbp-40h]
  int *v162; // [rsp+D8h] [rbp-38h]
  int *v163; // [rsp+E0h] [rbp-30h]
  int *v164; // [rsp+E8h] [rbp-28h]
  int *v165; // [rsp+F0h] [rbp-20h]
  int *v166; // [rsp+F8h] [rbp-18h]
  int *v167; // [rsp+100h] [rbp-10h]
  __int64 v168; // [rsp+108h] [rbp-8h]
  _DWORD *v169; // [rsp+108h] [rbp-8h]
  int *v170; // [rsp+110h] [rbp+0h]
  _QWORD v171[2]; // [rsp+118h] [rbp+8h] BYREF
  __int64 v172; // [rsp+128h] [rbp+18h]
  _QWORD v173[2]; // [rsp+130h] [rbp+20h] BYREF
  _QWORD v174[12]; // [rsp+140h] [rbp+30h] BYREF
  __int64 v175; // [rsp+1C0h] [rbp+B0h]
  __int64 v176; // [rsp+1C8h] [rbp+B8h] BYREF

  v3 = qword_1803AFC58;
  v4 = nullptr;
  v6 = a3[1];
  v167 = *a3;
  v155 = v6;
  v158 = nullptr;
  v159 = nullptr;
  v160 = nullptr;
  v161 = nullptr;
  v162 = nullptr;
  v163 = nullptr;
  v165 = nullptr;
  v164 = nullptr;
  v157 = nullptr;
  v166 = nullptr;
  v170 = nullptr;
  v175 = qword_1803AFC58;
  if ( qword_1803AFC58 == 0 )
  {
LABEL_8:
    v3 = sub_18036CD10(a1, *(_QWORD *)&qword_1803AFDF0, qword_1803CA828, 120);
    v175 = v3;
    qword_1803AFC58 = v3;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803AFC58 > 1 || *(_QWORD *)(qword_1803AFC58 + 16) != 0 )
  {
    if ( qword_1803AFC58 != 0 && *(int *)qword_1803AFC58 >= 0 )
    {
      v7 = (*(_QWORD *)qword_1803AFC58)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v3 + 8) + 48LL))(v3);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v3);
// Original IDA line 214 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v9 = sub_18037F0C0(v8, v6, qword_1803AFF88);
  v10 = (int *)v9;
  v11 = 555;
  if ( v9 == 0 )
    goto LABEL_279;
  *(_DWORD *)(v3 + 40) = 555;
// Original IDA line 220 | qword_1803B0210 -> novels.api.novels_api#88 bytes [3925897, 3925911]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925899, 3925910], "value": "sessionId"}, {"tag": "s", "fileRange": [3925910, 3925911], "value": ""}]}
  v12 = (int *)sub_180378210(a1, v9, qword_1803B0210);
  if ( *v10 >= 0 )
  {
    v7 = (*(_QWORD *)v10)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v10 + 1) + 48LL))(v10);
  }
  if ( v12 == nullptr )
    goto LABEL_279;
  v14 = (int *)sub_1803749A0(v12);
  if ( *v12 >= 0 )
  {
    v7 = (*(_QWORD *)v12)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v12 + 1) + 48LL))(v12);
  }
  if ( v14 == nullptr )
    goto LABEL_279;
// Original IDA line 238 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
  v15 = (int *)sub_18037F0C0(v13, v14, qword_1803B0140);
  if ( *v14 >= 0 )
  {
    v7 = (*(_QWORD *)v14)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v14 + 1) + 48LL))(v14);
  }
  if ( v15 == nullptr )
    goto LABEL_279;
  *(_DWORD *)(v3 + 40) = 555;
  v17 = sub_180376210(a1, v15);
  v176 = v17;
  if ( *v15 >= 0 )
  {
    v7 = (*(_QWORD *)v15)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v15 + 1) + 48LL))(v15);
  }
  if ( v17 == 0 )
    goto LABEL_279;
  v158 = (int *)v17;
// Original IDA line 259 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v18 = sub_18037F0C0(v16, v155, qword_1803AFF88);
  v19 = (int *)v18;
  v11 = 556;
  if ( v18 == 0 )
    goto LABEL_279;
  *(_DWORD *)(v3 + 40) = 556;
// Original IDA line 265 | qword_1803B05E0 -> novels.api.novels_api#210 bytes [3928184, 3928196]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928186, 3928195], "value": "message"}, {"tag": "s", "fileRange": [3928195, 3928196], "value": ""}]}
  v20 = (int *)sub_180378210(a1, v18, qword_1803B05E0);
  if ( *v19 >= 0 )
  {
    v7 = (*(_QWORD *)v19)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v19 + 1) + 48LL))(v19);
  }
  if ( v20 == nullptr )
    goto LABEL_279;
  v22 = (int *)sub_1803749A0(v20);
  if ( *v20 >= 0 )
  {
    v7 = (*(_QWORD *)v20)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v20 + 1) + 48LL))(v20);
  }
  if ( v22 == nullptr )
    goto LABEL_279;
// Original IDA line 283 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
  v23 = (int *)sub_18037F0C0(v21, v22, qword_1803B0140);
  if ( *v22 >= 0 )
  {
    v7 = (*(_QWORD *)v22)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v22 + 1) + 48LL))(v22);
  }
  if ( v23 == nullptr )
    goto LABEL_279;
  *(_DWORD *)(v3 + 40) = 556;
  v25 = sub_180376210(a1, v23);
  v172 = v25;
  if ( *v23 >= 0 )
  {
    v7 = (*(_QWORD *)v23)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v23 + 1) + 48LL))(v23);
  }
  if ( v25 == 0 )
    goto LABEL_278;
  v159 = (int *)v25;
// Original IDA line 304 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v26 = sub_18037F0C0(v24, v155, qword_1803AFF88);
  v27 = (int *)v26;
  v11 = 557;
  if ( v26 == 0 )
    goto LABEL_278;
  *(_DWORD *)(v175 + 40) = 557;
// Original IDA line 310 | qword_1803B0258 -> novels.api.novels_api#97 bytes [3926118, 3926127]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926120, 3926127], "value": "model"}]}
  v28 = sub_1803770D0(a1, v26, qword_1803B0258);
  v168 = v28;
  if ( *v27 >= 0 )
  {
    v7 = (*(_QWORD *)v27)-- == 1;
    if ( v7 )
    {
      (*(void (__fastcall **)(int *))(*((_QWORD *)v27 + 1) + 48LL))(v27);
      v28 = v168;
    }
  }
  if ( v28 == 0 )
    goto LABEL_278;
  v160 = (int *)v28;
  v29 = sub_180001510(v17);
  if ( v29 == -1 )
  {
    v11 = 558;
LABEL_278:
    v3 = v175;
    goto LABEL_279;
  }
  v30 = v29 == 0;
  if ( v29 != 0 )
  {
    v31 = sub_180001510(v25);
    if ( v31 == -1 )
    {
      v11 = 558;
      goto LABEL_278;
    }
    v30 = v31 == 0;
  }
  if ( v30 )
  {
// Original IDA line 345 | qword_1803B05E8 -> novels.api.novels_api#211 bytes [3928196, 3928253]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3928198, 3928207], "value": "success"}, {"tag": "a", "fileRange": [3928207, 3928216], "value": "message"}], "values": [{"tag": "F", "fileRange": [3928216, 3928217], "value": false}, {"tag... (truncated literal)
    v33 = (int *)sub_18036FB80(a1, qword_1803B05E8);
    goto LABEL_371;
  }
  v3 = v175;
  v11 = 560;
  *(_DWORD *)(v175 + 40) = 560;
// Original IDA line 351 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
  v34 = sub_18037E560(a1, v167, *(_QWORD *)&qword_1803B0058);
  v35 = (int *)v34;
  if ( v34 == 0 )
    goto LABEL_279;
  *(_DWORD *)(v175 + 40) = 560;
// Original IDA line 356 | qword_1803B0180 -> novels.api.novels_api#70 bytes [3925624, 3925638]: "load_session"
  v37 = sub_18037E6C0(a1, v34, qword_1803B0180, v17);
  if ( *v35 >= 0 )
  {
    v7 = (*(_QWORD *)v35)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v35 + 1) + 48LL))(v35);
  }
  if ( v37 == 0 )
    goto LABEL_278;
  v161 = (int *)v37;
// Original IDA line 366 | qword_1803B0220 -> novels.api.novels_api#90 bytes [3925956, 3925971]: "_clear_cancel"
  v39 = (int *)sub_18037F0C0(v36, v167, *(_QWORD *)&qword_1803B0220);
  if ( v39 == nullptr )
  {
    v11 = 561;
    goto LABEL_278;
  }
// Original IDA line 372 | qword_1803B0020 -> novels.api.novels_api#26 bytes [3924800, 3924808]: "bookId"
  v40 = sub_18037F0C0(v38, v37, qword_1803B0020);
  v41 = (int *)v40;
  if ( v40 == 0 )
  {
    v42 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v39 >= 0 )
    {
      v7 = (*(_QWORD *)v39)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v39 + 1) + 48LL))(v39);
    }
LABEL_67:
    v3 = v175;
    v11 = 561;
    goto LABEL_280;
  }
  v43 = sub_180001510(v40);
  if ( v43 == -1 )
  {
    v42 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v39 >= 0 )
    {
      v7 = (*(_QWORD *)v39)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v39 + 1) + 48LL))(v39);
    }
    if ( *v41 >= 0 )
    {
      v7 = (*(_QWORD *)v41)-- == 1;
      if ( v7 )
      {
        (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
        v3 = v175;
        v11 = 561;
        goto LABEL_280;
      }
    }
    goto LABEL_67;
  }
  if ( v43 != 1 )
  {
    if ( *v41 >= 0 )
    {
      v7 = (*(_QWORD *)v41)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
    }
    v44 = (int *)v176;
    if ( *(_DWORD *)v176 != -1 )
      ++*(_DWORD *)v176;
    v41 = v44;
  }
  v11 = 561;
  *(_DWORD *)(v175 + 40) = 561;
  v46 = (int *)sub_180376770(a1, v39, v41);
  if ( *v39 >= 0 )
  {
    v7 = (*(_QWORD *)v39)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v39 + 1) + 48LL))(v39);
  }
  if ( *v41 >= 0 )
  {
    v7 = (*(_QWORD *)v41)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
  }
  if ( v46 == nullptr )
    goto LABEL_278;
  if ( *v46 >= 0 )
  {
    v7 = (*(_QWORD *)v46)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v46 + 1) + 48LL))(v46);
  }
// Original IDA line 449 | qword_1803B0248 -> novels.api.novels_api#95 bytes [3926073, 3926089]: "_agent_service"
  v48 = (int *)sub_18037F0C0(v45, v167, *(_QWORD *)&qword_1803B0248);
  if ( v48 == nullptr )
  {
    v11 = 562;
    goto LABEL_278;
  }
// Original IDA line 455 | qword_1803B0020 -> novels.api.novels_api#26 bytes [3924800, 3924808]: "bookId"
  v49 = (int *)sub_18037F0C0(v47, v37, qword_1803B0020);
  if ( v49 == nullptr )
  {
    v42 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v48 >= 0 )
    {
      v7 = (*(_QWORD *)v48)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v48 + 1) + 48LL))(v48);
    }
    v3 = v175;
    v11 = 562;
    goto LABEL_280;
  }
  v3 = v175;
  v11 = 562;
  v173[0] = v176;
  v173[1] = v49;
  *(_DWORD *)(v175 + 40) = 562;
  v51 = sub_180377750(a1, v48, v173);
  if ( *v48 >= 0 )
  {
    v7 = (*(_QWORD *)v48)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v48 + 1) + 48LL))(v48);
  }
  if ( *v49 >= 0 )
  {
    v7 = (*(_QWORD *)v49)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v49 + 1) + 48LL))(v49);
  }
  if ( v51 == 0 )
    goto LABEL_279;
  v162 = (int *)v51;
// Original IDA line 491 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v52 = sub_18037F0C0(v50, v155, qword_1803AFF88);
  v53 = (int *)v52;
  v11 = 563;
  if ( v52 == 0 )
    goto LABEL_279;
  *(_DWORD *)(v175 + 40) = 563;
// Original IDA line 497 | qword_1803B05F0 -> novels.api.novels_api#212 bytes [3928253, 3928270]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928255, 3928270], "value": "contextLength"}]}
  v54 = sub_1803770D0(a1, v52, qword_1803B05F0);
  v151 = v54;
  if ( *v53 >= 0 )
  {
    v7 = (*(_QWORD *)v53)-- == 1;
    if ( v7 )
    {
      (*(void (__fastcall **)(int *))(*((_QWORD *)v53 + 1) + 48LL))(v53);
      v54 = v151;
    }
  }
  if ( v54 == 0 )
    goto LABEL_279;
  v163 = (int *)v54;
// Original IDA line 511 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v56 = sub_18037F0C0(v55, v155, qword_1803AFF88);
  v57 = (int *)v56;
  v11 = 564;
  if ( v56 == 0 )
    goto LABEL_279;
  *(_DWORD *)(v175 + 40) = 564;
// Original IDA line 517 | qword_1803B0228 -> novels.api.novels_api#91 bytes [3925971, 3925995]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925973, 3925995], "value": "codexReasoningEffort"}]}
  v58 = (int *)sub_1803770D0(a1, v56, qword_1803B0228);
  if ( *v57 >= 0 )
  {
    v7 = (*(_QWORD *)v57)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v57 + 1) + 48LL))(v57);
  }
  if ( v58 == nullptr )
    goto LABEL_279;
  v59 = sub_180001510(v58);
  if ( v59 != -1 )
  {
    if ( v59 != 1 )
    {
      if ( *v58 >= 0 )
      {
        v7 = (*(_QWORD *)v58)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
      }
// Original IDA line 537 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
      v61 = sub_18037F0C0(v60, v155, qword_1803AFF88);
      v62 = (int *)v61;
      if ( v61 == 0 )
        goto LABEL_279;
      *(_DWORD *)(v175 + 40) = 564;
// Original IDA line 542 | qword_1803B0230 -> novels.api.novels_api#92 bytes [3925995, 3926021]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925997, 3926021], "value": "codex_reasoning_effort"}]}
      v58 = (int *)sub_1803770D0(a1, v61, qword_1803B0230);
      if ( *v62 >= 0 )
      {
        v7 = (*(_QWORD *)v62)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v62 + 1) + 48LL))(v62);
      }
      if ( v58 == nullptr )
        goto LABEL_279;
    }
    v165 = v58;
// Original IDA line 553 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
    v63 = sub_18037F0C0(v60, v155, qword_1803AFF88);
    v64 = (int *)v63;
    v11 = 565;
    if ( v63 == 0 )
      goto LABEL_279;
    *(_DWORD *)(v175 + 40) = 565;
// Original IDA line 559 | qword_1803B0238 -> novels.api.novels_api#93 bytes [3926021, 3926046]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926023, 3926046], "value": "claudeReasoningEffort"}]}
    v58 = (int *)sub_1803770D0(a1, v63, qword_1803B0238);
    if ( *v64 >= 0 )
    {
      v7 = (*(_QWORD *)v64)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
    }
    if ( v58 == nullptr )
      goto LABEL_279;
    v65 = sub_180001510(v58);
    if ( v65 != -1 )
    {
      if ( v65 != 1 )
      {
        if ( *v58 >= 0 )
        {
          v7 = (*(_QWORD *)v58)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
        }
// Original IDA line 579 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
        v67 = sub_18037F0C0(v66, v155, qword_1803AFF88);
        v68 = (int *)v67;
        if ( v67 == 0 )
          goto LABEL_279;
        *(_DWORD *)(v175 + 40) = 565;
// Original IDA line 584 | qword_1803B0240 -> novels.api.novels_api#94 bytes [3926046, 3926073]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926048, 3926073], "value": "claude_reasoning_effort"}]}
        v58 = (int *)sub_1803770D0(a1, v67, qword_1803B0240);
        if ( *v68 >= 0 )
        {
          v7 = (*(_QWORD *)v68)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v68 + 1) + 48LL))(v68);
        }
        if ( v58 == nullptr )
          goto LABEL_279;
      }
      v164 = v58;
// Original IDA line 595 | qword_1803B05F8 -> novels.api.novels_api#213 bytes [3928270, 3928284]: "send_message"
      v58 = (int *)sub_18037F0C0(v66, v51, qword_1803B05F8);
      if ( v58 == nullptr )
      {
        v11 = 566;
LABEL_279:
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        goto LABEL_280;
      }
// Original IDA line 604 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
      v70 = sub_18037F0C0(v69, v155, qword_1803AFF88);
      v71 = (int *)v70;
      if ( v70 == 0 )
      {
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        if ( *v58 >= 0 )
        {
          v7 = (*(_QWORD *)v58)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
        }
        v11 = 570;
        goto LABEL_280;
      }
      v11 = 570;
      *(_DWORD *)(v175 + 40) = 570;
// Original IDA line 621 | qword_1803B0600 -> novels.api.novels_api#214 bytes [3928284, 3928304]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928286, 3928304], "value": "knowledgeBookIds"}]}
      v73 = (int *)sub_1803770D0(a1, v70, qword_1803B0600);
      if ( *v71 >= 0 )
      {
        v7 = (*(_QWORD *)v71)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v71 + 1) + 48LL))(v71);
      }
      if ( v73 != nullptr )
      {
// Original IDA line 630 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
        v74 = sub_18037F0C0(v72, v155, qword_1803AFF88);
        v75 = (int *)v74;
        if ( v74 == 0 )
        {
          v42 = *(_QWORD *)(a1 + 96);
          *(_QWORD *)(a1 + 96) = 0;
          if ( *v58 >= 0 )
          {
            v7 = (*(_QWORD *)v58)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
          }
          if ( *v73 >= 0 )
          {
            v7 = (*(_QWORD *)v73)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
          }
          v11 = 571;
          goto LABEL_280;
        }
        v11 = 571;
        *(_DWORD *)(v175 + 40) = 571;
// Original IDA line 653 | qword_1803B0608 -> novels.api.novels_api#215 bytes [3928304, 3928321]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928306, 3928320], "value": "thinkingMode"}, {"tag": "t", "fileRange": [3928320, 3928321], "value": true}]}
        v76 = (int *)sub_180378210(a1, v74, qword_1803B0608);
        if ( *v75 >= 0 )
        {
          v7 = (*(_QWORD *)v75)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v75 + 1) + 48LL))(v75);
        }
        if ( v76 != nullptr )
        {
          v77 = sub_180001510(v76);
          if ( *v76 >= 0 )
          {
            v7 = (*(_QWORD *)v76)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v76 + 1) + 48LL))(v76);
          }
          if ( v77 != -1 )
          {
            v78 = Py_FalseStruct;
            v79 = (int *)Py_NoneStruct;
            if ( v77 != 0 )
              v78 = Py_TrueStruct;
            v149 = v78;
            if ( v151 == Py_NoneStruct )
            {
              if ( Py_NoneStruct != -1 )
                ++Py_NoneStruct;
            }
            else
            {
              v79 = (int *)PyNumber_Long();
              if ( v79 == nullptr )
              {
                v42 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v58 >= 0 )
                {
                  v7 = (*(_QWORD *)v58)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
                }
                if ( *v73 >= 0 )
                {
                  v7 = (*(_QWORD *)v73)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
                }
                v3 = v175;
                v11 = 572;
                goto LABEL_280;
              }
            }
            v80 = sub_180001510(v165);
            if ( v80 == -1 )
              goto LABEL_192;
            if ( v80 != 0 )
            {
              v81 = sub_1803749A0(v165);
              v83 = (int *)v81;
              if ( v81 == 0 )
                goto LABEL_192;
// Original IDA line 714 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
              v84 = sub_18037F0C0(v82, v81, qword_1803B0140);
              v152 = (int *)v84;
              if ( *v83 >= 0 )
              {
                v7 = (*(_QWORD *)v83)-- == 1;
                if ( v7 )
                {
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v83 + 1) + 48LL))(v83);
                  v84 = (__int64)v152;
                }
              }
              if ( v84 == 0 )
              {
LABEL_192:
                v42 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v58 >= 0 )
                {
                  v7 = (*(_QWORD *)v58)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
                }
                if ( *v73 >= 0 )
                {
                  v7 = (*(_QWORD *)v73)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
                }
                if ( *v79 >= 0 )
                {
                  v7 = (*(_QWORD *)v79)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v79 + 1) + 48LL))(v79);
                }
                v3 = v175;
                v11 = 573;
                goto LABEL_280;
              }
              v11 = 573;
              *(_DWORD *)(v175 + 40) = 573;
              v85 = sub_180376210(a1, v84);
              v156 = (int *)v85;
              if ( *v152 >= 0 )
              {
                v7 = (*(_QWORD *)v152)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v152 + 1) + 48LL))(v152);
              }
              if ( v85 == 0 )
              {
                v42 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v58 >= 0 )
                {
                  v7 = (*(_QWORD *)v58)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
                }
                if ( *v73 >= 0 )
                {
                  v7 = (*(_QWORD *)v73)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
                }
                if ( *v79 >= 0 )
                {
                  v7 = (*(_QWORD *)v79)-- == 1;
                  if ( v7 )
                  {
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v79 + 1) + 48LL))(v79);
                    v3 = v175;
                    goto LABEL_280;
                  }
                }
                goto LABEL_173;
              }
              v86 = (int *)v85;
            }
            else
            {
              v86 = (int *)Py_NoneStruct;
              v156 = (int *)Py_NoneStruct;
              if ( Py_NoneStruct != -1 )
                ++Py_NoneStruct;
            }
            v87 = sub_180001510(v164);
            if ( v87 != -1 )
            {
              if ( v87 != 0 )
              {
                v88 = sub_1803749A0(v164);
                v90 = (int *)v88;
                if ( v88 == 0 )
                  goto LABEL_225;
// Original IDA line 808 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
                v153 = sub_18037F0C0(v89, v88, qword_1803B0140);
                if ( *v90 >= 0 )
                {
                  v7 = (*(_QWORD *)v90)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v90 + 1) + 48LL))(v90);
                }
                v91 = (int *)v153;
                if ( v153 == 0 )
                  goto LABEL_225;
                v11 = 574;
                *(_DWORD *)(v175 + 40) = 574;
                v92 = sub_180376210(a1, v153);
                v154 = (int *)v92;
                if ( *v91 >= 0 )
                {
                  v7 = (*(_QWORD *)v91)-- == 1;
                  if ( v7 )
                  {
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v91 + 1) + 48LL))(v91);
                    v92 = (__int64)v154;
                  }
                }
                if ( v92 == 0 )
                {
                  v42 = *(_QWORD *)(a1 + 96);
                  *(_QWORD *)(a1 + 96) = 0;
                  if ( *v58 >= 0 )
                  {
                    v7 = (*(_QWORD *)v58)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
                  }
                  if ( *v73 >= 0 )
                  {
                    v7 = (*(_QWORD *)v73)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
                  }
                  if ( *v79 >= 0 )
                  {
                    v7 = (*(_QWORD *)v79)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v79 + 1) + 48LL))(v79);
                  }
                  if ( *v156 >= 0 )
                  {
                    v7 = (*(_QWORD *)v156)-- == 1;
                    if ( v7 )
                    {
                      (*(void (**)(void))(*((_QWORD *)v156 + 1) + 48LL))();
                      v3 = v175;
                      goto LABEL_280;
                    }
                  }
                  goto LABEL_173;
                }
              }
              else
              {
                v92 = Py_NoneStruct;
                v154 = (int *)Py_NoneStruct;
                if ( Py_NoneStruct != -1 )
                  ++Py_NoneStruct;
              }
              v174[5] = v92;
              v11 = 566;
              v174[1] = v73;
              *(_DWORD *)(v175 + 40) = 566;
              v171[0] = v176;
              v171[1] = v172;
              v174[0] = v168;
              v174[2] = v149;
              v174[3] = v79;
              v174[4] = v156;
// Original IDA line 883 | qword_1803B0610 -> novels.api.novels_api#216 bytes [3928321, 3928430]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928323, 3928330], "value": "model"}, {"tag": "a", "fileRange": [3928330, 3928350], "value": "knowledge_book_ids"}, {"tag": "a", "fileRange": [3928350, 3928365], "value": "thinking_mode... (truncated literal)
              v150 = sub_18037D650(a1, v58, v171, v174, qword_1803B0610);
              if ( *v58 >= 0 )
              {
                v7 = (*(_QWORD *)v58)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
              }
              if ( *v73 >= 0 )
              {
                v7 = (*(_QWORD *)v73)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
              }
              if ( *v79 >= 0 )
              {
                v7 = (*(_QWORD *)v79)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v79 + 1) + 48LL))(v79);
              }
              if ( *v156 >= 0 )
              {
                v7 = (*(_QWORD *)v156)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v156 + 1) + 48LL))(v156);
              }
              if ( *v154 >= 0 )
              {
                v7 = (*(_QWORD *)v154)-- == 1;
                if ( v7 )
                  (*(void (**)(void))(*((_QWORD *)v154 + 1) + 48LL))();
              }
              if ( v150 != 0 )
              {
                v4 = (int *)v150;
                sub_180001D80(a1, v93);
// Original IDA line 918 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
                v146 = qword_1803B0120;
                v33 = (int *)PyDict_NewPresized(2);
                PyDict_SetItem(v33, v146, Py_TrueStruct);
// Original IDA line 921 | qword_1803AFF90 -> novels.api.novels_api#8 bytes [3924535, 3924541]: "data"
                PyDict_SetItem(v33, qword_1803AFF90, v150);
                v127 = nullptr;
                v129 = v164;
                v128 = nullptr;
                v130 = v165;
                v132 = nullptr;
                v131 = v158;
LABEL_458:
                if ( *v131 >= 0 )
                {
                  v7 = (*(_QWORD *)v131)-- == 1;
                  if ( v7 )
                    (*(void (**)(void))(*((_QWORD *)v131 + 1) + 48LL))();
                }
                goto LABEL_461;
              }
              goto LABEL_278;
            }
LABEL_225:
            v42 = *(_QWORD *)(a1 + 96);
            *(_QWORD *)(a1 + 96) = 0;
            if ( *v58 >= 0 )
            {
              v7 = (*(_QWORD *)v58)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
            }
            if ( *v73 >= 0 )
            {
              v7 = (*(_QWORD *)v73)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
            }
            if ( *v79 >= 0 )
            {
              v7 = (*(_QWORD *)v79)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v79 + 1) + 48LL))(v79);
            }
            if ( *v86 >= 0 )
            {
              v7 = (*(_QWORD *)v86)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v86 + 1) + 48LL))(v86);
            }
            v3 = v175;
            v11 = 574;
LABEL_280:
            v94 = *(_DWORD **)(a1 + 248);
            v176 = 0;
            v169 = v94;
            if ( v94 != nullptr && *v94 != -1 )
              ++*v94;
            v95 = *(_DWORD **)(v42 + 40);
            if ( v95 == nullptr )
            {
              v96 = sub_18038E2A0(v3, v11);
              v97 = (int *)v96;
              v98 = *(int **)(v42 + 40);
              if ( v96 == Py_NoneStruct )
              {
                v96 = 0;
                goto LABEL_293;
              }
LABEL_290:
              if ( v97 != nullptr && *(_DWORD *)v96 != -1 )
                ++*(_DWORD *)v96;
              goto LABEL_293;
            }
            v96 = sub_18038E2A0(v3, v11);
            v97 = (int *)v96;
            *(_QWORD *)(v96 + 16) = v95;
            if ( *v95 != -1 )
              ++*v95;
            v98 = *(int **)(v42 + 40);
            if ( v96 != Py_NoneStruct )
              goto LABEL_290;
            v96 = 0;
LABEL_293:
            *(_QWORD *)(v42 + 40) = v96;
            if ( v98 != nullptr && *v98 >= 0 )
            {
              v7 = (*(_QWORD *)v98)-- == 1;
              if ( v7 )
                (*(void (**)(void))(*((_QWORD *)v98 + 1) + 48LL))();
            }
            if ( v97 != nullptr && *v97 >= 0 )
            {
              v7 = (*(_QWORD *)v97)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v97 + 1) + 48LL))(v97);
            }
            v99 = *(int **)(a1 + 248);
            *(_QWORD *)(a1 + 248) = v42;
            if ( v99 != nullptr && *v99 >= 0 )
            {
              v7 = (*(_QWORD *)v99)-- == 1;
              if ( v7 )
                (*(void (**)(void))(*((_QWORD *)v99 + 1) + 48LL))();
            }
            v100 = *(_QWORD *)(*(_QWORD *)(a1 + 248) + 8LL);
            v101 = sub_180066230(a1);
            if ( v101 == 0 )
            {
// Original IDA line 1025 | qword_1803B0268 -> novels.api.novels_api#99 bytes [3926135, 3926160]: "OperationCancelledError"
              sub_180373340(a1, &v176, qword_1803B0268);
              v102 = v176;
              v103 = 577;
              goto LABEL_378;
            }
            v104 = sub_18000CE90(a1, v100, v101);
            if ( v104 == -1 )
            {
              v102 = *(_QWORD *)(a1 + 96);
              v103 = 577;
              *(_QWORD *)(a1 + 96) = 0;
              goto LABEL_378;
            }
            v105 = *(_QWORD *)(a1 + 248);
            if ( v104 != 0 )
            {
              if ( *(_DWORD *)v105 != -1 )
                ++*(_DWORD *)v105;
// Original IDA line 1043 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
              v106 = qword_1803B0120;
              v33 = (int *)PyDict_NewPresized(3);
              PyDict_SetItem(v33, v106, Py_FalseStruct);
// Original IDA line 1046 | qword_1803B0270 -> novels.api.novels_api#100 bytes [3926160, 3926171]: "cancelled"
              PyDict_SetItem(v33, qword_1803B0270, Py_TrueStruct);
// Original IDA line 1047 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
              v107 = qword_1803B0128;
              v108 = sub_1803749A0(v105);
              v109 = (int *)v108;
              if ( v108 == 0 )
              {
                v102 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v33 >= 0 )
                {
                  v7 = (*(_QWORD *)v33)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v33 + 1) + 48LL))(v33);
                }
                if ( v105 != 0 && *(int *)v105 >= 0 )
                {
                  v7 = (*(_QWORD *)v105)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(__int64))(*(_QWORD *)(v105 + 8) + 48LL))(v105);
                }
                v103 = 578;
                goto LABEL_378;
              }
              PyDict_SetItem(v33, v107, v108);
              if ( *v109 >= 0 )
              {
                v7 = (*(_QWORD *)v109)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v109 + 1) + 48LL))(v109);
              }
              if ( v105 == 0 )
                goto LABEL_367;
              if ( *(int *)v105 < 0 )
                goto LABEL_367;
              v7 = (*(_QWORD *)v105)-- == 1;
              if ( !v7 )
                goto LABEL_367;
              v110 = *(_QWORD *)(v105 + 8);
              v111 = v105;
              goto LABEL_366;
            }
            if ( (unsigned int)sub_18000CE90(a1, *(_QWORD *)(v105 + 8), PyExc_Exception) == 0 )
            {
              v133 = sub_180373290(a1, &v176);
              v102 = v176;
              v103 = 554;
              if ( v133 != 0 )
                v103 = 0;
              v134 = *(_QWORD *)(v176 + 40);
              if ( v134 != 0 && *(_QWORD *)(v134 + 24) == v3 )
                *(_DWORD *)(v3 + 40) = *(_DWORD *)(v134 + 36);
LABEL_378:
              v135 = *(int **)(a1 + 248);
              *(_QWORD *)(a1 + 248) = v169;
              if ( v135 != nullptr && *v135 >= 0 )
              {
                v7 = (*(_QWORD *)v135)-- == 1;
                if ( v7 )
                  (*(void (**)(void))(*((_QWORD *)v135 + 1) + 48LL))();
              }
              v136 = *(_QWORD *)(v102 + 40);
              if ( v136 != 0 )
              {
                if ( *(_QWORD *)(v136 + 24) == v175 )
                {
LABEL_401:
                  sub_18036D0A0(v175, "ooooooooooooooo", v167);
                  v143 = qword_1803AFC58;
                  if ( v175 == qword_1803AFC58 )
                  {
                    if ( *(int *)qword_1803AFC58 >= 0 )
                    {
                      v7 = (*(_QWORD *)qword_1803AFC58)-- == 1;
                      if ( v7 )
                        (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *, int *, int *, int *, int *, int *, int *, _QWORD))(*(_QWORD *)(v143 + 8) + 48LL))(
                          v143,
                          *(_QWORD *)(*(_QWORD *)(v143 + 8) + 48LL),
                          v141,
                          v142,
                          v158,
                          v159,
                          v160,
                          v161,
                          v162,
                          v163,
                          v165,
                          v164,
                          0);
                    }
                    qword_1803AFC58 = 0;
                  }
                  sub_180001D80(a1, v140);
                  if ( v158 != nullptr && *v158 >= 0 )
                  {
                    v7 = (*(_QWORD *)v158)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v158 + 1) + 48LL))();
                  }
                  if ( v159 != nullptr && *v159 >= 0 )
                  {
                    v7 = (*(_QWORD *)v159)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v159 + 1) + 48LL))();
                  }
                  if ( v160 != nullptr && *v160 >= 0 )
                  {
                    v7 = (*(_QWORD *)v160)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v160 + 1) + 48LL))();
                  }
                  if ( v161 != nullptr && *v161 >= 0 )
                  {
                    v7 = (*(_QWORD *)v161)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v161 + 1) + 48LL))();
                  }
                  if ( v162 != nullptr && *v162 >= 0 )
                  {
                    v7 = (*(_QWORD *)v162)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v162 + 1) + 48LL))();
                  }
                  if ( v163 != nullptr && *v163 >= 0 )
                  {
                    v7 = (*(_QWORD *)v163)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v163 + 1) + 48LL))();
                  }
                  if ( v165 != nullptr && *v165 >= 0 )
                  {
                    v7 = (*(_QWORD *)v165)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v165 + 1) + 48LL))(v165);
                  }
                  if ( v164 != nullptr && *v164 >= 0 )
                  {
                    v7 = (*(_QWORD *)v164)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v164 + 1) + 48LL))(v164);
                  }
                  if ( v157 != nullptr && *v157 >= 0 )
                  {
                    v7 = (*(_QWORD *)v157)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v157 + 1) + 48LL))();
                  }
                  if ( v166 != nullptr && *v166 >= 0 )
                  {
                    v7 = (*(_QWORD *)v166)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v166 + 1) + 48LL))();
                  }
                  if ( *v167 >= 0 )
                  {
                    v7 = (*(_QWORD *)v167)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v167 + 1) + 48LL))();
                  }
                  if ( *v155 >= 0 )
                  {
                    v7 = (*(_QWORD *)v155)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v155 + 1) + 48LL))(v155);
                  }
                  v144 = *(int **)(a1 + 96);
                  *(_QWORD *)(a1 + 96) = v102;
                  if ( v144 != nullptr && *v144 >= 0 )
                  {
                    v7 = (*(_QWORD *)v144)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v144 + 1) + 48LL))();
                  }
                  return nullptr;
                }
                v137 = sub_18038E2A0(v175, v103);
                v138 = (int *)v137;
                *(_QWORD *)(v137 + 16) = v136;
                if ( *(_DWORD *)v136 != -1 )
                  ++*(_DWORD *)v136;
                v139 = *(int **)(v102 + 40);
                if ( v137 == Py_NoneStruct )
                {
                  v137 = 0;
LABEL_393:
                  *(_QWORD *)(v102 + 40) = v137;
                  if ( v139 != nullptr && *v139 >= 0 )
                  {
                    v7 = (*(_QWORD *)v139)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v139 + 1) + 48LL))();
                  }
                  if ( v138 != nullptr && *v138 >= 0 )
                  {
                    v7 = (*(_QWORD *)v138)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v138 + 1) + 48LL))(v138);
                  }
                  goto LABEL_401;
                }
              }
              else
              {
                v137 = sub_18038E2A0(v175, v103);
                v138 = (int *)v137;
                v139 = *(int **)(v102 + 40);
                if ( v137 == Py_NoneStruct )
                {
                  v137 = 0;
                  goto LABEL_393;
                }
              }
              if ( v138 != nullptr && *(_DWORD *)v137 != -1 )
                ++*(_DWORD *)v137;
              goto LABEL_393;
            }
            v112 = *(_QWORD *)(a1 + 248);
            if ( *(_DWORD *)v112 != -1 )
              ++*(_DWORD *)v112;
            v103 = 580;
// Original IDA line 1265 | qword_1803B0278 -> novels.api.novels_api#101 bytes [3926171, 3926188]: "utils.llm_retry"
            v113 = qword_1803B0278;
            v114 = qword_1803CA830;
// Original IDA line 1267 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
            v148 = qword_1803B0288;
// Original IDA line 1268 | qword_1803B0280 -> novels.api.novels_api#102 bytes [3926188, 3926234]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926190, 3926210], "value": "friendly_llm_error"}, {"tag": "a", "fileRange": [3926210, 3926234], "value": "is_transient_llm_error"}]}
            v147 = qword_1803B0280;
            *(_DWORD *)(v3 + 40) = 580;
            v115 = sub_1803717E0(a1, v113, v114, Py_NoneStruct, v147, v148);
            v116 = (int *)v115;
            if ( v115 != 0 )
            {
              if ( *(_QWORD *)(v115 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 1275 | qword_1803B0290 -> novels.api.novels_api#104 bytes [3926236, 3926256]: "friendly_llm_error"
                v117 = sub_180371D80(a1, (_DWORD)v116, qword_1803CA830, qword_1803B0290, qword_1803AB250);
              else
// Original IDA line 1277 | qword_1803B0290 -> novels.api.novels_api#104 bytes [3926236, 3926256]: "friendly_llm_error"
                v117 = sub_180371890(a1, v116, qword_1803B0290);
              v118 = v117;
              if ( v117 == 0
                || ((v157 = (int *)v117, *((_QWORD *)v116 + 1) == PyModule_Type)
                 || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 1282 | qword_1803B0298 -> novels.api.novels_api#105 bytes [3926256, 3926280]: "is_transient_llm_error"
                  ? (v119 = sub_180371D80(a1, (_DWORD)v116, qword_1803CA830, qword_1803B0298, qword_1803AB250))
// Original IDA line 1283 | qword_1803B0298 -> novels.api.novels_api#105 bytes [3926256, 3926280]: "is_transient_llm_error"
                  : (v119 = sub_180371890(a1, v116, qword_1803B0298)),
                    v120 = v119,
                    v119 == 0) )
              {
LABEL_342:
                v102 = *(_QWORD *)(a1 + 96);
                *(_QWORD *)(a1 + 96) = 0;
                if ( *v116 >= 0 )
                {
                  v7 = (*(_QWORD *)v116)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v116 + 1) + 48LL))(v116);
                }
                goto LABEL_358;
              }
              v166 = (int *)v119;
              if ( *v116 >= 0 )
              {
                v7 = (*(_QWORD *)v116)-- == 1;
                if ( v7 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v116 + 1) + 48LL))(v116);
              }
              v103 = 581;
              *(_DWORD *)(v175 + 40) = 581;
              v121 = sub_180376770(a1, v120, v112);
              v116 = (int *)v121;
              if ( v121 != 0 )
              {
                v122 = sub_180001510(v121);
                if ( v122 == -1 )
                  goto LABEL_342;
                if ( *v116 >= 0 )
                {
                  v7 = (*(_QWORD *)v116)-- == 1;
                  if ( v7 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v116 + 1) + 48LL))(v116);
                }
                if ( v122 != 0 )
                {
                  *(_DWORD *)(v175 + 40) = 581;
                  v123 = sub_180376770(a1, v118, v112);
                }
                else
                {
                  v123 = sub_1803749A0(v112);
                }
                v170 = (int *)v123;
                v124 = v123;
                if ( v123 != 0 )
                {
// Original IDA line 1333 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
                  v125 = qword_1803B0120;
                  v33 = (int *)PyDict_NewPresized(2);
                  PyDict_SetItem(v33, v125, Py_FalseStruct);
// Original IDA line 1336 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
                  PyDict_SetItem(v33, qword_1803B0128, v124);
                  if ( v112 != 0 && *(int *)v112 >= 0 )
                  {
                    v7 = (*(_QWORD *)v112)-- == 1;
                    if ( v7 )
                    {
                      v110 = *(_QWORD *)(v112 + 8);
                      v111 = v112;
LABEL_366:
                      (*(void (__fastcall **)(__int64))(v110 + 48))(v111);
                    }
                  }
LABEL_367:
                  v126 = *(int **)(a1 + 248);
                  *(_QWORD *)(a1 + 248) = v169;
                  if ( v126 != nullptr && *v126 >= 0 )
                  {
                    v7 = (*(_QWORD *)v126)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v126 + 1) + 48LL))();
                  }
LABEL_371:
                  v127 = v170;
                  v128 = v166;
                  v129 = v164;
                  v130 = v165;
                  sub_180001D80(a1, v32);
                  v131 = v158;
                  v132 = v157;
                  if ( v158 != nullptr )
                    goto LABEL_458;
LABEL_461:
                  if ( v159 != nullptr && *v159 >= 0 )
                  {
                    v7 = (*(_QWORD *)v159)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v159 + 1) + 48LL))();
                  }
                  if ( v160 != nullptr && *v160 >= 0 )
                  {
                    v7 = (*(_QWORD *)v160)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v160 + 1) + 48LL))();
                  }
                  if ( v161 != nullptr && *v161 >= 0 )
                  {
                    v7 = (*(_QWORD *)v161)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v161 + 1) + 48LL))();
                  }
                  if ( v162 != nullptr && *v162 >= 0 )
                  {
                    v7 = (*(_QWORD *)v162)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v162 + 1) + 48LL))();
                  }
                  if ( v163 != nullptr && *v163 >= 0 )
                  {
                    v7 = (*(_QWORD *)v163)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v163 + 1) + 48LL))();
                  }
                  if ( v130 != nullptr && *v130 >= 0 )
                  {
                    v7 = (*(_QWORD *)v130)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v130 + 1) + 48LL))(v130);
                  }
                  if ( v129 != nullptr && *v129 >= 0 )
                  {
                    v7 = (*(_QWORD *)v129)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v129 + 1) + 48LL))(v129);
                  }
                  if ( v4 != nullptr && *v4 >= 0 )
                  {
                    v7 = (*(_QWORD *)v4)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v4 + 1) + 48LL))(v4);
                  }
                  if ( v132 != nullptr && *v132 >= 0 )
                  {
                    v7 = (*(_QWORD *)v132)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v132 + 1) + 48LL))(v132);
                  }
                  if ( v128 != nullptr && *v128 >= 0 )
                  {
                    v7 = (*(_QWORD *)v128)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v128 + 1) + 48LL))(v128);
                  }
                  if ( v127 != nullptr && *v127 >= 0 )
                  {
                    v7 = (*(_QWORD *)v127)-- == 1;
                    if ( v7 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v127 + 1) + 48LL))(v127);
                  }
                  if ( *v167 >= 0 )
                  {
                    v7 = (*(_QWORD *)v167)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v167 + 1) + 48LL))();
                  }
                  if ( *v155 >= 0 )
                  {
                    v7 = (*(_QWORD *)v155)-- == 1;
                    if ( v7 )
                      (*(void (**)(void))(*((_QWORD *)v155 + 1) + 48LL))();
                  }
                  return v33;
                }
              }
            }
            v102 = *(_QWORD *)(a1 + 96);
            *(_QWORD *)(a1 + 96) = 0;
LABEL_358:
            if ( v112 != 0 && *(int *)v112 >= 0 )
            {
              v7 = (*(_QWORD *)v112)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(__int64))(*(_QWORD *)(v112 + 8) + 48LL))(v112);
            }
            goto LABEL_378;
          }
        }
        v42 = *(_QWORD *)(a1 + 96);
        *(_QWORD *)(a1 + 96) = 0;
        if ( *v58 >= 0 )
        {
          v7 = (*(_QWORD *)v58)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
        }
        if ( *v73 >= 0 )
        {
          v7 = (*(_QWORD *)v73)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
        }
LABEL_173:
        v3 = v175;
        goto LABEL_280;
      }
    }
  }
  v42 = *(_QWORD *)(a1 + 96);
  *(_QWORD *)(a1 + 96) = 0;
  if ( *v58 >= 0 )
  {
    v7 = (*(_QWORD *)v58)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
  }
  goto LABEL_280;
}
