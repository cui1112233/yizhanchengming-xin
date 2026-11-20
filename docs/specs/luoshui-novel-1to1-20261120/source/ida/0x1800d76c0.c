// write access to const memory has been detected, the output may be wrong!
int *__fastcall sub_1800D76C0(_QWORD *a1, __int64 a2, int **a3)
{
  __int64 v3; // rbx
  __int64 v4; // r13
  int *v6; // r15
  int *v7; // r12
  bool v8; // zf
  __int64 v9; // rdi
  int v10; // edx
  __int64 v11; // rcx
  __int64 v12; // rbx
  __int64 v13; // rcx
  __int64 v14; // r14
  __int64 v15; // rax
  __int64 *v16; // rax
  __int64 *v17; // rax
  __int64 v18; // rbx
  unsigned int v19; // edi
  __int64 v20; // rax
  int *v21; // r14
  int *v22; // rbx
  __int64 v23; // rdx
  int v24; // r14d
  int *v25; // r14
  __int64 v26; // rbx
  int v27; // edx
  __int64 v28; // rcx
  __int64 v29; // rdx
  __int64 v30; // rdi
  __int64 v31; // rdx
  __int64 v32; // rax
  __int64 *v33; // rax
  __int64 *v34; // rax
  __int64 v35; // rax
  int *v36; // r14
  __int64 v37; // rdx
  int v38; // ebx
  __int64 v39; // rax
  __int64 v40; // rcx
  int *v41; // r14
  __int64 v42; // rax
  int *v43; // rbx
  int *v44; // r15
  __int64 v45; // rcx
  int *v46; // r12
  int *v47; // rbx
  __int64 v48; // r14
  int *v49; // rbx
  int v50; // eax
  __int64 v51; // rdx
  int *v52; // r12
  __int64 v53; // rax
  __int64 v54; // r14
  __int64 v55; // rbx
  int v56; // edx
  __int64 v57; // rcx
  __int64 v58; // rdx
  __int64 v59; // rdi
  __int64 v60; // rdx
  __int64 v61; // rax
  __int64 *v62; // rax
  __int64 *v63; // rax
  int *v64; // rax
  int *v65; // rcx
  __int64 v66; // rax
  int *v67; // rbx
  int *v68; // r14
  int v69; // eax
  __int64 v70; // r15
  __int64 v71; // rbx
  int v72; // edx
  __int64 v73; // rcx
  __int64 v74; // rax
  __int64 v75; // rdi
  __int64 v76; // rdx
  __int64 v77; // rax
  __int64 *v78; // rax
  __int64 *v79; // rax
  int *v80; // rax
  int *v81; // rbx
  int *v82; // rax
  int *v83; // rcx
  int *v84; // rbx
  __int64 v85; // rax
  __int64 v86; // rcx
  __int64 v87; // rdx
  int *v88; // r14
  __int64 v89; // rbx
  __int64 v90; // r15
  __int64 v91; // rax
  int *v92; // rbx
  int *v93; // rdi
  int *v94; // rbx
  __int64 v95; // r14
  __int64 v96; // rax
  int *v97; // rdi
  int *v98; // rcx
  int *v99; // r14
  __int64 v100; // r8
  __int64 v101; // r9
  __int64 v102; // rcx
  __int64 v103; // rcx
  __int64 v104; // rdi
  int *v105; // rcx
  _QWORD *v106; // rcx
  int *v107; // rcx
  int *v108; // rcx
  int *v109; // rcx
  int *v110; // rcx
  int *v111; // rcx
  int *v113; // rcx
  int *v114; // rcx
  int *v115; // rcx
  int *v116; // rcx
  int *v117; // rax
  int *v118; // [rsp+28h] [rbp-61h]
  __int64 v119; // [rsp+58h] [rbp-31h] BYREF
  int *v120; // [rsp+60h] [rbp-29h]
  _QWORD *v121; // [rsp+68h] [rbp-21h]
  int *v122; // [rsp+70h] [rbp-19h]
  int *v123; // [rsp+78h] [rbp-11h]
  int *v124; // [rsp+80h] [rbp-9h] BYREF
  __int64 v125; // [rsp+88h] [rbp-1h]
  _QWORD *v126; // [rsp+90h] [rbp+7h]
  __int64 v127; // [rsp+98h] [rbp+Fh]
  int *v128; // [rsp+A0h] [rbp+17h]
  int *v129; // [rsp+A8h] [rbp+1Fh]
  int *v130; // [rsp+100h] [rbp+77h]
  int *v131; // [rsp+108h] [rbp+7Fh] BYREF

  v3 = 0;
  v4 = qword_1803B3610;
  v6 = a3[1];
  v7 = a3[2];
  v120 = *a3;
  v123 = a3[3];
  v129 = a3[4];
  v119 = 0;
  v130 = v6;
  v131 = v7;
  v121 = nullptr;
  v122 = nullptr;
  v128 = nullptr;
  if ( qword_1803B3610 == 0 )
  {
LABEL_8:
    v4 = sub_18036CD10(a1, *(_QWORD *)&qword_1803B29A0, qword_1803CA7A0, 64);
    qword_1803B3610 = v4;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803B3610 > 1 || *(_QWORD *)(qword_1803B3610 + 16) != 0 )
  {
    if ( qword_1803B3610 != 0 && *(int *)qword_1803B3610 >= 0 )
    {
      v8 = (*(_QWORD *)qword_1803B3610)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v4 + 8) + 48LL))(v4);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v4);
  v9 = *(_QWORD *)(qword_1803CA798 + 32);
  if ( *(_BYTE *)(v9 + 10) == 0 )
  {
    v16 = (__int64 *)sub_1800015A0(qword_1803CA798, qword_1803B3278);
    if ( v16 != nullptr )
      v3 = *v16;
LABEL_21:
    if ( v3 != 0 )
      goto LABEL_25;
    goto LABEL_22;
  }
  v10 = *(_DWORD *)(v9 + 12);
  if ( v10 == 0 )
  {
    v11 = a1[2];
    v10 = *(_DWORD *)(v11 + 270616);
    *(_DWORD *)(v11 + 270616) = v10 + 1;
    *(_DWORD *)(v9 + 12) = v10;
  }
  if ( v10 == dword_1803A9784 )
  {
    v12 = qword_1803B2988;
  }
  else
  {
    dword_1803A9784 = v10;
    v12 = sub_1803707D0(v9, qword_1803B3278, *(_QWORD *)(qword_1803B3278 + 24));
    qword_1803B2988 = v12;
  }
  if ( v12 >= 0 )
  {
    v13 = *(unsigned __int8 *)(v9 + 9);
    v14 = v9 + (1LL << v13);
    v3 = *(_QWORD *)(v14 + 16 * v12 + 40);
    if ( v3 != 0 )
      goto LABEL_25;
    v15 = sub_1803707D0(v9, qword_1803B3278, *(_QWORD *)(qword_1803B3278 + 24));
    qword_1803B2988 = v15;
    if ( v15 >= 0 )
    {
      v3 = *(_QWORD *)(v14 + 16 * v15 + 40);
      goto LABEL_21;
    }
  }
