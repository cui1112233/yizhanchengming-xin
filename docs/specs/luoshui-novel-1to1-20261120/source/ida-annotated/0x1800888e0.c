// Analyst annotations only; original export: source/ida/0x1800888e0.c
// qword comments derive module constant slot + original file range; not runtime values.
// Original IDA line 1 | sub_1800888E0 -> NovelsApi.novels_get_book_rules
int *__fastcall sub_1800888E0(_QWORD *a1, __int64 a2, int **a3)
{
  __int64 v3; // r15
  int *v5; // rbp
  int *v6; // rdi
  int *v7; // r12
  int *v8; // r13
  bool v9; // zf
  unsigned int v10; // r14d
  int v11; // edx
  int v12; // r8d
  __int64 v13; // rax
  int *v14; // rbx
  __int64 v15; // rdi
  __int64 v16; // rax
  __int64 v17; // rax
  int *v18; // rbx
  __int64 v19; // rdi
  int *v20; // rdx
  __int64 v21; // rax
  __int64 v22; // rdi
  int *v23; // rbp
  __int64 v24; // rdi
  __int64 v25; // rax
  int *v26; // rbx
  _DWORD *v27; // rcx
  _DWORD *v28; // rbp
  __int64 v29; // rax
  int *v30; // rbx
  int *v31; // rcx
  int *v32; // rcx
  __int64 v33; // r14
  __int64 v34; // rdi
  __int64 v35; // rdi
  __int64 v36; // rax
  int *v37; // rbx
  int *v38; // rbx
  unsigned int v39; // edi
  int *v40; // rcx
  __int64 v41; // rax
  char v42; // al
  __int64 v43; // rax
  int *v44; // rcx
  __int64 v45; // r14
  __int64 v46; // rax
  int *v47; // rdi
  int *v48; // rcx
  int *v49; // rdi
  __int64 v50; // rdx
  __int64 v51; // r8
  __int64 v52; // r9
  __int64 v53; // rcx
  __int64 v54; // r8
  __int64 v55; // r9
  int *v56; // rcx
  __int64 v58; // rcx
  __int64 v59; // rbx
  int *v60; // rcx
  int *v61; // rcx
  __int64 v62; // [rsp+20h] [rbp-78h]
  __int64 v63; // [rsp+28h] [rbp-70h]
  _DWORD *v64; // [rsp+40h] [rbp-58h]
  int *v65; // [rsp+48h] [rbp-50h]
  int *v66; // [rsp+50h] [rbp-48h]
  int *v67; // [rsp+B0h] [rbp+18h] BYREF
  int *v68; // [rsp+B8h] [rbp+20h]

  v3 = qword_1803AFE78;
  v5 = *a3;
  v6 = a3[1];
  v7 = nullptr;
  v65 = *a3;
  v8 = nullptr;
  v66 = v6;
  v68 = nullptr;
  if ( qword_1803AFE78 == 0 )
  {
LABEL_8:
    v3 = sub_18036CD10(a1, *(_QWORD *)&qword_1803AFBE0, qword_1803CA828, 48);
    qword_1803AFE78 = v3;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803AFE78 > 1 || *(_QWORD *)(qword_1803AFE78 + 16) != 0 )
  {
    if ( qword_1803AFE78 != 0 && *(int *)qword_1803AFE78 >= 0 )
    {
      v9 = (*(_QWORD *)qword_1803AFE78)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v3 + 8) + 48LL))(v3);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v3);
  v10 = 665;
// Original IDA line 96 | qword_1803B0678 -> novels.api.novels_api#229 bytes [3928676, 3928695]: "models.book_rules"
  v11 = qword_1803B0678;
  v12 = qword_1803CA830;
