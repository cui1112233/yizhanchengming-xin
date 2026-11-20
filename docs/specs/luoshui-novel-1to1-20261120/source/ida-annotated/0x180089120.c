// Analyst annotations only; original export: source/ida/0x180089120.c
// qword comments derive module constant slot + original file range; not runtime values.
// Original IDA line 1 | sub_180089120 -> NovelsApi.novels_save_book_rules
int *__fastcall sub_180089120(__int64 a1, __int64 a2, int **a3)
{
  __int64 v4; // r15
  int *v5; // r14
  int *v6; // r13
  int *v7; // r12
  bool v8; // zf
  unsigned int v9; // edi
  int v10; // edx
  int v11; // r8d
  __int64 v12; // rax
  int *v13; // rbx
  __int64 v14; // rax
  __int64 v15; // rax
  __int64 v16; // rbp
  __int64 v17; // rbp
  __int64 v18; // rax
  int *v19; // rbx
  __int64 v20; // rax
  int *v21; // r14
  __int64 v22; // rax
  int *v23; // rax
  __int64 v24; // rdx
  _DWORD *v25; // rcx
  _DWORD *v26; // r14
  __int64 v27; // rax
  int *v28; // rbx
  int *v29; // rcx
  int *v30; // rcx
  __int64 v31; // r14
  __int64 v32; // rdi
  int *v33; // rbp
  __int64 v34; // rdi
  __int64 v35; // rax
  int *v36; // rbx
  __int64 v37; // rbx
  unsigned int v38; // edi
  __int64 v39; // rdx
  int *v40; // rcx
  int *v41; // rax
  char v42; // al
  __int64 v43; // rax
  int *v44; // rcx
  __int64 v45; // r14
  __int64 v46; // rax
  int *v47; // rdi
  int *v48; // rcx
  __int64 v49; // rdx
  __int64 v50; // r8
  __int64 v51; // r9
  __int64 v52; // rcx
  int *v53; // rcx
  __int64 v55; // [rsp+20h] [rbp-B8h]
  __int64 v56; // [rsp+28h] [rbp-B0h]
  int *v57; // [rsp+50h] [rbp-88h]
  int *v58; // [rsp+58h] [rbp-80h]
  int *v59; // [rsp+60h] [rbp-78h]
  __int64 v60; // [rsp+68h] [rbp-70h]
  _DWORD *v61; // [rsp+68h] [rbp-70h]
  __int64 v62; // [rsp+70h] [rbp-68h] BYREF
  int *v63; // [rsp+78h] [rbp-60h]
  _QWORD v64[10]; // [rsp+88h] [rbp-50h] BYREF
  int *v65; // [rsp+F0h] [rbp+18h]
  int *v66; // [rsp+F8h] [rbp+20h]

  v4 = qword_1803AFE88;
  v5 = a3[1];
  v66 = *a3;
  v59 = a3[2];
  v58 = v5;
  v6 = nullptr;
  v65 = nullptr;
  v7 = nullptr;
  v57 = nullptr;
  if ( qword_1803AFE88 == 0 )
  {
LABEL_8:
    v4 = sub_18036CD10(a1, *(_QWORD *)&qword_1803AFD80, qword_1803CA828, 64);
    qword_1803AFE88 = v4;
    goto LABEL_9;
  }
  if ( *(__int64 *)qword_1803AFE88 > 1 || *(_QWORD *)(qword_1803AFE88 + 16) != 0 )
  {
    if ( qword_1803AFE88 != 0 && *(int *)qword_1803AFE88 >= 0 )
    {
      v8 = (*(_QWORD *)qword_1803AFE88)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v4 + 8) + 48LL))(v4);
    }
    goto LABEL_8;
  }
LABEL_9:
  sub_180001D00(a1, v4);
  v9 = 674;
// Original IDA line 95 | qword_1803B0678 -> novels.api.novels_api#229 bytes [3928676, 3928695]: "models.book_rules"
  v10 = qword_1803B0678;
  v11 = qword_1803CA830;