LABEL_22:
  v17 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B3278);
  if ( v17 == nullptr || (v3 = *v17, *v17 == 0) )
  {
    sub_180373340(a1, &v119, qword_1803B3278);
    v18 = v119;
    v19 = 519;
    goto LABEL_207;
  }
LABEL_25:
  v20 = sub_18037F0C0(v13, v120, qword_1803B2BA8);
  v21 = (int *)v20;
  v19 = 519;
  if ( v20 == 0 )
  {
    v18 = a1[12];
    a1[12] = 0;
    goto LABEL_207;
  }
  *(_DWORD *)(v4 + 40) = 519;
  v22 = (int *)sub_180376770(a1, v3, v20);
  if ( *v21 >= 0 )
  {
    v8 = (*(_QWORD *)v21)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v21 + 1) + 48LL))(v21);
  }
  if ( v22 == nullptr )
  {
    v18 = a1[12];
    a1[12] = 0;
    goto LABEL_207;
  }
  v24 = sub_180001510(v22);
  if ( *v22 >= 0 )
  {
    v8 = (*(_QWORD *)v22)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v22 + 1) + 48LL))(v22);
  }
  if ( v24 == -1 )
  {
    v18 = a1[12];
    a1[12] = 0;
    goto LABEL_207;
  }
  if ( v24 != 0 )
  {
    v26 = *(_QWORD *)(qword_1803CA798 + 32);
    if ( *(_BYTE *)(v26 + 10) != 0 )
    {
      v27 = *(_DWORD *)(v26 + 12);
      if ( v27 == 0 )
      {
        v28 = a1[2];
        v27 = *(_DWORD *)(v28 + 270616);
        *(_DWORD *)(v28 + 270616) = v27 + 1;
        *(_DWORD *)(v26 + 12) = v27;
      }
      if ( v27 == dword_1803A9794 )
      {
        v29 = qword_1803B3608;
      }
      else
      {
        dword_1803A9794 = v27;
        v29 = sub_1803707D0(v26, qword_1803B3280, *(_QWORD *)(qword_1803B3280 + 24));
        qword_1803B3608 = v29;
      }
      if ( v29 < 0 )
        goto LABEL_55;
      v30 = v26 + (1LL << *(_BYTE *)(v26 + 9));
      v31 = *(_QWORD *)(v30 + 16 * v29 + 40);
      if ( v31 != 0 )
        goto LABEL_58;
      v32 = sub_1803707D0(v26, qword_1803B3280, *(_QWORD *)(qword_1803B3280 + 24));
      qword_1803B3608 = v32;
      if ( v32 < 0 )
      {
LABEL_55:
        v34 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B3280);
        if ( v34 == nullptr || (v31 = *v34, *v34 == 0) )
        {
          sub_180373340(a1, &v119, qword_1803B3280);
          v18 = v119;
          v19 = 521;
          goto LABEL_207;
        }
LABEL_58:
        v19 = 521;
        v125 = qword_1803B3288;
        *(_DWORD *)(v4 + 40) = 521;
        v124 = v6;
        v35 = sub_180377750(a1, v31, &v124);
        v36 = (int *)v35;
        if ( v35 == 0 )
        {
          v18 = a1[12];
          a1[12] = 0;
          goto LABEL_207;
        }
        v38 = sub_180001510(v35);
        if ( v38 == -1 )
        {
          v18 = a1[12];
          a1[12] = 0;
          if ( *v36 >= 0 )
          {
            v8 = (*(_QWORD *)v36)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
          }
          goto LABEL_207;
        }
        if ( *v36 >= 0 )
        {
          v8 = (*(_QWORD *)v36)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
        }
        if ( v38 != 0 )
        {
          v25 = (int *)Py_NoneStruct;
          if ( Py_NoneStruct != -1 )
            ++Py_NoneStruct;
          sub_180001D80(a1, v37);
          goto LABEL_286;
        }
        v39 = sub_1800C7020(a1);
        if ( v39 == 0 )
        {
          sub_180373340(a1, &v119, qword_1803B2B20);
          v18 = v119;
          v19 = 523;
          goto LABEL_207;
        }
        v41 = (int *)sub_18037F0C0(v40, v39, qword_1803B2CE8);
        if ( v41 == nullptr )
        {
          v18 = a1[12];
          v19 = 523;
          a1[12] = 0;
          goto LABEL_207;
        }
        v42 = sub_180382540(v6, qword_1803B2C90);
        v43 = (int *)v42;
        if ( v42 == 0 )
        {
          v18 = a1[12];
          a1[12] = 0;
          if ( *v41 >= 0 )
          {
            v8 = (*(_QWORD *)v41)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
          }
          v19 = 523;
LABEL_207:
          v95 = *(_QWORD *)(v18 + 40);
          if ( v95 != 0 )
          {
            if ( *(_QWORD *)(v95 + 24) == v4 )
            {
LABEL_226:
              v99 = v128;
              v118 = v123;
              sub_18036D0A0(v4, "oooooooo", v120);
              v102 = qword_1803B3610;
              if ( v4 == qword_1803B3610 )
              {
                if ( *(int *)qword_1803B3610 >= 0 )
                {
                  v8 = (*(_QWORD *)qword_1803B3610)-- == 1;
                  if ( v8 )
                    (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *))(*(_QWORD *)(v102 + 8)
                                                                                            + 48LL))(
                      v102,
                      *(_QWORD *)(*(_QWORD *)(v102 + 8) + 48LL),
                      v100,
                      v101,
                      v7,
                      v118);
                }
                qword_1803B3610 = 0;
              }
              v103 = a1[7];
              v104 = *(_QWORD *)(*(_QWORD *)v103 + 48LL);
              *(_QWORD *)v103 = *(_QWORD *)(*(_QWORD *)v103 + 8LL);
              *(_DWORD *)(v104 + 64) = -1;
              v105 = *(int **)(v104 + 16);
              if ( v105 != nullptr )
              {
                *(_QWORD *)(v104 + 16) = 0;
                if ( *v105 >= 0 )
                {
                  v8 = (*(_QWORD *)v105)-- == 1;
                  if ( v8 )
                    (*(void (**)(void))(*((_QWORD *)v105 + 1) + 48LL))();
                }
              }
              if ( *(int *)v104 >= 0 )
              {
                v8 = (*(_QWORD *)v104)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(__int64))(*(_QWORD *)(v104 + 8) + 48LL))(v104);
              }
              *(_QWORD *)(v104 + 80) = 0;
              v106 = v121;
              if ( v121 != nullptr && *(int *)v121 >= 0 )
              {
                v8 = (*v121)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(v106[1] + 48LL))();
              }
              v107 = v122;
              if ( v122 != nullptr && *v122 >= 0 )
              {
                v8 = (*(_QWORD *)v122)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v107 + 1) + 48LL))();
              }
              if ( v99 != nullptr && *v99 >= 0 )
              {
                v8 = (*(_QWORD *)v99)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v99 + 1) + 48LL))(v99);
              }
              v108 = v120;
              if ( *v120 >= 0 )
              {
                v8 = (*(_QWORD *)v120)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v108 + 1) + 48LL))();
              }
              if ( *v6 >= 0 )
              {
                v8 = (*(_QWORD *)v6)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
              }
              if ( *v7 >= 0 )
              {
                v8 = (*(_QWORD *)v7)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
              }
              v109 = v123;
              if ( *v123 >= 0 )
              {
                v8 = (*(_QWORD *)v123)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v109 + 1) + 48LL))();
              }
              v110 = v129;
              if ( *v129 >= 0 )
              {
                v8 = (*(_QWORD *)v129)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v110 + 1) + 48LL))();
              }
              v111 = (int *)a1[12];
              a1[12] = v18;
              if ( v111 != nullptr && *v111 >= 0 )
              {
                v8 = (*(_QWORD *)v111)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v111 + 1) + 48LL))();
              }
              return nullptr;
            }
            v96 = sub_18038E2A0(v4, v19);
            v97 = (int *)v96;
            *(_QWORD *)(v96 + 16) = v95;
            if ( *(_DWORD *)v95 != -1 )
              ++*(_DWORD *)v95;
            v98 = *(int **)(v18 + 40);
            if ( v96 == Py_NoneStruct )
            {
              v96 = 0;
LABEL_218:
              *(_QWORD *)(v18 + 40) = v96;
              if ( v98 != nullptr && *v98 >= 0 )
              {
                v8 = (*(_QWORD *)v98)-- == 1;
                if ( v8 )
                  (*(void (**)(void))(*((_QWORD *)v98 + 1) + 48LL))();
              }
              if ( v97 != nullptr && *v97 >= 0 )
              {
                v8 = (*(_QWORD *)v97)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v97 + 1) + 48LL))(v97);
              }
              goto LABEL_226;
            }
          }
          else
          {
            v96 = sub_18038E2A0(v4, v19);
            v97 = (int *)v96;
            v98 = *(int **)(v18 + 40);
            if ( v96 == Py_NoneStruct )
            {
              v96 = 0;
              goto LABEL_218;
            }
          }
          if ( v97 != nullptr && *(_DWORD *)v96 != -1 )
            ++*(_DWORD *)v96;
          goto LABEL_218;
        }
        v44 = (int *)sub_180382540(v42, qword_1803B2F70);
        if ( *v43 >= 0 )
        {
          v8 = (*(_QWORD *)v43)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v43 + 1) + 48LL))(v43);
        }
        if ( v44 == nullptr )
        {
          v18 = a1[12];
          a1[12] = 0;
          if ( *v41 >= 0 )
          {
            v8 = (*(_QWORD *)v41)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
          }
          v19 = 523;
          goto LABEL_206;
        }
        v19 = 523;
        v125 = qword_1803B2B10;
        *(_DWORD *)(v4 + 40) = 523;
        v124 = v44;
        v46 = (int *)sub_180377750(a1, v41, &v124);
        if ( *v41 >= 0 )
        {
          v8 = (*(_QWORD *)v41)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
        }
        if ( *v44 >= 0 )
        {
          v8 = (*(_QWORD *)v44)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v44 + 1) + 48LL))(v44);
        }
        if ( v46 == nullptr )
          goto LABEL_95;
        v47 = (int *)sub_18037F0C0(v45, v46, qword_1803B2B18);
        if ( *v46 >= 0 )
        {
          v8 = (*(_QWORD *)v46)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v46 + 1) + 48LL))(v46);
        }
        if ( v47 == nullptr )
          goto LABEL_95;
        *(_DWORD *)(v4 + 40) = 523;
        v48 = sub_180376210(a1, v47);
        if ( *v47 >= 0 )
        {
          v8 = (*(_QWORD *)v47)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v47 + 1) + 48LL))(v47);
        }
        if ( v48 == 0 )
          goto LABEL_95;
        v49 = (int *)v48;
        v121 = (_QWORD *)v48;
        v50 = sub_180001510(v48);
        if ( v50 == -1 )
        {
          v19 = 524;
LABEL_95:
          v18 = a1[12];
          v7 = v131;
          a1[12] = 0;
LABEL_206:
          v6 = v130;
          goto LABEL_207;
        }
        if ( v50 == 0 )
        {
          v25 = (int *)Py_NoneStruct;
          if ( Py_NoneStruct != -1 )
            ++Py_NoneStruct;
          v7 = v131;
          goto LABEL_274;
        }
        v52 = v123;
        v53 = sub_180370DC0(a1, v123);
        v54 = v53;
        if ( v53 == 0 )
        {
          v19 = 526;
          goto LABEL_95;
        }
        v122 = (int *)v53;
        v55 = *(_QWORD *)(qword_1803CA798 + 32);
        if ( *(_BYTE *)(v55 + 10) != 0 )
        {
          v56 = *(_DWORD *)(v55 + 12);
          if ( v56 == 0 )
          {
            v57 = a1[2];
            v56 = *(_DWORD *)(v57 + 270616);
            *(_DWORD *)(v57 + 270616) = v56 + 1;
            *(_DWORD *)(v55 + 12) = v56;
          }
          if ( v56 == dword_1803A9768 )
          {
            v58 = qword_1803B2A98;
          }
          else
          {
            dword_1803A9768 = v56;
            v58 = sub_1803707D0(v55, qword_1803B3290, *(_QWORD *)(qword_1803B3290 + 24));
            qword_1803B2A98 = v58;
          }
          if ( v58 < 0 )
            goto LABEL_126;
          v59 = v55 + (1LL << *(_BYTE *)(v55 + 9));
          v60 = *(_QWORD *)(v59 + 16 * v58 + 40);
          if ( v60 != 0 )
            goto LABEL_129;
          v61 = sub_1803707D0(v55, qword_1803B3290, *(_QWORD *)(qword_1803B3290 + 24));
          qword_1803B2A98 = v61;
          if ( v61 < 0 )
          {
LABEL_126:
            v63 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B3290);
            if ( v63 == nullptr || (v60 = *v63, *v63 == 0) )
            {
              sub_180373340(a1, &v119, qword_1803B3290);
              v18 = v119;
              v19 = 527;
              v7 = v131;
              goto LABEL_206;
            }
LABEL_129:
            v6 = v130;
            v19 = 527;
            v124 = v130;
            *(_DWORD *)(v4 + 40) = 527;
            v125 = v54;
            v64 = (int *)sub_180377750(a1, v60, &v124);
            v65 = v64;
            if ( v64 != nullptr )
            {
              if ( *v64 >= 0 )
              {
                v8 = (*(_QWORD *)v64)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v64 + 1) + 48LL))(v64);
              }
              v66 = sub_18037F0C0(v65, v52, qword_1803B2B00);
              v67 = (int *)v66;
              v19 = 528;
              if ( v66 != 0 )
              {
                *(_DWORD *)(v4 + 40) = 528;
                v68 = (int *)sub_1803770D0(a1, v66, qword_1803B2B40);
                if ( *v67 >= 0 )
                {
                  v8 = (*(_QWORD *)v67)-- == 1;
                  if ( v8 )
                    (*(void (__fastcall **)(int *))(*((_QWORD *)v67 + 1) + 48LL))(v67);
                }
                if ( v68 != nullptr )
                {
                  v69 = sub_180001510(v68);
                  if ( v69 == -1 )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    if ( *v68 >= 0 )
                    {
                      v8 = (*(_QWORD *)v68)-- == 1;
                      if ( v8 )
                      {
                        (*(void (__fastcall **)(int *))(*((_QWORD *)v68 + 1) + 48LL))(v68);
                        v7 = v131;
                        goto LABEL_207;
                      }
                    }
LABEL_131:
                    v7 = v131;
                    goto LABEL_207;
                  }
                  if ( v69 == 1 )
                  {
                    v7 = v131;
                  }
                  else
                  {
                    if ( *v68 >= 0 )
                    {
                      v8 = (*(_QWORD *)v68)-- == 1;
                      if ( v8 )
                        (*(void (__fastcall **)(int *))(*((_QWORD *)v68 + 1) + 48LL))(v68);
                    }
                    v7 = v131;
                    if ( *v131 != -1 )
                      ++*v131;
                    v68 = v7;
                  }
                  v70 = sub_1803749A0(v68);
                  if ( *v68 >= 0 )
                  {
                    v8 = (*(_QWORD *)v68)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v68 + 1) + 48LL))(v68);
                  }
                  if ( v70 == 0 )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    goto LABEL_206;
                  }
                  v128 = (int *)v70;
                  v71 = *(_QWORD *)(qword_1803CA798 + 32);
                  if ( *(_BYTE *)(v71 + 10) != 0 )
                  {
                    v72 = *(_DWORD *)(v71 + 12);
                    if ( v72 == 0 )
                    {
                      v73 = a1[2];
                      v72 = *(_DWORD *)(v73 + 270616);
                      *(_DWORD *)(v73 + 270616) = v72 + 1;
                      *(_DWORD *)(v71 + 12) = v72;
                    }
                    if ( v72 == dword_1803A978C )
                    {
                      v74 = qword_1803B2968;
                    }
                    else
                    {
                      dword_1803A978C = v72;
                      v74 = sub_1803707D0(v71, qword_1803B3298, *(_QWORD *)(qword_1803B3298 + 24));
                      qword_1803B2968 = v74;
                    }
                    if ( v74 < 0 )
                      goto LABEL_171;
                    v75 = v71 + (1LL << *(_BYTE *)(v71 + 9));
                    v76 = *(_QWORD *)(v75 + 16 * v74 + 40);
                    if ( v76 != 0 )
                      goto LABEL_174;
                    v77 = sub_1803707D0(v71, qword_1803B3298, *(_QWORD *)(qword_1803B3298 + 24));
                    qword_1803B2968 = v77;
                    if ( v77 < 0 )
                    {
LABEL_171:
                      v79 = (__int64 *)sub_1800015A0(qword_1803AB778, qword_1803B3298);
                      if ( v79 == nullptr || (v76 = *v79, *v79 == 0) )
                      {
                        sub_180373340(a1, &v119, qword_1803B3298);
                        v18 = v119;
                        v19 = 529;
                        goto LABEL_206;
                      }
                      goto LABEL_174;
                    }
                    v76 = *(_QWORD *)(v75 + 16 * v77 + 40);
                  }
                  else
                  {
                    v78 = (__int64 *)sub_1800015A0(qword_1803CA798, qword_1803B3298);
                    if ( v78 != nullptr )
                      v76 = *v78;
                    else
                      v76 = 0;
                  }
                  if ( v76 == 0 )
                    goto LABEL_171;