// Original IDA line 98 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v63 = qword_1803B0288;
// Original IDA line 99 | qword_1803B0680 -> novels.api.novels_api#230 bytes [3928695, 3928714]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928697, 3928714], "value": "load_book_rules"}]}
  v62 = qword_1803B0680;
  *(_DWORD *)(v3 + 40) = 665;
  v13 = sub_1803717E0((_DWORD)a1, v11, v12, Py_NoneStruct, v62, v63);
  v14 = (int *)v13;
  if ( v13 != 0 )
  {
    if ( *(_QWORD *)(v13 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 106 | qword_1803B0688 -> novels.api.novels_api#231 bytes [3928714, 3928731]: "load_book_rules"
      v16 = sub_180371D80((_DWORD)a1, (_DWORD)v14, qword_1803CA830, qword_1803B0688, qword_1803AB250);
    else
// Original IDA line 108 | qword_1803B0688 -> novels.api.novels_api#231 bytes [3928714, 3928731]: "load_book_rules"
      v16 = sub_180371890(a1, v14, qword_1803B0688);
    v67 = (int *)v16;
    if ( *v14 >= 0 )
    {
      v9 = (*(_QWORD *)v14)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v14 + 1) + 48LL))(v14);
    }
    if ( v67 != nullptr )
    {
      v10 = 666;
      v68 = v67;
      *(_DWORD *)(v3 + 40) = 666;
// Original IDA line 121 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
      v17 = sub_18037E560(a1, v5, *(_QWORD *)&qword_1803B0058);
      v18 = (int *)v17;
      if ( v17 != 0 )
      {
        *(_DWORD *)(v3 + 40) = 666;
// Original IDA line 126 | qword_1803B0690 -> novels.api.novels_api#232 bytes [3928731, 3928742]: "story_dir"
        v19 = sub_18037E6C0(a1, v17, qword_1803B0690, v6);
        if ( *v18 >= 0 )
        {
          v9 = (*(_QWORD *)v18)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v18 + 1) + 48LL))(v18);
        }
        if ( v19 != 0 )
        {
          v20 = v67;
          v10 = 667;
          *(_DWORD *)(v3 + 40) = 667;
          v7 = (int *)v19;
          v21 = sub_180376770(a1, v20, v19);
          if ( v21 != 0 )
          {
// Original IDA line 142 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
            v22 = qword_1803B0120;
            v8 = (int *)v21;
            v23 = (int *)PyDict_NewPresized(2);
            PyDict_SetItem(v23, v22, Py_TrueStruct);
// Original IDA line 146 | qword_1803AFF90 -> novels.api.novels_api#8 bytes [3924535, 3924541]: "data"
            v24 = qword_1803AFF90;
            v10 = 668;
            *(_DWORD *)(v3 + 40) = 668;
// Original IDA line 149 | qword_1803B0118 -> novels.api.novels_api#57 bytes [3925450, 3925459]: "to_dict"
            v25 = sub_18037E560(a1, v8, qword_1803B0118);
            v26 = (int *)v25;
            if ( v25 != 0 )
            {
              PyDict_SetItem(v23, v24, v25);
              if ( *v26 < 0 )
                goto LABEL_141;
              v9 = (*(_QWORD *)v26)-- == 1;
              if ( !v9 )
                goto LABEL_141;
              v41 = *((_QWORD *)v26 + 1);
              v40 = v26;
              goto LABEL_140;
            }
            v15 = a1[12];
            a1[12] = 0;
            if ( *v23 >= 0 )
            {
              v9 = (*(_QWORD *)v23)-- == 1;
              if ( v9 )
                (*(void (__fastcall **)(int *))(*((_QWORD *)v23 + 1) + 48LL))(v23);
            }
          }
          else
          {
            v15 = a1[12];
            a1[12] = 0;
          }
        }
        else
        {
          v15 = a1[12];
          a1[12] = 0;
        }
      }
      else
      {
        v15 = a1[12];
        a1[12] = 0;
      }
    }
    else
    {
      v15 = a1[12];
      a1[12] = 0;
    }
  }
  else
  {
    v15 = a1[12];
    a1[12] = 0;
  }
  v27 = (_DWORD *)a1[31];
  v67 = nullptr;
  v64 = v27;
  if ( v27 != nullptr && *v27 != -1 )
    ++*v27;
  v28 = *(_DWORD **)(v15 + 40);
  v29 = sub_18038E2A0(v3, v10);
  v30 = (int *)v29;
  if ( v28 != nullptr )
  {
    *(_QWORD *)(v29 + 16) = v28;
    if ( *v28 != -1 )
      ++*v28;
    v31 = *(int **)(v15 + 40);
    if ( v29 == Py_NoneStruct )
    {
      v29 = 0;
      goto LABEL_46;
    }
  }
  else
  {
    v31 = *(int **)(v15 + 40);
    if ( v29 == Py_NoneStruct )
    {
      v29 = 0;
      goto LABEL_46;
    }
  }
  if ( v29 != 0 && *(_DWORD *)v29 != -1 )
    ++*(_DWORD *)v29;