// Original IDA line 97 | qword_1803B0288 -> novels.api.novels_api#103 bytes [3926234, 3926236]: 2
  v56 = qword_1803B0288;
// Original IDA line 98 | qword_1803B0698 -> novels.api.novels_api#233 bytes [3928742, 3928772]: {"type": "tuple", "items": [{"tag": "a", "fileRange": [3928744, 3928755], "value": "BookRules"}, {"tag": "a", "fileRange": [3928755, 3928772], "value": "save_book_rules"}]}
  v55 = qword_1803B0698;
  *(_DWORD *)(v4 + 40) = 674;
  v12 = sub_1803717E0(a1, v10, v11, Py_NoneStruct, v55, v56);
  v13 = (int *)v12;
  if ( v12 == 0 )
  {
LABEL_33:
    v17 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    goto LABEL_34;
  }
  if ( *(_QWORD *)(v12 + 8) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 110 | qword_1803B06A0 -> novels.api.novels_api#234 bytes [3928772, 3928783]: "BookRules"
    v14 = sub_180371D80(a1, (_DWORD)v13, qword_1803CA830, qword_1803B06A0, qword_1803AB250);
  else
// Original IDA line 112 | qword_1803B06A0 -> novels.api.novels_api#234 bytes [3928772, 3928783]: "BookRules"
    v14 = sub_180371890(a1, v13, qword_1803B06A0);
  v60 = v14;
  if ( v14 != 0 )
  {
    v65 = (int *)v14;
    if ( *((_QWORD *)v13 + 1) == PyModule_Type || (unsigned int)PyType_IsSubtype() != 0 )
// Original IDA line 118 | qword_1803B06A8 -> novels.api.novels_api#235 bytes [3928783, 3928800]: "save_book_rules"
      v15 = sub_180371D80(a1, (_DWORD)v13, qword_1803CA830, qword_1803B06A8, qword_1803AB250);
    else
// Original IDA line 120 | qword_1803B06A8 -> novels.api.novels_api#235 bytes [3928783, 3928800]: "save_book_rules"
      v15 = sub_180371890(a1, v13, qword_1803B06A8);
    v16 = v15;
    if ( v15 != 0 )
    {
      v6 = (int *)v15;
      if ( *v13 >= 0 )
      {
        v8 = (*(_QWORD *)v13)-- == 1;
        if ( v8 )
          (*(void (__fastcall **)(int *))(*((_QWORD *)v13 + 1) + 48LL))(v13);
      }
      v9 = 675;
      *(_DWORD *)(v4 + 40) = 675;
// Original IDA line 133 | qword_1803B0058 -> novels.api.novels_api#33 bytes [3924931, 3924939]: "_state"
      v18 = sub_18037E560(a1, v66, *(_QWORD *)&qword_1803B0058);
      v19 = (int *)v18;
      if ( v18 != 0 )
      {
        *(_DWORD *)(v4 + 40) = 675;
// Original IDA line 138 | qword_1803B0690 -> novels.api.novels_api#232 bytes [3928731, 3928742]: "story_dir"
        v20 = sub_18037E6C0(a1, v18, qword_1803B0690, v5);
        v62 = v20;
        if ( *v19 >= 0 )
        {
          v8 = (*(_QWORD *)v19)-- == 1;
          if ( v8 )
          {
            (*(void (__fastcall **)(int *))(*((_QWORD *)v19 + 1) + 48LL))(v19);
            v20 = v62;
          }
        }
        if ( v20 != 0 )
        {
          v9 = 676;
          *(_DWORD *)(v4 + 40) = 676;
          v21 = (int *)v20;
          v57 = (int *)v20;
// Original IDA line 155 | qword_1803B06B0 -> novels.api.novels_api#236 bytes [3928800, 3928811]: "from_dict"
          v22 = sub_18037E6C0(a1, v60, qword_1803B06B0, v59);
          if ( v22 != 0 )
          {
            v64[0] = v62;
            v9 = 677;
            *(_DWORD *)(v4 + 40) = 677;
            v64[1] = v22;
            v7 = (int *)v22;
            v23 = (int *)sub_180377750(a1, v16, v64);
            if ( v23 != nullptr )
            {
              if ( *v23 >= 0 )
              {
                v8 = (*(_QWORD *)v23)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v23 + 1) + 48LL))(v23);
              }
              sub_180001D80(a1, v24);
// Original IDA line 173 | qword_1803B06B8 -> novels.api.novels_api#237 bytes [3928811, 3928858]: {"type": "dict", "keys": [{"tag": "a", "fileRange": [3928813, 3928822], "value": "success"}, {"tag": "a", "fileRange": [3928822, 3928831], "value": "message"}], "values": [{"tag": "t", "fileRange": [3928831, 3928832], "value": true}, {"tag"... (truncated literal)
              v33 = (int *)sub_18036FB80(a1, qword_1803B06B8);
              v41 = v65;
LABEL_151:
              if ( *v41 >= 0 )
              {
                v8 = (*(_QWORD *)v41)-- == 1;
                if ( v8 )
                  (*(void (__fastcall **)(int *))(*((_QWORD *)v41 + 1) + 48LL))(v41);
              }
              goto LABEL_154;
            }
          }
        }
      }
      goto LABEL_33;
    }
  }
  v17 = *(_QWORD *)(a1 + 96);
  *(_QWORD *)(a1 + 96) = 0;
  if ( *v13 >= 0 )
  {
    v8 = (*(_QWORD *)v13)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v13 + 1) + 48LL))(v13);
  }
