int *__fastcall sub_18007FC20(_QWORD *a1, __int64 a2, int **a3)
{
  __int64 v3; // r15
  int *v4; // rbp
  int *v6; // r13
  int *v7; // r12
  bool v8; // zf
  __int64 v9; // rcx
  __int64 v10; // rax
  int *v11; // rbx
  __int64 v12; // rbx
  unsigned int v13; // edi
  int *v14; // r14
  __int64 v15; // rcx
  int *v16; // rbx
  int *v17; // r14
  __int64 v18; // rbx
  __int64 v19; // rax
  __int64 v20; // rcx
  __int64 v21; // rcx
  int *v22; // r14
  __int64 v23; // rax
  int *v24; // rbx
  int *v25; // rbp
  __int64 v26; // rax
  int *v27; // rcx
  __int64 v28; // rbx
  __int64 v29; // rdi
  int *v30; // r14
  int *v31; // rbp
  __int64 v32; // rax
  __int64 v33; // rax
  int *v34; // rbx
  int *v35; // rcx
  _DWORD *v36; // r14
  __int64 v37; // rax
  int *v38; // rdi
  int *v39; // rcx
  int *v40; // rcx
  __int64 v41; // r14
  __int64 v42; // rdi
  __int64 v43; // rdi
  __int64 v44; // rax
  int *v45; // rbx
  int *v46; // rbx
  unsigned int v47; // edi
  int *v48; // rcx
  __int64 v49; // rcx
  __int64 v50; // rbx
  int *v51; // rcx
  char v53; // al
  __int64 v54; // rax
  int *v55; // rcx
  __int64 v56; // r14
  __int64 v57; // rax
  int *v58; // rdi
  int *v59; // rcx
  __int64 v60; // r8
  __int64 v61; // r9
  __int64 v62; // rcx
  __int64 v63; // rcx
  __int64 v64; // rdi
  int *v65; // rcx
  int *v66; // rcx
  int *v67; // [rsp+40h] [rbp-58h]
  int *v68; // [rsp+48h] [rbp-50h] BYREF
  int *v69; // [rsp+50h] [rbp-48h] BYREF
  int *v70; // [rsp+B0h] [rbp+18h]
  int *v71; // [rsp+B8h] [rbp+20h]

  v3 = qword_1803AFB30;
  v4 = *a3;
  v6 = a3[1];
  v7 = nullptr;
  v67 = *a3;
  v71 = nullptr;
  v70 = nullptr;
  if ( qword_1803AFB30 == 0 )
  {
LABEL_8:
    v3 = sub_18036CD10(a1, *(_QWORD *)&qword_1803B0FD0, qword_1803CA828, 48);
    qword_1803AFB30 = v3;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803AFB30 > 1 || *(_QWORD *)(qword_1803AFB30 + 16) != 0 )
  {
    if ( qword_1803AFB30 != 0 && *(int *)qword_1803AFB30 >= 0 )
    {
      v8 = (*(_QWORD *)qword_1803AFB30)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v3 + 8) + 48LL))(v3);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v3);
  v10 = sub_18037F0C0(v9, v6, qword_1803AFF88);
  v11 = (int *)v10;
  if ( v10 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    v13 = 499;
    goto LABEL_60;
  }
  v13 = 499;
  *(_DWORD *)(v3 + 40) = 499;
  v14 = (int *)sub_180378210(a1, v10, qword_1803B0428);
  if ( *v11 >= 0 )
  {
    v8 = (*(_QWORD *)v11)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v11 + 1) + 48LL))(v11);
  }
  if ( v14 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  v16 = (int *)sub_1803749A0(v14);
  if ( *v14 >= 0 )
  {
    v8 = (*(_QWORD *)v14)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v14 + 1) + 48LL))(v14);
  }
  if ( v16 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  v17 = (int *)sub_18037F0C0(v15, v16, qword_1803B0140);
  if ( *v16 >= 0 )
  {
    v8 = (*(_QWORD *)v16)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v16 + 1) + 48LL))(v16);
  }
  if ( v17 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  *(_DWORD *)(v3 + 40) = 499;
  v18 = sub_180376210(a1, v17);
  if ( *v17 >= 0 )
  {
    v8 = (*(_QWORD *)v17)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v17 + 1) + 48LL))(v17);
  }
  if ( v18 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  v13 = 500;
  v71 = (int *)v18;
  *(_DWORD *)(v3 + 40) = 500;
  v19 = sub_18037E6C0(a1, v4, *(_QWORD *)&qword_1803B00D0, v18);
  if ( v19 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  v7 = (int *)v19;
  v22 = (int *)sub_18037F0C0(v20, v19, qword_1803B0530);
  if ( v22 == nullptr )
  {
    v12 = a1[12];
    v13 = 501;
    a1[12] = 0;
    goto LABEL_60;
  }
  v23 = sub_18037F0C0(v21, v6, qword_1803AFF88);
  v24 = (int *)v23;
  if ( v23 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    if ( *v22 >= 0 )
    {
      v8 = (*(_QWORD *)v22)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v22 + 1) + 48LL))(v22);
    }
    v13 = 501;
    goto LABEL_60;
  }
  v13 = 501;
  *(_DWORD *)(v3 + 40) = 501;
  v25 = (int *)sub_1803770D0(a1, v23, qword_1803B0258);
  if ( *v24 >= 0 )
  {
    v8 = (*(_QWORD *)v24)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v24 + 1) + 48LL))(v24);
  }
  if ( v25 == nullptr )
  {
    v12 = a1[12];
    a1[12] = 0;
    if ( *v22 < 0 )
      goto LABEL_60;
    v8 = (*(_QWORD *)v22)-- == 1;
    if ( !v8 )
      goto LABEL_60;
    v26 = *((_QWORD *)v22 + 1);
    v27 = v22;
LABEL_59:
    (*(void (__fastcall **)(int *))(v26 + 48))(v27);
    goto LABEL_60;
  }
  *(_DWORD *)(v3 + 40) = 501;
  v68 = v6;
  v69 = v25;
  v28 = sub_18037D030(a1, v22, &v68, &v69, qword_1803B0258);
  if ( *v22 >= 0 )
  {
    v8 = (*(_QWORD *)v22)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v22 + 1) + 48LL))(v22);
  }
  if ( *v25 >= 0 )
  {
    v8 = (*(_QWORD *)v25)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v25 + 1) + 48LL))(v25);
  }
  if ( v28 == 0 )
  {
    v12 = a1[12];
    a1[12] = 0;
    goto LABEL_60;
  }
  v29 = qword_1803B0120;
  v30 = (int *)v28;
  v70 = (int *)v28;
  v31 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v31, v29, Py_TrueStruct);
  v32 = qword_1803AFF90;
  v13 = 502;
  *(_DWORD *)(v3 + 40) = 502;
  v68 = (int *)v32;
  v33 = sub_18037E560(a1, v28, qword_1803B0118);
  v34 = (int *)v33;
  if ( v33 != 0 )
  {
    PyDict_SetItem(v31, v68, v33);
    if ( *v34 >= 0 )
    {
      v8 = (*(_QWORD *)v34)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v34 + 1) + 48LL))(v34);
    }
    goto LABEL_109;
  }
  v12 = a1[12];
  a1[12] = 0;
  if ( *v31 >= 0 )
  {
    v8 = (*(_QWORD *)v31)-- == 1;
    if ( v8 )
    {
      v26 = *((_QWORD *)v31 + 1);
      v27 = v31;
      goto LABEL_59;
    }
  }