LABEL_46:
  *(_QWORD *)(v15 + 40) = v29;
  if ( v31 != nullptr && *v31 >= 0 )
  {
    v9 = (*(_QWORD *)v31)-- == 1;
    if ( v9 )
      (*(void (**)(void))(*((_QWORD *)v31 + 1) + 48LL))();
  }
  if ( v30 != nullptr && *v30 >= 0 )
  {
    v9 = (*(_QWORD *)v30)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v30 + 1) + 48LL))(v30);
  }
  v32 = (int *)a1[31];
  a1[31] = v15;
  if ( v32 != nullptr && *v32 >= 0 )
  {
    v9 = (*(_QWORD *)v32)-- == 1;
    if ( v9 )
      (*(void (**)(void))(*((_QWORD *)v32 + 1) + 48LL))();
  }
  if ( (unsigned int)sub_18000CE90(a1, *(_QWORD *)(a1[31] + 8LL), PyExc_Exception) == 0 )
  {
    v42 = sub_180373290(a1, &v67);
    v38 = v67;
    v39 = 664;
    if ( v42 != 0 )
      v39 = 0;
    v43 = *((_QWORD *)v67 + 5);
    if ( v43 != 0 && *(_QWORD *)(v43 + 24) == v3 )
      *(_DWORD *)(v3 + 40) = *(_DWORD *)(v43 + 36);
LABEL_86:
    v44 = (int *)a1[31];
    a1[31] = v64;
    if ( v44 != nullptr && *v44 >= 0 )
    {
      v9 = (*(_QWORD *)v44)-- == 1;
      if ( v9 )
        (*(void (**)(void))(*((_QWORD *)v44 + 1) + 48LL))();
    }
    v45 = *((_QWORD *)v38 + 5);
    if ( v45 != 0 )
    {
      if ( *(_QWORD *)(v45 + 24) == v3 )
      {
LABEL_109:
        v49 = v68;
        sub_18036D0A0(v3, "oooooo", v65);
        v53 = qword_1803AFE78;
        if ( v3 == qword_1803AFE78 )
        {
          if ( *(int *)qword_1803AFE78 >= 0 )
          {
            v9 = (*(_QWORD *)qword_1803AFE78)-- == 1;
            if ( v9 )
              (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64))(*(_QWORD *)(v53 + 8) + 48LL))(
                v53,
                *(_QWORD *)(*(_QWORD *)(v53 + 8) + 48LL),
                v51,
                v52);
          }
          qword_1803AFE78 = 0;
        }
        sub_180001D80(a1, v50);
        if ( v49 != nullptr && *v49 >= 0 )
        {
          v9 = (*(_QWORD *)v49)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v49 + 1) + 48LL))(v49);
        }
        if ( v7 != nullptr && *v7 >= 0 )
        {
          v9 = (*(_QWORD *)v7)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
        }
        if ( v8 != nullptr && *v8 >= 0 )
        {
          v9 = (*(_QWORD *)v8)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v8 + 1) + 48LL))(v8);
        }
        if ( *v65 >= 0 )
        {
          v9 = (*(_QWORD *)v65)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v65 + 1) + 48LL))(v65);
        }
        if ( *v66 >= 0 )
        {
          v9 = (*(_QWORD *)v66)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v66 + 1) + 48LL))(v66);
        }
        v56 = (int *)a1[12];
        a1[12] = v38;
        if ( v56 != nullptr && *v56 >= 0 )
        {
          v9 = (*(_QWORD *)v56)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *, _QWORD, __int64, __int64))(*((_QWORD *)v56 + 1) + 48LL))(
              v56,
              *(_QWORD *)(*((_QWORD *)v56 + 1) + 48LL),
              v54,
              v55);
        }
        return nullptr;
      }
      v46 = sub_18038E2A0(v3, v39);
      v47 = (int *)v46;
      *(_QWORD *)(v46 + 16) = v45;
      if ( *(_DWORD *)v45 != -1 )
        ++*(_DWORD *)v45;
      v48 = *((int **)v38 + 5);
      if ( v46 == Py_NoneStruct )
      {
        v46 = 0;
LABEL_101:
        *((_QWORD *)v38 + 5) = v46;
        if ( v48 != nullptr && *v48 >= 0 )
        {
          v9 = (*(_QWORD *)v48)-- == 1;
          if ( v9 )
            (*(void (**)(void))(*((_QWORD *)v48 + 1) + 48LL))();
        }
        if ( v47 != nullptr && *v47 >= 0 )
        {
          v9 = (*(_QWORD *)v47)-- == 1;
          if ( v9 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v47 + 1) + 48LL))(v47);
        }
        goto LABEL_109;
      }
    }
    else
    {
      v46 = sub_18038E2A0(v3, v39);
      v47 = (int *)v46;
      v48 = *((int **)v38 + 5);
      if ( v46 == Py_NoneStruct )
      {
        v46 = 0;
        goto LABEL_101;
      }
    }
    if ( v47 != nullptr && *(_DWORD *)v46 != -1 )
      ++*(_DWORD *)v46;
    goto LABEL_101;
  }
  v33 = a1[31];
  if ( *(_DWORD *)v33 != -1 )
    ++*(_DWORD *)v33;