LABEL_34:
  v25 = *(_DWORD **)(a1 + 248);
  v63 = v65;
  v62 = 0;
  v61 = v25;
  if ( v25 != nullptr && *v25 != -1 )
    ++*v25;
  v26 = *(_DWORD **)(v17 + 40);
  v27 = sub_18038E2A0(v4, v9);
  v28 = (int *)v27;
  if ( v26 != nullptr )
  {
    *(_QWORD *)(v27 + 16) = v26;
    if ( *v26 != -1 )
      ++*v26;
    v29 = *(int **)(v17 + 40);
    if ( v27 == Py_NoneStruct )
    {
      v27 = 0;
      goto LABEL_47;
    }
  }
  else
  {
    v29 = *(int **)(v17 + 40);
    if ( v27 == Py_NoneStruct )
    {
      v27 = 0;
      goto LABEL_47;
    }
  }
  if ( v27 != 0 && *(_DWORD *)v27 != -1 )
    ++*(_DWORD *)v27;
LABEL_47:
  *(_QWORD *)(v17 + 40) = v27;
  if ( v29 != nullptr && *v29 >= 0 )
  {
    v8 = (*(_QWORD *)v29)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v29 + 1) + 48LL))();
  }
  if ( v28 != nullptr && *v28 >= 0 )
  {
    v8 = (*(_QWORD *)v28)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v28 + 1) + 48LL))(v28);
  }
  v30 = *(int **)(a1 + 248);
  *(_QWORD *)(a1 + 248) = v17;
  if ( v30 != nullptr && *v30 >= 0 )
  {
    v8 = (*(_QWORD *)v30)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v30 + 1) + 48LL))();
  }
  if ( (unsigned int)sub_18000CE90(a1, *(_QWORD *)(*(_QWORD *)(a1 + 248) + 8LL), PyExc_Exception) == 0 )
  {
    v42 = sub_180373290(a1, &v62);
    v37 = v62;
    v38 = 673;
    if ( v42 != 0 )
      v38 = 0;
    v43 = *(_QWORD *)(v62 + 40);
    if ( v43 != 0 && *(_QWORD *)(v43 + 24) == v4 )
      *(_DWORD *)(v4 + 40) = *(_DWORD *)(v43 + 36);
LABEL_89:
    v44 = *(int **)(a1 + 248);
    *(_QWORD *)(a1 + 248) = v61;
    if ( v44 != nullptr && *v44 >= 0 )
    {
      v8 = (*(_QWORD *)v44)-- == 1;
      if ( v8 )
        (*(void (**)(void))(*((_QWORD *)v44 + 1) + 48LL))();
    }
    v45 = *(_QWORD *)(v37 + 40);
    if ( v45 != 0 )
    {
      if ( *(_QWORD *)(v45 + 24) == v4 )
      {
LABEL_112:
        sub_18036D0A0(v4, "oooooooo", v66);
        v52 = qword_1803AFE88;
        if ( v4 == qword_1803AFE88 )
        {
          if ( *(int *)qword_1803AFE88 >= 0 )
          {
            v8 = (*(_QWORD *)qword_1803AFE88)-- == 1;
            if ( v8 )
              (*(void (__fastcall **)(__int64, _QWORD, __int64, __int64, int *, int *))(*(_QWORD *)(v52 + 8) + 48LL))(
                v52,
                *(_QWORD *)(*(_QWORD *)(v52 + 8) + 48LL),
                v50,
                v51,
                v59,
                v65);
          }
          qword_1803AFE88 = 0;
        }
        sub_180001D80(a1, v49);
        if ( v65 != nullptr && *v65 >= 0 )
        {
          v8 = (*(_QWORD *)v65)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v65 + 1) + 48LL))(v65);
        }
        if ( v6 != nullptr && *v6 >= 0 )
        {
          v8 = (*(_QWORD *)v6)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
        }
        if ( v57 != nullptr && *v57 >= 0 )
        {
          v8 = (*(_QWORD *)v57)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v57 + 1) + 48LL))(v57);
        }
        if ( v7 != nullptr && *v7 >= 0 )
        {
          v8 = (*(_QWORD *)v7)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
        }
        if ( *v66 >= 0 )
        {
          v8 = (*(_QWORD *)v66)-- == 1;
          if ( v8 )
            (*(void (**)(void))(*((_QWORD *)v66 + 1) + 48LL))();
        }
        if ( *v58 >= 0 )
        {
          v8 = (*(_QWORD *)v58)-- == 1;
          if ( v8 )
            (*(void (**)(void))(*((_QWORD *)v58 + 1) + 48LL))();
        }
        if ( *v59 >= 0 )
        {
          v8 = (*(_QWORD *)v59)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v59 + 1) + 48LL))(v59);
        }
        v53 = *(int **)(a1 + 96);
        *(_QWORD *)(a1 + 96) = v37;
        if ( v53 != nullptr && *v53 >= 0 )
        {
          v8 = (*(_QWORD *)v53)-- == 1;
          if ( v8 )
            (*(void (**)(void))(*((_QWORD *)v53 + 1) + 48LL))();
        }
        return nullptr;
      }
      v46 = sub_18038E2A0(v4, v38);
      v47 = (int *)v46;
      *(_QWORD *)(v46 + 16) = v45;
      if ( *(_DWORD *)v45 != -1 )
        ++*(_DWORD *)v45;
      v48 = *(int **)(v37 + 40);
      if ( v46 == Py_NoneStruct )
      {
        v46 = 0;
LABEL_104:
        *(_QWORD *)(v37 + 40) = v46;
        if ( v48 != nullptr && *v48 >= 0 )
        {
          v8 = (*(_QWORD *)v48)-- == 1;
          if ( v8 )
            (*(void (**)(void))(*((_QWORD *)v48 + 1) + 48LL))();
        }
        if ( v47 != nullptr && *v47 >= 0 )
        {
          v8 = (*(_QWORD *)v47)-- == 1;
          if ( v8 )
            (*(void (__fastcall **)(int *))(*((_QWORD *)v47 + 1) + 48LL))(v47);
        }
        goto LABEL_112;
      }
    }
    else
    {
      v46 = sub_18038E2A0(v4, v38);
      v47 = (int *)v46;
      v48 = *(int **)(v37 + 40);
      if ( v46 == Py_NoneStruct )
      {
        v46 = 0;
        goto LABEL_104;
      }
    }
    if ( v47 != nullptr && *(_DWORD *)v46 != -1 )
      ++*(_DWORD *)v46;
    goto LABEL_104;
  }
  v31 = *(_QWORD *)(a1 + 248);
  if ( *(_DWORD *)v31 != -1 )
    ++*(_DWORD *)v31;
