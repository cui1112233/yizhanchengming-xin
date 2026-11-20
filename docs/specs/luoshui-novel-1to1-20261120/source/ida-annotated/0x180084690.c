// Analyst annotations only; original export: source/ida/0x180084690.c
// qword comments derive module constant slot + original file range; not runtime values.
// write access to const memory has been detected, the output may be wrong!
// Original IDA line 2 | sub_180084690 -> NovelsApi.novels_confirm_action
__int64 __fastcall sub_180084690(_QWORD *a1, __int64 a2, int **a3)
{
  __int64 v3; // r12
  int *v4; // r14
  int *v6; // rdi
  bool v7; // zf
  __int64 v8; // rcx
  __int64 v9; // rax
  int *v10; // rbx
  unsigned int v11; // edi
  __int64 v12; // rbx
  int *v13; // r15
  __int64 v14; // rcx
  int *v15; // rbx
  int *v16; // r15
  __int64 v17; // rcx
  int *v18; // r13
  __int64 v19; // rax
  int *v20; // rbx
  __int64 v21; // rax
  int v22; // eax
  __int64 v23; // r12
  __int64 v24; // rdx
  int *v25; // rbx
  int *v26; // rdi
  int *v27; // rsi
  int *v28; // rdx
  __int64 v29; // rax
  int *v30; // rbx
  __int64 v31; // rcx
  __int64 v32; // r15
  __int64 v33; // rcx
  int *v34; // r12
  __int64 v35; // rax
  int *v36; // r15
  int v37; // eax
  __int64 v38; // rcx
  int *v39; // rbx
  __int64 v40; // rax
  int *v41; // rbx
  int *v42; // r15
  int v43; // eax
  __int64 v44; // rcx
  __int64 v45; // rax
  int *v46; // rbx
  __int64 v47; // rax
  int *v48; // rbx
  int v49; // eax
  __int64 v50; // rcx
  __int64 v51; // rax
  int *v52; // rbx
  int v53; // eax
  __int64 v54; // rcx
  int v55; // eax
  __int64 v56; // rax
  __int64 v57; // rcx
  int *v58; // rbx
  int *v59; // r12
  __int64 v60; // r15
  int v61; // eax
  __int64 v62; // rcx
  __int64 v63; // rax
  __int64 v64; // rcx
  __int64 v65; // r12
  __int64 v66; // rcx
  int *v67; // r15
  __int64 v68; // rax
  int *v69; // rbx
  __int64 v70; // rcx
  __int64 v71; // r12
  __int64 v72; // rax
  int *v73; // rbx
  __int64 v74; // rcx
  int *v75; // r13
  __int64 v76; // rax
  int *v77; // rbx
  int *v78; // r15
  __int64 v79; // rdx
  int v80; // ebx
  __int64 v81; // rdi
  __int64 v82; // rcx
  int *v83; // r15
  int *v84; // rbx
  _DWORD *v85; // rcx
  int *v86; // r13
  _DWORD *v87; // r15
  __int64 v88; // rax
  int *v89; // rdi
  int *v90; // rcx
  int *v91; // rcx
  __int64 v92; // rbx
  __int64 v93; // rax
  int *v94; // rbx
  unsigned int v95; // edi
  int v96; // eax
  __int64 v97; // r15
  __int64 v98; // rdi
  __int64 v99; // rdi
  __int64 v100; // rax
  int *v101; // rbx
  __int64 v102; // rdx
  __int64 v103; // r15
  int v104; // edx
  int v105; // r8d
  __int64 v106; // rax
  int *v107; // r12
  __int64 v108; // rax
  __int64 v109; // r13
  __int64 v110; // rcx
  __int64 v111; // rax
  __int64 v112; // rbx
  __int64 v113; // rax
  int v114; // ebx
  __int64 v115; // rax
  __int64 v116; // r13
  __int64 v117; // rdi
  int *v118; // rcx
  int *v119; // r15
  char v120; // al
  __int64 v121; // rax
  int *v122; // rcx
  __int64 v123; // r15
  __int64 v124; // rax
  int *v125; // rdi
  int *v126; // rcx
  int *v127; // r15
  int *v128; // r13
  __int64 v129; // r8
  __int64 v130; // r9
  __int64 v131; // rcx
  __int64 v132; // rcx
  __int64 v133; // rdi
  int *v134; // rcx
  int *v135; // rcx
  int *v136; // rcx
  int *v137; // rcx
  int *v138; // rcx
  int *v139; // rcx
  int *v140; // rcx
  int *v141; // rcx
  int *v142; // rcx
  int *v143; // rcx
  int *v144; // rcx
  int *v146; // r13
  int *v147; // rcx
  int *v148; // rcx
  int *v149; // rcx
  int *v150; // rcx
  int *v151; // rcx
  int *v152; // rcx
  int *v153; // rcx
  __int64 v154; // [rsp+20h] [rbp-B9h]
  int *v155; // [rsp+20h] [rbp-B9h]
  __int64 v156; // [rsp+28h] [rbp-B1h]
  int *v157; // [rsp+28h] [rbp-B1h]
  int *v158; // [rsp+30h] [rbp-A9h]
  int *v159; // [rsp+40h] [rbp-99h]
  int *v160; // [rsp+48h] [rbp-91h]
  int *v161; // [rsp+50h] [rbp-89h]
  int *v162; // [rsp+58h] [rbp-81h]
  int *v163; // [rsp+80h] [rbp-59h] BYREF
  __int64 v164; // [rsp+88h] [rbp-51h] BYREF
  _QWORD v165[2]; // [rsp+90h] [rbp-49h] BYREF
  int *v166; // [rsp+A0h] [rbp-39h]
  int *v167; // [rsp+A8h] [rbp-31h]
  int *v168; // [rsp+B0h] [rbp-29h]
  int *v169; // [rsp+B8h] [rbp-21h]
  int *v170; // [rsp+C0h] [rbp-19h]
  int *v171; // [rsp+C8h] [rbp-11h]
  int *v172; // [rsp+D0h] [rbp-9h]
  int *v173; // [rsp+D8h] [rbp-1h]
  int *v174; // [rsp+E0h] [rbp+7h]
  int *v175; // [rsp+E8h] [rbp+Fh]
  int *v176; // [rsp+F0h] [rbp+17h]
  _DWORD *v177; // [rsp+F8h] [rbp+1Fh]
  __int64 v178; // [rsp+150h] [rbp+77h]
  int *v179; // [rsp+158h] [rbp+7Fh]

  v3 = qword_1803AFE70;
  v4 = nullptr;
  v6 = a3[1];
  v176 = *a3;
  v166 = v6;
  v171 = nullptr;
  v172 = nullptr;
  v173 = nullptr;
  v179 = nullptr;
  v168 = nullptr;
  v167 = nullptr;
  v174 = nullptr;
  v175 = nullptr;
  v170 = nullptr;
  v169 = nullptr;
  v178 = qword_1803AFE70;
  if ( qword_1803AFE70 == 0 )
  {
LABEL_8:
    v3 = sub_18036CD10(a1, *(_QWORD *)&qword_1803AFD30, qword_1803CA828, 112);
    v178 = v3;
    qword_1803AFE70 = v3;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803AFE70 > 1 || *(_QWORD *)(qword_1803AFE70 + 16) != 0 )
  {
    if ( qword_1803AFE70 != 0 && *(int *)qword_1803AFE70 >= 0 )
    {
      v7 = (*(_QWORD *)qword_1803AFE70)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v3 + 8) + 48LL))(v3);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v3);
