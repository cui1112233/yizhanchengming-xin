// write access to const memory has been detected, the output may be wrong!
__int64 __fastcall sub_180157690(__int64 a1, __int64 a2, int **a3)
{
  __int64 v3; // r15
  int *v4; // r13
  int *v5; // rbp
  int *v7; // r12
  bool v8; // zf
  __int64 v9; // rcx
  int v10; // edi
  __int64 v11; // rax
  int *v12; // r14
  __int64 v13; // rbx
  unsigned int v14; // edi
  __int64 v15; // rax
  __int64 *v16; // rdi
  int *v17; // rbx
  __int64 *v18; // r13
  __int64 v19; // rcx
  _DWORD *v20; // rdi
  int *v21; // r14
  int *v22; // rbx
  __int64 v23; // rdx
  unsigned __int64 v24; // rax
  __int64 v25; // rax
  __int64 v26; // rcx
  __int64 v27; // rax
  int *v28; // rbp
  __int64 v29; // rax
  __int64 *v30; // rdi
  int *v31; // rbx
  __int64 *v32; // r13
  __int64 v33; // rbp
  __int64 *v34; // rdi
  __int64 v35; // rcx
  __int64 *v36; // rbp
  int *v37; // rbx
  int *v38; // r14
  __int64 *v39; // r13
  __int64 v40; // rbx
  int v41; // edx
  __int64 v42; // rcx
  __int64 v43; // rdx
  __int64 v44; // rdi
  __int64 v45; // rdx
  __int64 v46; // rax
  __int64 *v47; // rax
  __int64 *v48; // rax
  __int64 v49; // rax
  int *v50; // rbx
  int v51; // eax
  __int64 v52; // rdx
  __int64 v53; // r13
  __int64 v54; // rcx
  _DWORD *v55; // rdi
  int *v56; // rbx
  __int64 v57; // rdx
  unsigned __int64 v58; // rcx
  __int64 v59; // rbp
  __int64 v60; // rax
  int *v61; // rbx
  int *v62; // rdi
  __int64 v63; // rbx
  __int64 v64; // rcx
  int *v65; // rbx
  int *v66; // r14
  __int64 v67; // r14
  __int64 v68; // rax
  int *v69; // rdi
  int *v70; // rcx
  __int64 v71; // rdx
  __int64 v72; // rcx
  int *v73; // rcx
  __int64 v75; // [rsp+40h] [rbp-48h]
  int *v76; // [rsp+48h] [rbp-40h]
  int *v77; // [rsp+A0h] [rbp+18h]
  __int64 Src; // [rsp+A8h] [rbp+20h] BYREF

  v3 = qword_1803B6BB0;
  v4 = nullptr;
  v5 = *a3;
  v7 = a3[1];
  v76 = *a3;
  v77 = nullptr;
  Src = 0;
  if ( qword_1803B6BB0 == 0 )
  {
LABEL_8:
    v3 = sub_18036CD10(a1, *(_QWORD *)&qword_1803B6A38, *(_QWORD *)&qword_1803CA758, 24);
    qword_1803B6BB0 = v3;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803B6BB0 > 1 || *(_QWORD *)(qword_1803B6BB0 + 16) != 0 )
  {
    if ( qword_1803B6BB0 != 0 && *(int *)qword_1803B6BB0 >= 0 )
    {
      v8 = (*(_QWORD *)qword_1803B6BB0)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v3 + 8) + 48LL))(v3);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v3);
  v10 = v7 != (int *)Py_NoneStruct;
  if ( v7 != (int *)Py_NoneStruct )
  {
    v11 = sub_18037F0C0(v9, v5, qword_1803B6E28);
    v12 = (int *)v11;
    if ( v11 == 0 )
    {
      v13 = *(_QWORD *)(a1 + 96);
      v14 = 259;
      *(_QWORD *)(a1 + 96) = 0;
      goto LABEL_135;
    }
    v15 = sub_180373860(v11);
    v16 = (__int64 *)(a1 + 96);
    if ( v15 < 0 && *v16 != 0 )
      v17 = nullptr;
    else
      v17 = (int *)PyLong_FromSsize_t(v15);
    v18 = (__int64 *)(a1 + 96);
    if ( *v12 >= 0 )
    {
      v8 = (*(_QWORD *)v12)-- == 1;
      if ( v8 )
      {
        (*(void (__fastcall **)(int *))(*((_QWORD *)v12 + 1) + 48LL))(v12);
        v18 = (__int64 *)(a1 + 96);
      }
    }
    if ( v17 == nullptr )
    {
      v13 = *v16;
      *v16 = 0;
      v14 = 259;
LABEL_134:
      v4 = v77;
LABEL_135:
      v67 = *(_QWORD *)(v13 + 40);
      if ( v67 != 0 )
      {
        if ( *(_QWORD *)(v67 + 24) == v3 )
        {
LABEL_154:
          sub_18036D0A0(v3, "ooo", v76);
          v72 = qword_1803B6BB0;
          if ( v3 == qword_1803B6BB0 )
          {
            if ( *(int *)qword_1803B6BB0 >= 0 )
            {
              v8 = (*(_QWORD *)qword_1803B6BB0)-- == 1;
              if ( v8 )
                (*(void (__fastcall **)(__int64))(*(_QWORD *)(v72 + 8) + 48LL))(v72);
            }
            qword_1803B6BB0 = 0;
          }
          sub_180001D80(a1, v71);
          if ( v4 != nullptr && *v4 >= 0 )
          {
            v8 = (*(_QWORD *)v4)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v4 + 1) + 48LL))(v4);
          }
          if ( *v76 >= 0 )
          {
            v8 = (*(_QWORD *)v76)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v76 + 1) + 48LL))(v76);
          }
          if ( *v7 >= 0 )
          {
            v8 = (*(_QWORD *)v7)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
          }
          v73 = *(int **)(a1 + 96);
          *(_QWORD *)(a1 + 96) = v13;
          if ( v73 != nullptr && *v73 >= 0 )
          {
            v8 = (*(_QWORD *)v73)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v73 + 1) + 48LL))(v73);
          }
          return 0;
        }
        v68 = sub_18038E2A0(v3, v14);
        v69 = (int *)v68;
        *(_QWORD *)(v68 + 16) = v67;
        if ( *(_DWORD *)v67 != -1 )
          ++*(_DWORD *)v67;
        v70 = *(int **)(v13 + 40);
        if ( v68 == Py_NoneStruct )
        {
          v68 = 0;
LABEL_146:
          *(_QWORD *)(v13 + 40) = v68;
          if ( v70 != nullptr && *v70 >= 0 )
          {
            v8 = (*(_QWORD *)v70)-- == 1;
            if ( v8 )
              (*(void (**)(void))(*((_QWORD *)v70 + 1) + 48LL))();
          }
          if ( v69 != nullptr && *v69 >= 0 )
          {
            v8 = (*(_QWORD *)v69)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *))(*((_QWORD *)v69 + 1) + 48LL))(v69);
          }
          goto LABEL_154;
        }
      }
      else
      {
        v68 = sub_18038E2A0(v3, v14);
        v69 = (int *)v68;
        v70 = *(int **)(v13 + 40);
        if ( v68 == Py_NoneStruct )
        {
          v68 = 0;
          goto LABEL_146;
        }
      }
      if ( v69 != nullptr && *(_DWORD *)v68 != -1 )
        ++*(_DWORD *)v68;
      goto LABEL_146;
    }
    v10 = sub_180386680(v17, v7);
    if ( *v17 >= 0 )
    {
      v8 = (*(_QWORD *)v17)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v17 + 1) + 48LL))(v17);
    }
    if ( v10 == -1 )
    {
      v13 = *v18;
      v14 = 259;
      *v18 = 0;
      goto LABEL_134;
    }
  }
  if ( v10 == 1 )
  {
    v75 = sub_18014BC50(a1);
    if ( v75 == 0 )
    {
      sub_180373340(a1, &Src, qword_1803B6EC0);
      v13 = Src;
      v14 = 260;
      goto LABEL_134;
    }
    v19 = *(_QWORD *)(a1 + 16);
    v20 = (_DWORD *)qword_1803B6EC8;
    v21 = *(int **)(v19 + 269752);
    if ( v21 != nullptr )
    {
      *(_QWORD *)(v19 + 269752) = *((_QWORD *)v21 + 3);
      --*(_DWORD *)(v19 + 269896);
      *(_QWORD *)v21 = 1;
    }
    else
    {
      v21 = (int *)sub_18035EBD0(PyTuple_Type, 5);
    }
    *((_QWORD *)v21 + 3) = 0;
    v22 = v21 - 4;
    *((_QWORD *)v21 + 4) = 0;
    *((_QWORD *)v21 + 5) = 0;
    *((_QWORD *)v21 + 6) = 0;
    *((_QWORD *)v21 + 7) = 0;
    v23 = *(_QWORD *)(*(_QWORD *)(PyThreadState_GetCurrent() + 16) + 208LL);
    v24 = *(_QWORD *)(v23 + 8);
    *(_QWORD *)v24 = v21 - 4;
    *((_QWORD *)v22 + 1) &= 3uLL;
    *((_QWORD *)v22 + 1) |= v24;
    *(_QWORD *)v22 = v23;
    *(_QWORD *)(v23 + 8) = v21 - 4;
    if ( *v20 != -1 )
      ++*v20;
    *((_QWORD *)v21 + 3) = v20;
    v25 = sub_180375520(a1, v7, qword_1803B6C50);
    if ( v25 == 0 )
    {
      v13 = *(_QWORD *)(a1 + 96);
      *(_QWORD *)(a1 + 96) = 0;
LABEL_55:
      v14 = 261;
      if ( *v21 < 0 )
        goto LABEL_134;
      v8 = (*(_QWORD *)v21)-- == 1;
      if ( !v8 )
        goto LABEL_134;
      goto LABEL_57;
    }
    *((_QWORD *)v21 + 4) = v25;
    v26 = qword_1803B6ED0;
    if ( *(_DWORD *)qword_1803B6ED0 != -1 )
      ++*(_DWORD *)qword_1803B6ED0;
    *((_QWORD *)v21 + 5) = v26;
    v27 = sub_18037F0C0(v26, v5, qword_1803B6E28);
    v28 = (int *)v27;
    if ( v27 == 0 )
    {
      v13 = *(_QWORD *)(a1 + 96);
      *(_QWORD *)(a1 + 96) = 0;
      goto LABEL_55;
    }
    v29 = sub_180373860(v27);
    v30 = (__int64 *)(a1 + 96);
    if ( v29 < 0 && *v30 != 0 )
      v31 = nullptr;
    else
      v31 = (int *)PyLong_FromSsize_t(v29);
    v32 = (__int64 *)(a1 + 96);
    if ( *v28 >= 0 )
    {
      v8 = (*(_QWORD *)v28)-- == 1;
      if ( v8 )
      {
        (*(void (__fastcall **)(int *))(*((_QWORD *)v28 + 1) + 48LL))(v28);
        v32 = (__int64 *)(a1 + 96);
      }
    }
    if ( v31 == nullptr )
    {
      v13 = *v30;
      *v30 = 0;
      goto LABEL_55;
    }
    v33 = sub_180375520(a1, v31, qword_1803B6C50);
    v34 = v32;
    if ( *v31 >= 0 )
    {
      v8 = (*(_QWORD *)v31)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v31 + 1) + 48LL))(v31);
      v34 = (__int64 *)(a1 + 96);
    }
    if ( v33 == 0 )
    {
      v13 = *v32;
      *v32 = 0;
      goto LABEL_55;
    }
    *((_QWORD *)v21 + 6) = v33;
    v35 = qword_1803B6ED8;
    if ( *(_DWORD *)qword_1803B6ED8 != -1 )
    {
      ++*(_DWORD *)qword_1803B6ED8;
      v34 = (__int64 *)(a1 + 96);
    }
    *((_QWORD *)v21 + 7) = v35;
    v36 = v34;
    v37 = (int *)PyUnicode_Join(qword_1803B6C50, v21);
    if ( *v21 >= 0 )
    {
      v8 = (*(_QWORD *)v21)-- == 1;
      if ( v8 )
      {
        (*(void (__fastcall **)(int *))(*((_QWORD *)v21 + 1) + 48LL))(v21);
        v36 = (__int64 *)(a1 + 96);
      }
    }
    if ( v37 == nullptr )
    {
      v13 = *v34;
      *v34 = 0;
      v14 = 261;
      goto LABEL_134;
    }
    v14 = 260;
    *(_DWORD *)(v3 + 40) = 260;
    v38 = (int *)sub_180376770(a1, v75, v37);
    v39 = v36;
    if ( *v37 >= 0 )
    {
      v8 = (*(_QWORD *)v37)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v37 + 1) + 48LL))(v37);
      v39 = (__int64 *)(a1 + 96);
    }
    if ( v38 == nullptr )
    {
      v13 = *v36;
      *v36 = 0;
      v4 = nullptr;
      goto LABEL_135;
    }
    v13 = sub_180373490(a1);
    if ( *v38 >= 0 )
    {
      v8 = (*(_QWORD *)v38)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v38 + 1) + 48LL))(v38);
      v39 = (__int64 *)(a1 + 96);
    }
    if ( v13 == 0 )
    {
      v13 = *v39;
      *v39 = 0;
      goto LABEL_134;
    }