// Original IDA line 385 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
  v34 = qword_1803B0120;
  v23 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v23, v34, Py_FalseStruct);
// Original IDA line 388 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
  v35 = qword_1803B0128;
  v36 = sub_1803749A0(v33);
  v37 = (int *)v36;
  if ( v36 == 0 )
  {
    v38 = (int *)a1[12];
    a1[12] = 0;
    if ( *v23 >= 0 )
    {
      v9 = (*(_QWORD *)v23)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v23 + 1) + 48LL))(v23);
    }
    if ( v33 != 0 && *(int *)v33 >= 0 )
    {
      v9 = (*(_QWORD *)v33)-- == 1;
      if ( v9 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v33 + 8) + 48LL))(v33);
    }
    v39 = 670;
    goto LABEL_86;
  }
  PyDict_SetItem(v23, v35, v36);
  if ( *v37 >= 0 )
  {
    v9 = (*(_QWORD *)v37)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v37 + 1) + 48LL))(v37);
  }
  if ( v33 != 0 && *(int *)v33 >= 0 )
  {
    v9 = (*(_QWORD *)v33)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v33 + 8) + 48LL))(v33);
  }
  v40 = (int *)a1[31];
  a1[31] = v64;
  if ( v40 != nullptr && *v40 >= 0 )
  {
    v9 = (*(_QWORD *)v40)-- == 1;
    if ( v9 )
    {
      v41 = *((_QWORD *)v40 + 1);
LABEL_140:
      (*(void (__fastcall **)(int *))(v41 + 48))(v40);
    }
  }
LABEL_141:
  v58 = a1[7];
  v59 = *(_QWORD *)(*(_QWORD *)v58 + 48LL);
  *(_QWORD *)v58 = *(_QWORD *)(*(_QWORD *)v58 + 8LL);
  *(_DWORD *)(v59 + 64) = -1;
  v60 = *(int **)(v59 + 16);
  if ( v60 != nullptr )
  {
    *(_QWORD *)(v59 + 16) = 0;
    if ( *v60 >= 0 )
    {
      v9 = (*(_QWORD *)v60)-- == 1;
      if ( v9 )
        (*(void (**)(void))(*((_QWORD *)v60 + 1) + 48LL))();
    }
  }
  if ( *(int *)v59 >= 0 )
  {
    v9 = (*(_QWORD *)v59)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v59 + 8) + 48LL))(v59);
  }
  *(_QWORD *)(v59 + 80) = 0;
  v61 = v68;
  if ( v68 != nullptr && *v68 >= 0 )
  {
    v9 = (*(_QWORD *)v68)-- == 1;
    if ( v9 )
      (*(void (**)(void))(*((_QWORD *)v61 + 1) + 48LL))();
  }
  if ( v7 != nullptr && *v7 >= 0 )
  {
    v9 = (*(_QWORD *)v7)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
  }
  if ( v8 != nullptr && *v8 >= 0 )
  {
    v9 = (*(_QWORD *)v8)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v8 + 1) + 48LL))(v8);
  }
  if ( *v65 >= 0 )
  {
    v9 = (*(_QWORD *)v65)-- == 1;
    if ( v9 )
      (*(void (**)(void))(*((_QWORD *)v65 + 1) + 48LL))();
  }
  if ( *v66 >= 0 )
  {
    v9 = (*(_QWORD *)v66)-- == 1;
    if ( v9 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v66 + 1) + 48LL))(v66);
  }
  return v23;
}