// Original IDA line 216 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v9 = sub_18037F0C0(v8, v6, qword_1803AFF88);
  v10 = (int *)v9;
  v11 = 586;
  if ( v9 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  *(_DWORD *)(v3 + 40) = 586;
// Original IDA line 226 | qword_1803B0210 -> novels.api.novels_api#88 bytes [3925897, 3925911]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925899, 3925910], "value": "sessionId"}, {"tag": "s", "fileRange": [3925910, 3925911], "value": ""}]}
  v13 = (int *)sub_180378210(a1, v9, qword_1803B0210);
  if ( *v10 >= 0 )
  {
    v7 = (*(_QWORD *)v10)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v10 + 1) + 48LL))(v10);
  }
  if ( v13 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  v15 = (int *)sub_1803749A0(v13);
  if ( *v13 >= 0 )
  {
    v7 = (*(_QWORD *)v13)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v13 + 1) + 48LL))(v13);
  }
  if ( v15 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
// Original IDA line 252 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
  v16 = (int *)sub_18037F0C0(v14, v15, qword_1803B0140);
  if ( *v15 >= 0 )
  {
    v7 = (*(_QWORD *)v15)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v15 + 1) + 48LL))(v15);
  }
  if ( v16 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  *(_DWORD *)(v3 + 40) = 586;
  v18 = (int *)sub_180376210(a1, v16);
  if ( *v16 >= 0 )
  {
    v7 = (*(_QWORD *)v16)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v16 + 1) + 48LL))(v16);
  }
  if ( v18 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  v171 = v18;
// Original IDA line 280 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
  v19 = sub_18037F0C0(v17, v166, qword_1803AFF88);
  v20 = (int *)v19;
  v11 = 587;
  if ( v19 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  *(_DWORD *)(v3 + 40) = 587;
// Original IDA line 290 | qword_1803B0258 -> novels.api.novels_api#97 bytes [3926118, 3926127]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926120, 3926127], "value": "model"}]}
  v21 = sub_1803770D0(a1, v19, qword_1803B0258);
  v177 = (_DWORD *)v21;
  if ( *v20 >= 0 )
  {
    v7 = (*(_QWORD *)v20)-- == 1;
    if ( v7 )
    {
      (*(void (__fastcall **)(int *))(*((_QWORD *)v20 + 1) + 48LL))(v20);
      v21 = (__int64)v177;
    }
  }
  if ( v21 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_200;
  }
  v172 = (int *)v21;
  v22 = sub_180001510(v18);
  if ( v22 == -1 )
  {
    v12 = a1[12];
    v11 = 588;
    a1[12] = 0;
    goto LABEL_200;
  }
  if ( v22 != 0 )
  {
    v28 = v176;
    v11 = 590;
    *(_DWORD *)(v3 + 40) = 590;
// Original IDA line 321 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
    v29 = sub_18037E560(a1, v28, *(_QWORD *)&qword_1803B0058);
    v30 = (int *)v29;
    if ( v29 == 0 )
    {
      v12 = a1[12];
      a1[12] = 0;
      goto LABEL_200;
    }
    *(_DWORD *)(v3 + 40) = 590;
// Original IDA line 330 | qword_1803B0180 -> novels.api.novels_api#70 bytes [3925624, 3925638]: "load_session"
    v32 = sub_18037E6C0(a1, v29, qword_1803B0180, v18);
    v164 = v32;
    if ( *v30 >= 0 )
    {
      v7 = (*(_QWORD *)v30)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v30 + 1) + 48LL))(v30);
    }
    if ( v32 == 0 )
    {
      v12 = a1[12];
      a1[12] = 0;
      goto LABEL_200;
    }
    v173 = (int *)v32;