LABEL_133:
    sub_18000D1D0(a1, v13);
    goto LABEL_134;
  }
  v40 = *(_QWORD *)(qword_1803CA760 + 32);
  if ( *(_BYTE *)(v40 + 10) == 0 )
  {
    v47 = (__int64 *)sub_1800015A0(qword_1803CA760, *(_QWORD *)&qword_1803B6EE0);
    v45 = 0;
    if ( v47 != nullptr )
      v45 = *v47;
LABEL_89:
    if ( v45 != 0 )
      goto LABEL_93;
    goto LABEL_90;
  }
  v41 = *(_DWORD *)(v40 + 12);
  if ( v41 == 0 )
  {
    v42 = *(_QWORD *)(a1 + 16);
    v41 = *(_DWORD *)(v42 + 270616);
    *(_DWORD *)(v42 + 270616) = v41 + 1;
    *(_DWORD *)(v40 + 12) = v41;
  }
  if ( v41 == dword_1803A98E8 )
  {
    v43 = qword_1803B72D0;
  }
  else
  {
    dword_1803A98E8 = v41;
    v43 = sub_1803707D0(v40, *(_QWORD *)&qword_1803B6EE0, *(_QWORD *)(*(_QWORD *)&qword_1803B6EE0 + 24LL));
    qword_1803B72D0 = v43;
  }
  if ( v43 >= 0 )
  {
    v44 = v40 + (1LL << *(_BYTE *)(v40 + 9));
    v45 = *(_QWORD *)(v44 + 16 * v43 + 40);
    if ( v45 != 0 )
      goto LABEL_93;
    v46 = sub_1803707D0(v40, *(_QWORD *)&qword_1803B6EE0, *(_QWORD *)(*(_QWORD *)&qword_1803B6EE0 + 24LL));
    qword_1803B72D0 = v46;
    if ( v46 >= 0 )
    {
      v45 = *(_QWORD *)(v44 + 16 * v46 + 40);
      goto LABEL_89;
    }
  }