LABEL_60:
  v35 = (int *)a1[31];
  v68 = nullptr;
  v69 = v35;
  if ( v35 != nullptr && *v35 != -1 )
    ++*v35;
  v36 = *(_DWORD **)(v12 + 40);
  v37 = sub_18038E2A0(v3, v13);
  v38 = (int *)v37;
  if ( v36 == nullptr )
  {
    v39 = *(int **)(v12 + 40);
    if ( v37 == Py_NoneStruct )
    {
      v37 = 0;
      goto LABEL_73;
    }
LABEL_70:
    if ( v37 != 0 && *(_DWORD *)v37 != -1 )
      ++*(_DWORD *)v37;
    goto LABEL_73;
  }
  *(_QWORD *)(v37 + 16) = v36;
  if ( *v36 != -1 )
    ++*v36;
  v39 = *(int **)(v12 + 40);
  if ( v37 != Py_NoneStruct )
    goto LABEL_70;
  v37 = 0;
LABEL_73:
  *(_QWORD *)(v12 + 40) = v37;
  if ( v39 != nullptr && *v39 >= 0 )
  {
    v8 = (*(_QWORD *)v39)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v39 + 1) + 48LL))();
  }
  if ( v38 != nullptr && *v38 >= 0 )
  {
    v8 = (*(_QWORD *)v38)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v38 + 1) + 48LL))(v38);
  }
  v40 = (int *)a1[31];
  a1[31] = v12;
  if ( v40 != nullptr && *v40 >= 0 )
  {
    v8 = (*(_QWORD *)v40)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v40 + 1) + 48LL))();
  }
  if ( (unsigned int)sub_18000CE90(a1, *(_QWORD *)(a1[31] + 8LL), PyExc_Exception) == 0 )
  {
    v53 = sub_180373290(a1, &v68);
    v46 = v68;
    v47 = 498;
    if ( v53 != 0 )
      v47 = 0;
    v54 = *((_QWORD *)v68 + 5);
    if ( v54 != 0 && *(_QWORD *)(v54 + 24) == v3 )
      *(_DWORD *)(v3 + 40) = *(_DWORD *)(v54 + 36);
LABEL_140:
    v55 = (int *)a1[31];
    a1[31] = v69;
    if ( v55 != nullptr && *v55 >= 0 )
    {
      v8 = (*(_QWORD *)v55)-- == 1;
      if ( v8 )
        (*(void (**)(void))(*((_QWORD *)v55 + 1) + 48LL))();
    }
    v56 = *((_QWORD *)v46 + 5);
    if ( v56 != 0 )
    {
      if ( *(_QWORD *)(v56 + 24) == v3 )
      {
LABEL_163:
        sub_18036D0A0(v3, "oooooo", v67);
        v62 = qword_1803AFB30;
        if ( v3 == qword_1803AFB30 )
        {
          if ( *(int *)qword_1803AFB30 >= 0 )
          {
            v8 = (*(_QWORD *)qword_1803AFB30)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64))(*(_QWORD *)(v62 + 8) + 48LL))(
                v62,
                *(_QWORD *)(*(_QWORD *)(v62 + 8) + 48LL),
                v60,
                v61);
          }
          qword_1803AFB30 = 0;
        }
        v63 = a1[7];
        v64 = *(_QWORD *)(*(_QWORD *)v63 + 48LL);
        *(_QWORD *)v63 = *(_QWORD *)(*(_QWORD *)v63 + 8LL);
        *(_DWORD *)(v64 + 64) = -1;
        v65 = *(int **)(v64 + 16);
        if ( v65 != nullptr )
        {
          *(_QWORD *)(v64 + 16) = 0;
          if ( *v65 >= 0 )
          {
            v8 = (*(_QWORD *)v65)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(int *, _QWORD, __int64, __int64))(*((_QWORD *)v65 + 1) + 48LL))(
                v65,
                *(_QWORD *)(*((_QWORD *)v65 + 1) + 48LL),
                v60,
                v61);
          }
        }
        if ( *(int *)v64 >= 0 )
        {
          v8 = (*(_QWORD *)v64)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(__int64))(*(_QWORD *)(v64 + 8) + 48LL))(v64);
        }
        *(_QWORD *)(v64 + 80) = 0;
        if ( v71 != nullptr && *v71 >= 0 )
        {
          v8 = (*(_QWORD *)v71)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v71 + 1) + 48LL))(v71);
        }
        if ( v7 != nullptr && *v7 >= 0 )
        {
          v8 = (*(_QWORD *)v7)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
        }
        if ( v70 != nullptr && *v70 >= 0 )
        {
          v8 = (*(_QWORD *)v70)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v70 + 1) + 48LL))(v70);
        }
        if ( *v67 >= 0 )
        {
          v8 = (*(_QWORD *)v67)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *, _QWORD, __int64, __int64))(*((_QWORD *)v67 + 1) + 48LL))(
              v67,
              *(_QWORD *)(*((_QWORD *)v67 + 1) + 48LL),
              v60,
              v61);
        }
        if ( *v6 >= 0 )
        {
          v8 = (*(_QWORD *)v6)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
        }
        v66 = (int *)a1[12];
        a1[12] = v46;
        if ( v66 != nullptr && *v66 >= 0 )
        {
          v8 = (*(_QWORD *)v66)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *, _QWORD, __int64, __int64))(*((_QWORD *)v66 + 1) + 48LL))(
              v66,
              *(_QWORD *)(*((_QWORD *)v66 + 1) + 48LL),
              v60,
              v61);
        }
        return nullptr;
      }
      v57 = sub_18038E2A0(v3, v47);
      v58 = (int *)v57;
      *(_QWORD *)(v57 + 16) = v56;
      if ( *(_DWORD *)v56 != -1 )
        ++*(_DWORD *)v56;
      v59 = *((int **)v46 + 5);
      if ( v57 == Py_NoneStruct )
      {
        v57 = 0;
LABEL_155:
        *((_QWORD *)v46 + 5) = v57;
        if ( v59 != nullptr && *v59 >= 0 )
        {
          v8 = (*(_QWORD *)v59)-- == 1;
          if ( v8 )
            (*(void (**)(void))(*((_QWORD *)v59 + 1) + 48LL))();
        }
        if ( v58 != nullptr && *v58 >= 0 )
        {
          v8 = (*(_QWORD *)v58)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v58 + 1) + 48LL))(v58);
        }
        goto LABEL_163;
      }
    }
    else
    {
      v57 = sub_18038E2A0(v3, v47);
      v58 = (int *)v57;
      v59 = *((int **)v46 + 5);
      if ( v57 == Py_NoneStruct )
      {
        v57 = 0;
        goto LABEL_155;
      }
    }
    if ( v58 != nullptr && *(_DWORD *)v57 != -1 )
      ++*(_DWORD *)v57;
    goto LABEL_155;
  }
  v41 = a1[31];
  if ( *(_DWORD *)v41 != -1 )
    ++*(_DWORD *)v41;
  v42 = qword_1803B0120;
  v31 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v31, v42, Py_FalseStruct);
  v43 = qword_1803B0128;
  v44 = sub_1803749A0(v41);
  v45 = (int *)v44;
  if ( v44 == 0 )
  {
    v46 = (int *)a1[12];
    a1[12] = 0;
    if ( *v31 >= 0 )
    {
      v8 = (*(_QWORD *)v31)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v31 + 1) + 48LL))(v31);
    }
    if ( v41 != 0 && *(int *)v41 >= 0 )
    {
      v8 = (*(_QWORD *)v41)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v41 + 8) + 48LL))(v41);
    }
    v47 = 504;
    goto LABEL_140;
  }
  PyDict_SetItem(v31, v43, v44);
  if ( *v45 >= 0 )
  {
    v8 = (*(_QWORD *)v45)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v45 + 1) + 48LL))(v45);
  }
  if ( v41 != 0 && *(int *)v41 >= 0 )
  {
    v8 = (*(_QWORD *)v41)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v41 + 8) + 48LL))(v41);
  }
  v48 = (int *)a1[31];
  a1[31] = v69;
  if ( v48 != nullptr && *v48 >= 0 )
  {
    v8 = (*(_QWORD *)v48)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v48 + 1) + 48LL))();
  }
  v30 = v70;