// Original IDA line 345 | qword_1803B0220 -> novels.api.novels_api#90 bytes [3925956, 3925971]: "_clear_cancel"
    v34 = (int *)sub_18037F0C0(v31, v176, *(_QWORD *)&qword_1803B0220);
    if ( v34 == nullptr )
    {
      v12 = a1[12];
      v11 = 591;
      a1[12] = 0;
LABEL_199:
      v3 = v178;
      goto LABEL_200;
    }
// Original IDA line 355 | qword_1803B0020 -> novels.api.novels_api#26 bytes [3924800, 3924808]: "bookId"
    v35 = sub_18037F0C0(v33, v32, qword_1803B0020);
    v36 = (int *)v35;
    if ( v35 == 0 )
    {
      v12 = a1[12];
      a1[12] = 0;
      if ( *v34 >= 0 )
      {
        v7 = (*(_QWORD *)v34)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v34 + 1) + 48LL))(v34);
      }
LABEL_55:
      v11 = 591;
      goto LABEL_199;
    }
    v37 = sub_180001510(v35);
    if ( v37 == -1 )
    {
      v12 = a1[12];
      a1[12] = 0;
      if ( *v34 >= 0 )
      {
        v7 = (*(_QWORD *)v34)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v34 + 1) + 48LL))(v34);
      }
      if ( *v36 >= 0 )
      {
        v7 = (*(_QWORD *)v36)-- == 1;
        if ( v7 )
        {
          (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
          v11 = 591;
          goto LABEL_199;
        }
      }
      goto LABEL_55;
    }
    if ( v37 != 1 )
    {
      if ( *v36 >= 0 )
      {
        v7 = (*(_QWORD *)v36)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
      }
      if ( *v18 != -1 )
        ++*v18;
      v36 = v18;
    }
    v11 = 591;
    *(_DWORD *)(v178 + 40) = 591;
    v39 = (int *)sub_180376770(a1, v34, v36);
    if ( *v34 >= 0 )
    {
      v7 = (*(_QWORD *)v34)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v34 + 1) + 48LL))(v34);
    }
    if ( *v36 >= 0 )
    {
      v7 = (*(_QWORD *)v36)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
    }
    if ( v39 == nullptr )
    {
      v12 = a1[12];
      a1[12] = 0;
      goto LABEL_199;
    }
    if ( *v39 >= 0 )
    {
      v7 = (*(_QWORD *)v39)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v39 + 1) + 48LL))(v39);
    }
// Original IDA line 433 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
    v40 = sub_18037F0C0(v38, v166, qword_1803AFF88);
    v3 = v178;
    v41 = (int *)v40;
    v11 = 592;
    if ( v40 == 0 )
    {
      v12 = a1[12];
      a1[12] = 0;
      goto LABEL_200;
    }
    *(_DWORD *)(v178 + 40) = 592;
// Original IDA line 444 | qword_1803B0228 -> novels.api.novels_api#91 bytes [3925971, 3925995]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925973, 3925995], "value": "codexReasoningEffort"}]}
    v42 = (int *)sub_1803770D0(a1, v40, qword_1803B0228);
    if ( *v41 >= 0 )
    {
      v7 = (*(_QWORD *)v41)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
    }
    if ( v42 == nullptr )
    {
      v12 = a1[12];
      a1[12] = 0;
      goto LABEL_200;
    }
    v43 = sub_180001510(v42);
    if ( v43 != -1 )
    {
      if ( v43 != 1 )
      {
        if ( *v42 >= 0 )
        {
          v7 = (*(_QWORD *)v42)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v42 + 1) + 48LL))(v42);
        }
// Original IDA line 468 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
        v45 = sub_18037F0C0(v44, v166, qword_1803AFF88);
        v46 = (int *)v45;
        if ( v45 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        *(_DWORD *)(v178 + 40) = 592;
// Original IDA line 477 | qword_1803B0230 -> novels.api.novels_api#92 bytes [3925995, 3926021]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925997, 3926021], "value": "codex_reasoning_effort"}]}
        v42 = (int *)sub_1803770D0(a1, v45, qword_1803B0230);
        if ( *v46 >= 0 )
        {
          v7 = (*(_QWORD *)v46)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v46 + 1) + 48LL))(v46);
        }
        if ( v42 == nullptr )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
      }
      v179 = v42;
// Original IDA line 492 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
      v47 = sub_18037F0C0(v44, v166, qword_1803AFF88);
      v48 = (int *)v47;
      v11 = 593;
      if ( v47 == 0 )
      {
        v12 = a1[12];
        a1[12] = 0;
        goto LABEL_200;
      }
      *(_DWORD *)(v178 + 40) = 593;
// Original IDA line 502 | qword_1803B0238 -> novels.api.novels_api#93 bytes [3926021, 3926046]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926023, 3926046], "value": "claudeReasoningEffort"}]}
      v42 = (int *)sub_1803770D0(a1, v47, qword_1803B0238);
      if ( *v48 >= 0 )
      {
        v7 = (*(_QWORD *)v48)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v48 + 1) + 48LL))(v48);
      }
      if ( v42 == nullptr )
      {
        v12 = a1[12];
        a1[12] = 0;
        goto LABEL_200;
      }
      v49 = sub_180001510(v42);
      if ( v49 != -1 )
      {
        if ( v49 != 1 )
        {
          if ( *v42 >= 0 )
          {
            v7 = (*(_QWORD *)v42)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v42 + 1) + 48LL))(v42);
          }