LABEL_90:
  v48 = (__int64 *)sub_1800015A0(qword_1803AB778, *(_QWORD *)&qword_1803B6EE0);
  if ( v48 == nullptr || (v45 = *v48, *v48 == 0) )
  {
    sub_180373340(a1, &Src, *(_QWORD *)&qword_1803B6EE0);
    v13 = Src;
    v14 = 263;
    goto LABEL_134;
  }
LABEL_93:
  v14 = 263;
  *(_DWORD *)(v3 + 40) = 263;
  v49 = sub_180376770(a1, v45, v5);
  v50 = (int *)v49;
  if ( v49 == 0 )
  {
    v13 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_134;
  }
  v77 = (int *)v49;
  v4 = (int *)v49;
  v51 = sub_180001510(v49);
  if ( v51 == -1 )
  {
    v13 = *(_QWORD *)(a1 + 96);
    v14 = 264;
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_135;
  }
  if ( v51 != 0 )
  {
    v53 = sub_18014BC50(a1);
    if ( v53 == 0 )
    {
      sub_180373340(a1, &Src, qword_1803B6EC0);
      v13 = Src;
      v14 = 265;
      goto LABEL_134;
    }
    v54 = *(_QWORD *)(a1 + 16);
    v55 = (_DWORD *)qword_1803B6EE8;
    v21 = *(int **)(v54 + 269736);
    if ( v21 != nullptr )
    {
      *(_QWORD *)(v54 + 269736) = *((_QWORD *)v21 + 3);
      --*(_DWORD *)(v54 + 269888);
      *(_QWORD *)v21 = 1;
    }
    else
    {
      v21 = (int *)sub_18035EBD0(PyTuple_Type, 3);
    }
    v56 = v21 - 4;
    *((_QWORD *)v21 + 3) = 0;
    *((_QWORD *)v21 + 4) = 0;
    *((_QWORD *)v21 + 5) = 0;
    v57 = *(_QWORD *)(*(_QWORD *)(PyThreadState_GetCurrent() + 16) + 208LL);
    v58 = *(_QWORD *)(v57 + 8);
    *(_QWORD *)v58 = v21 - 4;
    *((_QWORD *)v56 + 1) &= 3uLL;
    *((_QWORD *)v56 + 1) |= v58;
    *(_QWORD *)v56 = v57;
    *(_QWORD *)(v57 + 8) = v21 - 4;
    if ( *v55 != -1 )
      ++*v55;
    *((_QWORD *)v21 + 3) = v55;
    v59 = qword_1803B6EF0;
    v60 = sub_18000D280(v58, v77);
    v61 = (int *)v60;
    if ( v60 != 0 )
    {
      Src = sub_18036D750(v60);
      v62 = (int *)sub_180367890(
                     (int)sub_180158130,
                     qword_1803CA758,
                     qword_1803B6F00,
                     qword_1803B6F08,
                     qword_1803B6A28,
                     &Src,
                     1,
                     1096);
      if ( *v61 >= 0 )
      {
        v8 = (*(_QWORD *)v61)-- == 1;
        if ( v8 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v61 + 1) + 48LL))(v61);
      }
      v63 = PyUnicode_Join(v59, v62);
      if ( *v62 >= 0 )
      {
        v8 = (*(_QWORD *)v62)-- == 1;
        if ( v8 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v62 + 1) + 48LL))(v62);
      }
      if ( v63 != 0 )
      {
        *((_QWORD *)v21 + 4) = v63;
        v64 = qword_1803B6EF8;
        if ( *(_DWORD *)qword_1803B6EF8 != -1 )
          ++*(_DWORD *)qword_1803B6EF8;
        *((_QWORD *)v21 + 5) = v64;
        v65 = (int *)PyUnicode_Join(qword_1803B6C50, v21);
        if ( *v21 >= 0 )
        {
          v8 = (*(_QWORD *)v21)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v21 + 1) + 48LL))(v21);
        }
        if ( v65 == nullptr )
        {
          v13 = *(_QWORD *)(a1 + 96);
          v14 = 266;
          *(_QWORD *)(a1 + 96) = 0;
          goto LABEL_134;
        }
        v14 = 265;
        *(_DWORD *)(v3 + 40) = 265;
        v66 = (int *)sub_180376770(a1, v53, v65);
        if ( *v65 >= 0 )
        {
          v8 = (*(_QWORD *)v65)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v65 + 1) + 48LL))(v65);
        }
        if ( v66 == nullptr )
        {
          v13 = *(_QWORD *)(a1 + 96);
          *(_QWORD *)(a1 + 96) = 0;
          goto LABEL_134;
        }
        v13 = sub_180373490(a1);
        if ( *v66 >= 0 )
        {
          v8 = (*(_QWORD *)v66)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v66 + 1) + 48LL))(v66);
        }
        if ( v13 == 0 )
        {
          v13 = *(_QWORD *)(a1 + 96);
          *(_QWORD *)(a1 + 96) = 0;
          goto LABEL_134;
        }
        goto LABEL_133;
      }
    }
    v14 = 266;
    v13 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v21 < 0 )
      goto LABEL_134;
    v8 = (*(_QWORD *)v21)-- == 1;
    if ( !v8 )
      goto LABEL_134;
LABEL_57:
    (*(void (__fastcall **)(int *))(*((_QWORD *)v21 + 1) + 48LL))(v21);
    goto LABEL_134;
  }
  sub_180001D80(a1, v52);
  if ( Py_NoneStruct != -1 )
    ++Py_NoneStruct;
  if ( *v50 >= 0 )
  {
    v8 = (*(_QWORD *)v50)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v50 + 1) + 48LL))(v50);
  }
  if ( *v5 >= 0 )
  {
    v8 = (*(_QWORD *)v5)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v5 + 1) + 48LL))(v5);
  }
  if ( *v7 >= 0 )
  {
    v8 = (*(_QWORD *)v7)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
  }
  return Py_NoneStruct;
}