// Original IDA line 393 | qword_1803B0120 -> novels.api.novels_api#58 bytes [3925459, 3925468]: "success"
  v32 = qword_1803B0120;
  v33 = (int *)PyDict_NewPresized(2);
  PyDict_SetItem(v33, v32, Py_FalseStruct);
// Original IDA line 396 | qword_1803B0128 -> novels.api.novels_api#59 bytes [3925468, 3925477]: "message"
  v34 = qword_1803B0128;
  v35 = sub_1803749A0(v31);
  v36 = (int *)v35;
  if ( v35 == 0 )
  {
    v37 = *(_QWORD *)(a1 + 96);
    *(_QWORD *)(a1 + 96) = 0;
    if ( *v33 >= 0 )
    {
      v8 = (*(_QWORD *)v33)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(int *))(*((_QWORD *)v33 + 1) + 48LL))(v33);
    }
    if ( v31 != 0 && *(int *)v31 >= 0 )
    {
      v8 = (*(_QWORD *)v31)-- == 1;
      if ( v8 )
        (*(void (__fastcall **)(__int64))(*(_QWORD *)(v31 + 8) + 48LL))(v31);
    }
    v38 = 680;
    goto LABEL_89;
  }
  PyDict_SetItem(v33, v34, v35);
  if ( *v36 >= 0 )
  {
    v8 = (*(_QWORD *)v36)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v36 + 1) + 48LL))(v36);
  }
  if ( v31 != 0 && *(int *)v31 >= 0 )
  {
    v8 = (*(_QWORD *)v31)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(__int64))(*(_QWORD *)(v31 + 8) + 48LL))(v31);
  }
  v40 = *(int **)(a1 + 248);
  *(_QWORD *)(a1 + 248) = v61;
  if ( v40 != nullptr && *v40 >= 0 )
  {
    v8 = (*(_QWORD *)v40)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v40 + 1) + 48LL))();
  }
  sub_180001D80(a1, v39);
  v21 = v57;
  v41 = v63;
  if ( v63 != nullptr )
    goto LABEL_151;
LABEL_154:
  if ( v6 != nullptr && *v6 >= 0 )
  {
    v8 = (*(_QWORD *)v6)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v6 + 1) + 48LL))(v6);
  }
  if ( v21 != nullptr && *v21 >= 0 )
  {
    v8 = (*(_QWORD *)v21)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v21 + 1) + 48LL))(v21);
  }
  if ( v7 != nullptr && *v7 >= 0 )
  {
    v8 = (*(_QWORD *)v7)-- == 1;
    if ( v8 )
      (*(void (__fastcall **)(int *))(*((_QWORD *)v7 + 1) + 48LL))(v7);
  }
  if ( *v66 >= 0 )
  {
    v8 = (*(_QWORD *)v66)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v66 + 1) + 48LL))();
  }
  if ( *v58 >= 0 )
  {
    v8 = (*(_QWORD *)v58)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v58 + 1) + 48LL))();
  }
  if ( *v59 >= 0 )
  {
    v8 = (*(_QWORD *)v59)-- == 1;
    if ( v8 )
      (*(void (**)(void))(*((_QWORD *)v59 + 1) + 48LL))();
  }
  return v33;
}