// Original IDA line 526 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
          v51 = sub_18037F0C0(v50, v166, qword_1803AFF88);
          v52 = (int *)v51;
          if ( v51 == 0 )
          {
            v12 = a1[12];
            a1[12] = 0;
            goto LABEL_200;
          }
          *(_DWORD *)(v178 + 40) = 593;
// Original IDA line 535 | qword_1803B0240 -> novels.api.novels_api#94 bytes [3926046, 3926073]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926048, 3926073], "value": "claude_reasoning_effort"}]}
          v42 = (int *)sub_1803770D0(a1, v51, qword_1803B0240);
          if ( *v52 >= 0 )
          {
            v7 = (*(_QWORD *)v52)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v52 + 1) + 48LL))(v52);
          }
          if ( v42 == nullptr )
          {
            v12 = a1[12];
            a1[12] = 0;
            goto LABEL_200;
          }
        }
        v168 = v42;
        v53 = sub_180001510(v179);
        if ( v53 == -1 )
        {
          v12 = a1[12];
          v11 = 594;
          a1[12] = 0;
          goto LABEL_200;
        }
        if ( v53 == 1 )
        {
          v54 = (__int64)v179;
        }
        else
        {
          v55 = sub_180001510(v42);
          if ( v55 == -1 )
          {
            v12 = a1[12];
            v11 = 594;
            a1[12] = 0;
            goto LABEL_200;
          }
// Original IDA line 572 | qword_1803AFFD8 -> novels.api.novels_api#17 bytes [3924645, 3924646]: ""
          v54 = qword_1803AFFD8;
          if ( v55 == 1 )
            v54 = (__int64)v42;
        }
        v56 = sub_1803749A0(v54);
        v58 = (int *)v56;
        if ( v56 == 0 )
        {
          v12 = a1[12];
          v11 = 594;
          a1[12] = 0;
          goto LABEL_200;
        }
// Original IDA line 585 | qword_1803B0140 -> novels.api.novels_api#62 bytes [3925506, 3925513]: "strip"
        v59 = (int *)sub_18037F0C0(v57, v56, qword_1803B0140);
        if ( *v58 >= 0 )
        {
          v7 = (*(_QWORD *)v58)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
        }
        v11 = 594;
        if ( v59 == nullptr )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_199;
        }
        *(_DWORD *)(v178 + 40) = 594;
        v60 = sub_180376210(a1, v59);
        if ( *v59 >= 0 )
        {
          v7 = (*(_QWORD *)v59)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v59 + 1) + 48LL))(v59);
        }
        if ( v60 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_199;
        }
        v61 = sub_180001510(v60);
        if ( v61 == -1 )
        {
          v12 = a1[12];
          a1[12] = 0;
          if ( *(int *)v60 < 0 )
            goto LABEL_199;
          v7 = (*(_QWORD *)v60)-- == 1;
          if ( !v7 )
            goto LABEL_199;
          v63 = *(_QWORD *)(v60 + 8);
          v64 = v60;
          goto LABEL_198;
        }
        if ( v61 == 1 )
        {
          v65 = v60;
          v167 = (int *)v60;
        }
        else
        {
          if ( *(int *)v60 >= 0 )
          {
            v7 = (*(_QWORD *)v60)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(__int64))(*(_QWORD *)(v60 + 8) + 48LL))(v60);
          }
          v65 = Py_NoneStruct;
          v167 = (int *)Py_NoneStruct;
          if ( Py_NoneStruct != -1 )
            ++Py_NoneStruct;
        }
// Original IDA line 645 | qword_1803B0248 -> novels.api.novels_api#95 bytes [3926073, 3926089]: "_agent_service"
        v67 = (int *)sub_18037F0C0(v62, v176, *(_QWORD *)&qword_1803B0248);
        if ( v67 == nullptr )
        {
          v12 = a1[12];
          v11 = 595;
          a1[12] = 0;
          goto LABEL_199;
        }
// Original IDA line 653 | qword_1803B0020 -> novels.api.novels_api#26 bytes [3924800, 3924808]: "bookId"
        v68 = sub_18037F0C0(v66, v164, qword_1803B0020);
        v69 = (int *)v68;
        if ( v68 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          if ( *v67 >= 0 )
          {
            v7 = (*(_QWORD *)v67)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v67 + 1) + 48LL))(v67);
          }
          v11 = 595;
          goto LABEL_199;
        }
        v11 = 595;
        v165[0] = v18;
        v165[1] = v68;
        v164 = v65;
        *(_DWORD *)(v178 + 40) = 595;
// Original IDA line 673 | qword_1803B00D8 -> novels.api.novels_api#49 bytes [3925263, 3925281]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925265, 3925281], "value": "channel_effort"}]}
        v71 = sub_18037D650(a1, v67, v165, &v164, qword_1803B00D8);
        if ( *v67 >= 0 )
        {
          v7 = (*(_QWORD *)v67)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v67 + 1) + 48LL))(v67);
        }
        if ( *v69 >= 0 )
        {
          v7 = (*(_QWORD *)v69)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v69 + 1) + 48LL))(v69);
        }
        if ( v71 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_199;
        }
        v174 = (int *)v71;