LABEL_109:
  v49 = a1[7];
  v50 = *(_QWORD *)(*(_QWORD *)v49 + 48LL);
  *(_QWORD *)v49 = *(_QWORD *)(*(_QWORD *)v49 + 8LL);
  *(_DWORD *)(v50 + 64) = -1;
  v51 = *(int **)(v50 + 16);
  if ( v51 != nullptr )
  {
    *(_QWORD *)(v50 + 16) = 0;
    if ( *v51 >= 0 )
    {
      v8 = (*(_QWORD *)v51)-- == 1;
      if ( v8 )
        (*(void (**)(void))(*((_QWORD *)v51 + 1) + 48LL))();
    }
  }
  if ( *(int *)v50 >= 0 )
  {
    v8 = (*(_QWORD *)v50)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v50 + 8) + 48LL))(v50);
  }
  *(_QWORD *)(v50 + 80) = 0;
  if ( v71 != nullptr && *v71 >= 0 )
  {
    v8 = (*(_QWORD *)v71)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v71 + 1) + 48LL))();
  }
  if ( v7 != nullptr && *v7 >= 0 )
  {
    v8 = (*(_QWORD *)v7)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
  }
  if ( v30 != nullptr && *v30 >= 0 )
  {
    v8 = (*(_QWORD *)v30)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v30 + 1) + 48LL))(v30);
  }
  if ( *v67 >= 0 )
  {
    v8 = (*(_QWORD *)v67)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v67 + 1) + 48LL))();
  }
  if ( *v6 >= 0 )
  {
    v8 = (*(_QWORD *)v6)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
  }
  return v31;
}