LABEL_174:
                  v125 = qword_1803B3288;
                  v19 = 529;
                  *(_DWORD *)(v4 + 40) = 529;
                  v131 = v130;
                  v124 = v7;
                  v126 = (_QWORD *)v70;
                  v80 = (int *)sub_18037D030(a1, v76, &v131, &v124, qword_1803B32A0);
                  if ( v80 == nullptr )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    goto LABEL_206;
                  }
                  if ( *v80 >= 0 )
                  {
                    v8 = (*(_QWORD *)v80)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v80 + 1) + 48LL))(v80);
                  }
                  v81 = v120;
                  v19 = 530;
                  *(_DWORD *)(v4 + 40) = 530;
                  v82 = (int *)sub_18037E980(a1, v81, *(_QWORD *)&qword_1803B2EE0, qword_1803B32A8 + 24);
                  v83 = v82;
                  if ( v82 == nullptr )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    goto LABEL_206;
                  }
                  if ( *v82 >= 0 )
                  {
                    v8 = (*(_QWORD *)v82)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v82 + 1) + 48LL))(v82);
                  }
                  v84 = (int *)sub_18037F0C0(v83, v81, qword_1803B2B90);
                  v19 = 531;
                  if ( v84 == nullptr )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    goto LABEL_206;
                  }
                  v85 = qword_1803B2D28;
                  v86 = qword_1803B3288;
                  v87 = qword_1803B32B0;
                  *(_DWORD *)(v4 + 40) = 531;
                  v124 = (int *)v85;
                  v125 = v86;
                  v127 = v87;
                  v126 = v7;
                  v88 = (int *)sub_18037D280(a1, v84, qword_1803B32B8, &v124, qword_1803B32C0);
                  if ( *v84 >= 0 )
                  {
                    v8 = (*(_QWORD *)v84)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v84 + 1) + 48LL))(v84);
                  }
                  if ( v88 == nullptr )
                  {
                    v18 = a1[12];
                    a1[12] = 0;
                    goto LABEL_206;
                  }
                  if ( *v88 >= 0 )
                  {
                    v8 = (*(_QWORD *)v88)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v88 + 1) + 48LL))(v88);
                  }
                  v89 = qword_1803B2D08;
                  v25 = (int *)PyDict_NewPresized(5);
                  PyDict_SetItem(v25, v89, v7);
                  PyDict_SetItem(v25, qword_1803B2D10, v70);
                  PyDict_SetItem(v25, qword_1803B32C8, Py_TrueStruct);
                  PyDict_SetItem(v25, qword_1803B32D0, qword_1803B3288);
                  v90 = qword_1803B2F68;
                  v91 = sub_180382540(v130, qword_1803B2C90);
                  v92 = (int *)v91;
                  if ( v91 != 0 )
                  {
                    v93 = (int *)sub_180382540(v91, qword_1803B2F70);
                    if ( *v92 >= 0 )
                    {
                      v8 = (*(_QWORD *)v92)-- == 1;
                      if ( v8 )
                        (*(void (__fastcall **)(int *))(*((_QWORD *)v92 + 1) + 48LL))(v92);
                    }
                    if ( v93 != nullptr )
                    {
                      v94 = (int *)sub_1803749A0(v93);
                      if ( *v93 >= 0 )
                      {
                        v8 = (*(_QWORD *)v93)-- == 1;
                        if ( v8 )
                          (*(void (__fastcall **)(int *))(*((_QWORD *)v93 + 1) + 48LL))(v93);
                      }
                      if ( v94 != nullptr )
                      {
                        PyDict_SetItem(v25, v90, v94);
                        if ( *v94 >= 0 )
                        {
                          v8 = (*(_QWORD *)v94)-- == 1;
                          if ( v8 )
                            (*(void (__fastcall **)(int *))(*((_QWORD *)v94 + 1) + 48LL))(v94);
                        }
                        v49 = (int *)v121;
LABEL_274:
                        sub_180001D80(a1, v51);
                        if ( *v49 >= 0 )
                        {
                          v8 = (*(_QWORD *)v49)-- == 1;
                          if ( v8 )
                            (*(void (__fastcall **)(int *))(*((_QWORD *)v49 + 1) + 48LL))(v49);
                        }
                        v113 = v122;
                        if ( v122 != nullptr && *v122 >= 0 )
                        {
                          v8 = (*(_QWORD *)v122)-- == 1;
                          if ( v8 )
                            (*(void (**)(void))(*((_QWORD *)v113 + 1) + 48LL))();
                        }
                        v114 = v128;
                        if ( v128 != nullptr && *v128 >= 0 )
                        {
                          v8 = (*(_QWORD *)v128)-- == 1;
                          if ( v8 )
                            (*(void (**)(void))(*((_QWORD *)v114 + 1) + 48LL))();
                        }
                        v6 = v130;
                        goto LABEL_286;
                      }
                    }
                  }
                  v18 = a1[12];
                  v19 = 543;
                  a1[12] = 0;
                  if ( *v25 >= 0 )
                  {
                    v8 = (*(_QWORD *)v25)-- == 1;
                    if ( v8 )
                      (*(void (__fastcall **)(int *))(*((_QWORD *)v25 + 1) + 48LL))(v25);
                  }
                  goto LABEL_206;
                }
              }
            }
            v18 = a1[12];
            a1[12] = 0;
            goto LABEL_131;
          }
          v60 = *(_QWORD *)(v59 + 16 * v61 + 40);
        }
        else
        {
          v62 = (__int64 *)sub_1800015A0(qword_1803CA798, qword_1803B3290);
          if ( v62 != nullptr )
            v60 = *v62;
          else
            v60 = 0;
        }
        if ( v60 != 0 )
          goto LABEL_129;
        goto LABEL_126;
      }
      v31 = *(_QWORD *)(v30 + 16 * v32 + 40);
    }
    else
    {
      v33 = (__int64 *)sub_1800015A0(qword_1803CA798, qword_1803B3280);
      if ( v33 != nullptr )
        v31 = *v33;
      else
        v31 = 0;
    }
    if ( v31 != 0 )
      goto LABEL_58;
    goto LABEL_55;
  }
  v25 = (int *)Py_NoneStruct;
  if ( Py_NoneStruct != -1 )
    ++Py_NoneStruct;
  sub_180001D80(a1, v23);
LABEL_286:
  v115 = v120;
  if ( *v120 >= 0 )
  {
    v8 = (*(_QWORD *)v120)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v115 + 1) + 48LL))();
  }
  if ( *v6 >= 0 )
  {
    v8 = (*(_QWORD *)v6)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
  }
  if ( *v7 >= 0 )
  {
    v8 = (*(_QWORD *)v7)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
  }
  v116 = v123;
  if ( *v123 >= 0 )
  {
    v8 = (*(_QWORD *)v123)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v116 + 1) + 48LL))();
  }
  v117 = v129;
  if ( *v129 >= 0 )
  {
    v8 = (*(_QWORD *)v129)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v117 + 1) + 48LL))(v117);
  }
  return v25;
}