// Original IDA line 693 | qword_1803B0618 -> novels.api.novels_api#217 bytes [3928430, 3928446]: "confirm_action"
        v72 = sub_18037F0C0(v70, v71, qword_1803B0618);
        v3 = v178;
        v73 = (int *)v72;
        v11 = 596;
        if ( v72 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        v164 = (__int64)v177;
        *(_DWORD *)(v178 + 40) = 596;
        v163 = v18;
// Original IDA line 706 | qword_1803B0258 -> novels.api.novels_api#97 bytes [3926118, 3926127]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926120, 3926127], "value": "model"}]}
        v75 = (int *)sub_18037D030(a1, v72, &v163, &v164, qword_1803B0258);
        if ( *v73 >= 0 )
        {
          v7 = (*(_QWORD *)v73)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
        }
        if ( v75 == nullptr )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        v175 = v75;
// Original IDA line 720 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
        v76 = sub_18037F0C0(v74, v75, qword_1803AFF88);
        v77 = (int *)v76;
        v11 = 597;
        if ( v76 == 0 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        *(_DWORD *)(v178 + 40) = 597;
// Original IDA line 730 | qword_1803B0088 -> novels.api.novels_api#39 bytes [3925047, 3925058]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3925049, 3925058], "value": "success"}]}
        v78 = (int *)sub_1803770D0(a1, v76, qword_1803B0088);
        if ( *v77 >= 0 )
        {
          v7 = (*(_QWORD *)v77)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v77 + 1) + 48LL))(v77);
        }
        if ( v78 == nullptr )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        v80 = sub_180001510(v78);
        if ( *v78 >= 0 )
        {
          v7 = (*(_QWORD *)v78)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v78 + 1) + 48LL))(v78);
        }
        if ( v80 == -1 )
        {
          v12 = a1[12];
          a1[12] = 0;
          goto LABEL_200;
        }
        if ( v80 != 0 )
        {
// Original IDA line 758 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
          v81 = qword_1803B0120;
          v23 = PyDict_NewPresized(2);
          PyDict_SetItem(v23, v81, Py_TrueStruct);
// Original IDA line 761 | qword_1803AFF90 -> novels.api.novels_api#8 bytes [3924535, 3924541]: "data"
          v165[0] = qword_1803AFF90;
// Original IDA line 762 | qword_1803AFF88 -> novels.api.novels_api#7 bytes [3924530, 3924535]: "get"
          v83 = (int *)sub_18037F0C0(v82, v75, qword_1803AFF88);
          v11 = 599;
          if ( v83 == nullptr )
            goto LABEL_195;
          *(_DWORD *)(v178 + 40) = 599;
// Original IDA line 767 | qword_1803B0260 -> novels.api.novels_api#98 bytes [3926127, 3926135]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926129, 3926135], "value": "data"}]}
          v84 = (int *)sub_1803770D0(a1, v83, qword_1803B0260);
          if ( *v83 >= 0 )
          {
            v7 = (*(_QWORD *)v83)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v83 + 1) + 48LL))(v83);
          }
          if ( v84 == nullptr )
          {
LABEL_195:
            v12 = a1[12];
            a1[12] = 0;
            if ( *(int *)v23 < 0 )
              goto LABEL_199;
            v7 = (*(_QWORD *)v23)-- == 1;
            if ( !v7 )
              goto LABEL_199;
            v63 = *(_QWORD *)(v23 + 8);
            v64 = v23;
LABEL_198:
            (*(void (__fastcall **)(__int64))(v63 + 48))(v64);
            goto LABEL_199;
          }
          PyDict_SetItem(v23, v165[0], v84);
          if ( *v84 >= 0 )
          {
            v7 = (*(_QWORD *)v84)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v84 + 1) + 48LL))(v84);
          }
        }
        else
        {
          v23 = (__int64)v75;
          if ( *v75 != -1 )
            ++*v75;
        }
        sub_180001D80(a1, v79);
        v27 = v168;
        v26 = v179;
        v25 = v167;
        goto LABEL_388;
      }
    }
    v12 = a1[12];
    a1[12] = 0;
    if ( *v42 >= 0 )
    {
      v7 = (*(_QWORD *)v42)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v42 + 1) + 48LL))(v42);
    }
LABEL_200:
    v85 = (_DWORD *)a1[31];
    v86 = v171;
    v165[0] = v171;
    v163 = nullptr;
    v177 = v85;
    if ( v85 != nullptr && *v85 != -1 )
      ++*v85;
    v87 = *(_DWORD **)(v12 + 40);
    if ( v87 != nullptr )
    {
      v88 = sub_18038E2A0(v3, v11);
      v89 = (int *)v88;
      *(_QWORD *)(v88 + 16) = v87;
      if ( *v87 != -1 )
        ++*v87;
      v90 = *(int **)(v12 + 40);
      if ( v88 == Py_NoneStruct )
      {
        v88 = 0;
        goto LABEL_213;
      }
    }
    else
    {
      v88 = sub_18038E2A0(v3, v11);
      v89 = (int *)v88;
      v90 = *(int **)(v12 + 40);
      if ( v88 == Py_NoneStruct )
      {
        v88 = 0;
        goto LABEL_213;
      }
    }
    if ( v89 != nullptr && *(_DWORD *)v88 != -1 )
      ++*(_DWORD *)v88;
LABEL_213:
    *(_QWORD *)(v12 + 40) = v88;
    if ( v90 != nullptr && *v90 >= 0 )
    {
      v7 = (*(_QWORD *)v90)-- == 1;
      if ( v7 )
        (*(void (**)(void))(*((_QWORD *)v90 + 1) + 48LL))();
    }
    if ( v89 != nullptr && *v89 >= 0 )
    {
      v7 = (*(_QWORD *)v89)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v89 + 1) + 48LL))(v89);
    }
    v91 = (int *)a1[31];
    a1[31] = v12;
    if ( v91 != nullptr && *v91 >= 0 )
    {
      v7 = (*(_QWORD *)v91)-- == 1;
      if ( v7 )
        (*(void (**)(void))(*((_QWORD *)v91 + 1) + 48LL))();
    }
    v92 = *(_QWORD *)(a1[31] + 8LL);
    v93 = sub_180066230(a1);
    if ( v93 == 0 )
    {
// Original IDA line 881 | qword_1803B0268 -> novels.api.novels_api#99 bytes [3926135, 3926160]: "OperationCancelledError"
      sub_180373340(a1, &v163, qword_1803B0268);
      v94 = v163;
      v95 = 600;
      goto LABEL_298;
    }
    v96 = sub_18000CE90(a1, v92, v93);
    if ( v96 == -1 )
    {
      v94 = (int *)a1[12];
      v95 = 600;
      a1[12] = 0;
      goto LABEL_298;
    }
    v97 = a1[31];
    if ( v96 != 0 )
    {
      if ( *(_DWORD *)v97 != -1 )
        ++*(_DWORD *)v97;
// Original IDA line 899 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
      v98 = qword_1803B0120;
      v23 = PyDict_NewPresized(3);
      PyDict_SetItem(v23, v98, Py_FalseStruct);
// Original IDA line 902 | qword_1803B0270 -> novels.api.novels_api#100 bytes [3926160, 3926171]: "cancelled"
      PyDict_SetItem(v23, qword_1803B0270, Py_TrueStruct);
// Original IDA line 903 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
      v99 = qword_1803B0128;
      v100 = sub_1803749A0(v97);
      v101 = (int *)v100;
      if ( v100 == 0 )
      {
        v94 = (int *)a1[12];
        a1[12] = 0;
        if ( *(int *)v23 >= 0 )
        {
          v7 = (*(_QWORD *)v23)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(__int64))(*(_QWORD *)(v23 + 8) + 48LL))(v23);
        }
        if ( v97 != 0 && *(int *)v97 >= 0 )
        {
          v7 = (*(_QWORD *)v97)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(__int64))(*(_QWORD *)(v97 + 8) + 48LL))(v97);
        }
        v95 = 601;
        goto LABEL_298;
      }
      PyDict_SetItem(v23, v99, v100);
      if ( *v101 >= 0 )
      {
        v7 = (*(_QWORD *)v101)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v101 + 1) + 48LL))(v101);
      }
      if ( v97 != 0 && *(int *)v97 >= 0 )
      {
        v7 = (*(_QWORD *)v97)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(__int64))(*(_QWORD *)(v97 + 8) + 48LL))(v97);
      }
LABEL_287:
      v118 = (int *)a1[31];
      v119 = v169;
      a1[31] = v177;
      if ( v118 != nullptr && *v118 >= 0 )
      {
        v7 = (*(_QWORD *)v118)-- == 1;
        if ( v7 )
          (*(void (**)(void))(*((_QWORD *)v118 + 1) + 48LL))();
      }
      sub_180001D80(a1, v102);
      v26 = v179;
      v27 = v168;
      v25 = v167;
      v171 = v86;
      if ( v86 == nullptr )
        goto LABEL_392;
      goto LABEL_389;
    }
    if ( (unsigned int)sub_18000CE90(a1, *(_QWORD *)(v97 + 8), PyExc_Exception) == 0 )
    {
      v120 = sub_180373290(a1, &v163);
      v94 = v163;
      v95 = 585;
      if ( v120 != 0 )
        v95 = 0;
      v121 = *((_QWORD *)v163 + 5);
      if ( v121 != 0 && *(_QWORD *)(v121 + 24) == v3 )
        *(_DWORD *)(v3 + 40) = *(_DWORD *)(v121 + 36);
LABEL_298:
      v122 = (int *)a1[31];
      a1[31] = v177;
      if ( v122 != nullptr && *v122 >= 0 )
      {
        v7 = (*(_QWORD *)v122)-- == 1;
        if ( v7 )
          (*(void (**)(void))(*((_QWORD *)v122 + 1) + 48LL))();
      }
      v123 = *((_QWORD *)v94 + 5);
      if ( v123 != 0 )
      {
        if ( *(_QWORD *)(v123 + 24) == v178 )
        {
LABEL_321:
          v127 = v167;
          v128 = v168;
          v162 = v175;
          v161 = v174;
          v160 = v167;
          v159 = v168;
          v158 = v173;
          v157 = v172;
          v155 = v171;
          sub_18036D0A0(v178, "oooooooooooooo", v176);
          v131 = qword_1803AFE70;
          if ( v178 == qword_1803AFE70 )
          {
            if ( *(int *)qword_1803AFE70 >= 0 )
            {
              v7 = (*(_QWORD *)qword_1803AFE70)-- == 1;
              if ( v7 )
                (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *, int *, int *, int *, int *, int *, int *))(*(_QWORD *)(v131 + 8) + 48LL))(
                  v131,
                  *(_QWORD *)(*(_QWORD *)(v131 + 8) + 48LL),
                  v129,
                  v130,
                  v155,
                  v157,
                  v158,
                  v179,
                  v159,
                  v160,
                  v161,
                  v162);
            }
            qword_1803AFE70 = 0;
          }
          v132 = a1[7];
          v133 = *(_QWORD *)(*(_QWORD *)v132 + 48LL);
          *(_QWORD *)v132 = *(_QWORD *)(*(_QWORD *)v132 + 8LL);
          *(_DWORD *)(v133 + 64) = -1;
          v134 = *(int **)(v133 + 16);
          if ( v134 != nullptr )
          {
            *(_QWORD *)(v133 + 16) = 0;
            if ( *v134 >= 0 )
            {
              v7 = (*(_QWORD *)v134)-- == 1;
              if ( v7 )
                (*(void (**)(void))(*((_QWORD *)v134 + 1) + 48LL))();
            }
          }
          if ( *(int *)v133 >= 0 )
          {
            v7 = (*(_QWORD *)v133)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(__int64))(*(_QWORD *)(v133 + 8) + 48LL))(v133);
          }
          *(_QWORD *)(v133 + 80) = 0;
          v135 = v171;
          if ( v171 != nullptr && *v171 >= 0 )
          {
            v7 = (*(_QWORD *)v171)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v135 + 1) + 48LL))();
          }
          v136 = v172;
          if ( v172 != nullptr && *v172 >= 0 )
          {
            v7 = (*(_QWORD *)v172)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v136 + 1) + 48LL))();
          }
          v137 = v173;
          if ( v173 != nullptr && *v173 >= 0 )
          {
            v7 = (*(_QWORD *)v173)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v137 + 1) + 48LL))();
          }
          if ( v179 != nullptr && *v179 >= 0 )
          {
            v7 = (*(_QWORD *)v179)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v179 + 1) + 48LL))(v179);
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
          v138 = v174;
          if ( v174 != nullptr && *v174 >= 0 )
          {
            v7 = (*(_QWORD *)v174)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v138 + 1) + 48LL))();
          }
          v139 = v175;
          if ( v175 != nullptr && *v175 >= 0 )
          {
            v7 = (*(_QWORD *)v175)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v139 + 1) + 48LL))();
          }
          v140 = v170;
          if ( v170 != nullptr && *v170 >= 0 )
          {
            v7 = (*(_QWORD *)v170)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v140 + 1) + 48LL))();
          }
          v141 = v169;
          if ( v169 != nullptr && *v169 >= 0 )
          {
            v7 = (*(_QWORD *)v169)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v141 + 1) + 48LL))();
          }
          v142 = v176;
          if ( *v176 >= 0 )
          {
            v7 = (*(_QWORD *)v176)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v142 + 1) + 48LL))();
          }
          v143 = v166;
          if ( *v166 >= 0 )
          {
            v7 = (*(_QWORD *)v166)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v143 + 1) + 48LL))();
          }
          v144 = (int *)a1[12];
          a1[12] = v94;
          if ( v144 != nullptr && *v144 >= 0 )
          {
            v7 = (*(_QWORD *)v144)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v144 + 1) + 48LL))();
          }
          return 0;
        }
        v124 = sub_18038E2A0(v178, v95);
        v125 = (int *)v124;
        *(_QWORD *)(v124 + 16) = v123;
        if ( *(_DWORD *)v123 != -1 )
          ++*(_DWORD *)v123;
        v126 = *((int **)v94 + 5);
        if ( v124 == Py_NoneStruct )
        {
          v124 = 0;
LABEL_313:
          *((_QWORD *)v94 + 5) = v124;
          if ( v126 != nullptr && *v126 >= 0 )
          {
            v7 = (*(_QWORD *)v126)-- == 1;
            if ( v7 )
              (*(void (**)(void))(*((_QWORD *)v126 + 1) + 48LL))();
          }
          if ( v125 != nullptr && *v125 >= 0 )
          {
            v7 = (*(_QWORD *)v125)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v125 + 1) + 48LL))(v125);
          }
          goto LABEL_321;
        }
      }
      else
      {
        v124 = sub_18038E2A0(v178, v95);
        v125 = (int *)v124;
        v126 = *((int **)v94 + 5);
        if ( v124 == Py_NoneStruct )
        {
          v124 = 0;
          goto LABEL_313;
        }
      }
      if ( v125 != nullptr && *(_DWORD *)v124 != -1 )
        ++*(_DWORD *)v124;
      goto LABEL_313;
    }
    v103 = a1[31];
    if ( *(_DWORD *)v103 != -1 )
      ++*(_DWORD *)v103;
    v95 = 603;
// Original IDA line 1173 | qword_1803B0278 -> novels.api.novels_api#101 bytes [3926171, 3926188]: "utils.llm_retry"
    v104 = qword_1803B0278;
    v105 = qword_1803CA830;
// Original IDA line 1175 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
    v156 = qword_1803B0288;
// Original IDA line 1176 | qword_1803B0280 -> novels.api.novels_api#102 bytes [3926188, 3926234]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3926190, 3926210], "value": "friendly_llm_error"}, {"tag": "a", "fileRange": [3926210, 3926234], "value": "is_transient_llm_error"}]}
    v154 = qword_1803B0280;
    *(_DWORD *)(v3 + 40) = 603;
    v106 = sub_1803717E0((_DWORD)a1, v104, v105, Py_NoneStruct, v154, v156);
    v107 = (int *)v106;
    if ( v106 != 0 )
    {
      if ( *(_QWORD *)(v106 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 1183 | qword_1803B0290 -> novels.api.novels_api#104 bytes [3926236, 3926256]: "friendly_llm_error"
        v108 = sub_180371D80((_DWORD)a1, (_DWORD)v107, qword_1803CA830, qword_1803B0290, qword_1803AB250);
      else
// Original IDA line 1185 | qword_1803B0290 -> novels.api.novels_api#104 bytes [3926236, 3926256]: "friendly_llm_error"
        v108 = sub_180371890(a1, v107, qword_1803B0290);
      v109 = v108;
      if ( v108 == 0
        || ((v110 = *((_QWORD *)v107 + 1), v170 = (int *)v108, v110 == PyModule_Type)
         || (unsigned int)PyType_IsSubtype() != 0
// Original IDA line 1190 | qword_1803B0298 -> novels.api.novels_api#105 bytes [3926256, 3926280]: "is_transient_llm_error"
          ? (v111 = sub_180371D80((_DWORD)a1, (_DWORD)v107, qword_1803CA830, qword_1803B0298, qword_1803AB250))
// Original IDA line 1191 | qword_1803B0298 -> novels.api.novels_api#105 bytes [3926256, 3926280]: "is_transient_llm_error"
          : (v111 = sub_180371890(a1, v107, qword_1803B0298)),
            v112 = v111,
            v111 == 0) )
      {
LABEL_262:
        v94 = (int *)a1[12];
        a1[12] = 0;
        if ( *v107 >= 0 )
        {
          v7 = (*(_QWORD *)v107)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v107 + 1) + 48LL))(v107);
        }
        goto LABEL_278;
      }
      v169 = (int *)v111;
      if ( *v107 >= 0 )
      {
        v7 = (*(_QWORD *)v107)-- == 1;
        if ( v7 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v107 + 1) + 48LL))(v107);
      }
      v95 = 604;
      *(_DWORD *)(v178 + 40) = 604;
      v113 = sub_180376770(a1, v112, v103);
      v107 = (int *)v113;
      if ( v113 != 0 )
      {
        v114 = sub_180001510(v113);
        if ( v114 == -1 )
          goto LABEL_262;
        if ( *v107 >= 0 )
        {
          v7 = (*(_QWORD *)v107)-- == 1;
          if ( v7 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v107 + 1) + 48LL))(v107);
        }
        if ( v114 != 0 )
        {
          *(_DWORD *)(v178 + 40) = 604;
          v115 = sub_180376770(a1, v109, v103);
        }
        else
        {
          v115 = sub_1803749A0(v103);
        }
        v116 = v115;
        if ( v115 != 0 )
        {
// Original IDA line 1240 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
          v117 = qword_1803B0120;
          v4 = (int *)v115;
          v23 = PyDict_NewPresized(2);
          PyDict_SetItem(v23, v117, Py_FalseStruct);
// Original IDA line 1244 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
          PyDict_SetItem(v23, qword_1803B0128, v116);
          if ( v103 != 0 && *(int *)v103 >= 0 )
          {
            v7 = (*(_QWORD *)v103)-- == 1;
            if ( v7 )
              (*(void (__fastcall **)(__int64))(*(_QWORD *)(v103 + 8) + 48LL))(v103);
          }
          v86 = (int *)v165[0];
          goto LABEL_287;
        }
      }
    }
    v94 = (int *)a1[12];
    a1[12] = 0;
LABEL_278:
    if ( v103 != 0 && *(int *)v103 >= 0 )
    {
      v7 = (*(_QWORD *)v103)-- == 1;
      if ( v7 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v103 + 8) + 48LL))(v103);
    }
    goto LABEL_298;
  }
// Original IDA line 1267 | qword_1803B0218 -> novels.api.novels_api#89 bytes [3925911, 3925956]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3925913, 3925922], "value": "success"}, {"tag": "a", "fileRange": [3925922, 3925931], "value": "message"}], "values": [{"tag": "F", "fileRange": [3925931, 3925932], "value": false}, {"tag... (truncated literal)
  v23 = sub_18036FB80(a1, qword_1803B0218);
  sub_180001D80(a1, v24);
  v25 = nullptr;
  v26 = nullptr;
  v27 = nullptr;
LABEL_388:
  v119 = nullptr;
LABEL_389:
  v146 = v171;
  if ( *v171 >= 0 )
  {
    v7 = (*(_QWORD *)v171)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v146 + 1) + 48LL))(v146);
  }
LABEL_392:
  v147 = v172;
  if ( v172 != nullptr && *v172 >= 0 )
  {
    v7 = (*(_QWORD *)v172)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v147 + 1) + 48LL))();
  }
  v148 = v173;
  if ( v173 != nullptr && *v173 >= 0 )
  {
    v7 = (*(_QWORD *)v173)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v148 + 1) + 48LL))();
  }
  if ( v26 != nullptr && *v26 >= 0 )
  {
    v7 = (*(_QWORD *)v26)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v26 + 1) + 48LL))(v26);
  }
  if ( v27 != nullptr && *v27 >= 0 )
  {
    v7 = (*(_QWORD *)v27)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v27 + 1) + 48LL))(v27);
  }
  if ( v25 != nullptr && *v25 >= 0 )
  {
    v7 = (*(_QWORD *)v25)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v25 + 1) + 48LL))(v25);
  }
  v149 = v174;
  if ( v174 != nullptr && *v174 >= 0 )
  {
    v7 = (*(_QWORD *)v174)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v149 + 1) + 48LL))();
  }
  v150 = v175;
  if ( v175 != nullptr && *v175 >= 0 )
  {
    v7 = (*(_QWORD *)v175)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v150 + 1) + 48LL))();
  }
  v151 = v170;
  if ( v170 != nullptr && *v170 >= 0 )
  {
    v7 = (*(_QWORD *)v170)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v151 + 1) + 48LL))();
  }
  if ( v119 != nullptr && *v119 >= 0 )
  {
    v7 = (*(_QWORD *)v119)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v119 + 1) + 48LL))(v119);
  }
  if ( v4 != nullptr && *v4 >= 0 )
  {
    v7 = (*(_QWORD *)v4)-- == 1;
    if ( v7 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v4 + 1) + 48LL))(v4);
  }
  v152 = v176;
  if ( *v176 >= 0 )
  {
    v7 = (*(_QWORD *)v176)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v152 + 1) + 48LL))();
  }
  v153 = v166;
  if ( *v166 >= 0 )
  {
    v7 = (*(_QWORD *)v166)-- == 1;
    if ( v7 )
      (*(void (**)(void))(*((_QWORD *)v153 + 1) + 48LL))();
  }
  return v23;
}
