/* REA ev_fad984d2e37e65f7f3b0d97197106a716dce83d13967f8a01b1a37f70a79eba2; original SHA b5b22086...; constant comments are analyst annotations. */

int * FUN_1801edcc0(longlong param_1,int *param_2,undefined8 param_3)

{
  int *piVar1;
  undefined8 uVar2;
  undefined8 uVar3;
  undefined8 uVar4;
  int *piVar5;
  char cVar6;
  undefined8 *puVar7;
  undefined8 uVar8;
  longlong lVar9;
  longlong *plVar10;
  longlong lVar11;
  code *pcVar12;
  code *pcVar13;
  longlong *plVar14;
  longlong *plVar15;
  longlong *plVar16;
  code *pcVar17;
  code *pcVar18;
  int iVar19;
  code *pcVar20;
  code *pcVar21;
  bool bVar22;
  code *local_res10 [2];
  longlong *local_res20;
  undefined8 in_stack_ffffffffffffff48;
  undefined4 uVar23;
  longlong *local_78;
  longlong *local_60;
  code *local_58;
  
  uVar23 = (undefined4)((ulonglong)in_stack_ffffffffffffff48 >> 0x20);
  DAT_1803ca5d8 = *(undefined8 *)(param_2 + 4);
  DAT_1803ca5e0 = param_2;
  if (DAT_1803b6c2b == '\0') {
    if (DAT_1803b6c2a == '\0') {
/* novels.prompts.inkos.writer_prompts#0="language" */
      FUN_180388a30(param_1,&DAT_1803bd150,"novels.prompts.inkos.writer_prompts");
      DAT_1803b6c2a = '\x01';
    }
    FUN_1801df300();
    DAT_1803b6c2b = '\x01';
  }
  uVar2 = DAT_1803ca5d8;
  plVar10 = DAT_1803ab560;
  uVar8 = DAT_1803ab330;
  puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803ab330);
  if (puVar7 == (undefined8 *)0x0) {
LAB_1801edd7e:
    PyDict_SetItem(uVar2,uVar8,plVar10);
  }
  else {
    plVar14 = (longlong *)*puVar7;
    if (plVar10 != plVar14) {
      if (plVar14 == (longlong *)0x0) goto LAB_1801edd7e;
      iVar19 = (int)*plVar10 + 1;
      if (iVar19 != 0) {
        *(int *)plVar10 = iVar19;
      }
      *puVar7 = plVar10;
      if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
        (**(code **)(plVar14[1] + 0x30))(plVar14);
      }
    }
  }
  puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803ab2a8);
  pcVar20 = (code *)0x0;
  pcVar21 = pcVar20;
  if (puVar7 != (undefined8 *)0x0) {
    pcVar21 = (code *)*puVar7;
  }
  uVar8 = PyUnicode_GetLength(pcVar21);
  lVar9 = PyUnicode_Find(pcVar21,DAT_1803ab4a8,0,uVar8,CONCAT44(uVar23,0xffffffff));
  if (lVar9 != -1) {
    PyUnicode_Substring(pcVar21,0,lVar9);
    FUN_1801ddb80();
  }
  plVar10 = (longlong *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803ab300);
  if ((plVar10 == (longlong *)0x0) || (*plVar10 == 0)) {
    PyModule_GetDict(DAT_1803ab780);
    FUN_1801ddaf0();
  }
  plVar10 = (longlong *)FUN_18038d790(param_3);
  uVar2 = DAT_1803ca5d8;
  uVar8 = DAT_1803ab328;
  puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803ab328);
  if (puVar7 == (undefined8 *)0x0) {
LAB_1801ede8b:
    PyDict_SetItem(uVar2,uVar8,plVar10);
  }
  else {
    plVar14 = (longlong *)*puVar7;
    if (plVar10 != plVar14) {
      if (plVar14 == (longlong *)0x0) goto LAB_1801ede8b;
      iVar19 = (int)*plVar10 + 1;
      if (iVar19 != 0) {
        *(int *)plVar10 = iVar19;
      }
      *puVar7 = plVar10;
      if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
        (**(code **)(plVar14[1] + 0x30))(plVar14);
      }
    }
  }
  if (DAT_1803c9fd0 == 0) {
    DAT_1803c9fd0 = PyImport_ImportModule("importlib._bootstrap");
  }
  plVar10 = (longlong *)PyObject_GetAttrString(DAT_1803c9fd0,"_spec_from_module");
  lVar9 = FUN_180376770(param_1,plVar10,DAT_1803ca5e0);
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if (lVar9 == 0) {
    PyErr_PrintEx(0);
                    /* WARNING: Subroutine does not return */
    abort();
  }
  FUN_18037f3a0();
  FUN_1801ddb80();
  pcVar21 = (code *)0x0;
  local_res20 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_60 = (longlong *)0x0;
  FUN_1801ddaf0();
  FUN_1801ddaf0();
  piVar5 = DAT_1803ca5e0;
  piVar1 = *(int **)(DAT_1803ca5e0 + 4);
  if (*piVar1 + 1 != 0) {
    *piVar1 = *piVar1 + 1;
  }
  lVar9 = FUN_18036cbb0(DAT_1803bced8,piVar5,piVar1,0);
  FUN_180001d00(param_1,lVar9);
/* novels.prompts.inkos.writer_prompts#299="__spec__" */
  plVar10 = (longlong *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bdaa8);
  if ((plVar10 == (longlong *)0x0) || (*plVar10 == 0)) {
/* novels.prompts.inkos.writer_prompts#299="__spec__" */
    FUN_1800015a0(DAT_1803ab778,DAT_1803bdaa8);
  }
  cVar6 = FUN_18037f3a0();
  if (cVar6 == '\0') {
LAB_1801ee010:
    plVar14 = *(longlong **)(param_1 + 0x60);
    uVar8 = 1;
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    FUN_1801ddc60();
    cVar6 = FUN_18037f3a0();
    if (cVar6 == '\0') goto LAB_1801ee010;
    FUN_1801ddaf0();
    FUN_180372070();
    FUN_18037f0c0();
    FUN_1801ddb80();
    plVar10 = (longlong *)FUN_180372110();
    iVar19 = (int)*plVar10 + 1;
    if (iVar19 != 0) {
      *(int *)plVar10 = iVar19;
    }
    if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
/* novels.prompts.inkos.writer_prompts#207="Any" */
      lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7c8,DAT_1803ab250);
    }
    else {
/* novels.prompts.inkos.writer_prompts#207="Any" */
      lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd7c8);
    }
    if (lVar11 == 0) {
LAB_1801ee1b8:
      uVar8 = 3;
LAB_1801ee1be:
      plVar14 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
    }
    else {
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#208="Literal" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7d0,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#208="Literal" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd7d0);
      }
      if (lVar11 == 0) goto LAB_1801ee1b8;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#209="TypedDict" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7d8,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#209="TypedDict" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd7d8);
      }
      if (lVar11 == 0) goto LAB_1801ee1b8;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd7e8;
/* novels.prompts.inkos.writer_prompts#210="en_prompt_sections" */
      uVar2 = DAT_1803bd7e0;
      uVar8 = DAT_1803ab258;
      pcVar12 = _Py_NoneStruct_exref;
      *(undefined4 *)(lVar9 + 0x28) = 5;
      plVar10 = (longlong *)FUN_1803717e0(param_1,uVar2,uVar4,pcVar12,uVar3,uVar8);
      if (plVar10 == (longlong *)0x0) {
        plVar14 = *(longlong **)(param_1 + 0x60);
        uVar8 = 5;
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1801eee3a;
      }
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#14="build_english_core_rules" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd1c0,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#14="build_english_core_rules" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd1c0);
      }
      if (lVar11 == 0) {
LAB_1801ee330:
        uVar8 = 5;
        goto LAB_1801ee1be;
      }
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#11="build_english_genre_intro" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd1a8,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#11="build_english_genre_intro" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd1a8);
      }
      if (lVar11 == 0) goto LAB_1801ee330;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd7f8;
/* novels.prompts.inkos.writer_prompts#212="fanfic_prompt_sections" */
      uVar2 = DAT_1803bd7f0;
      uVar8 = DAT_1803ab258;
      pcVar12 = _Py_NoneStruct_exref;
      *(undefined4 *)(lVar9 + 0x28) = 6;
      plVar10 = (longlong *)FUN_1803717e0(param_1,uVar2,uVar4,pcVar12,uVar3,uVar8);
      if (plVar10 == (longlong *)0x0) {
        plVar14 = *(longlong **)(param_1 + 0x60);
        uVar8 = 6;
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1801eee3a;
      }
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#214="FanficMode" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd800,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#214="FanficMode" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd800);
      }
      if (lVar11 == 0) {
LAB_1801ee555:
        uVar8 = 6;
        goto LAB_1801ee1be;
      }
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#35="build_character_voice_profiles" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd268,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#35="build_character_voice_profiles" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd268);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#31="build_fanfic_canon_section" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd248,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#31="build_fanfic_canon_section" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd248);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
/* novels.prompts.inkos.writer_prompts#36="build_fanfic_mode_instructions" */
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd270,DAT_1803ab250);
      }
      else {
/* novels.prompts.inkos.writer_prompts#36="build_fanfic_mode_instructions" */
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd270);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd810;
/* novels.prompts.inkos.writer_prompts#215="genre_profile_defaults" */
      uVar2 = DAT_1803bd808;
      uVar8 = DAT_1803ab258;
      pcVar12 = _Py_NoneStruct_exref;
      *(undefined4 *)(lVar9 + 0x28) = 0xc;
      plVar10 = (longlong *)FUN_1803717e0(param_1,uVar2,uVar4,pcVar12,uVar3,uVar8);
      if (plVar10 != (longlong *)0x0) {
        if (((code *)plVar10[1] == PyModule_Type_exref) ||
           (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
/* novels.prompts.inkos.writer_prompts#217="GenreProfile" */
          lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd818,DAT_1803ab250);
        }
        else {
/* novels.prompts.inkos.writer_prompts#217="GenreProfile" */
          lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd818);
        }
        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
          (**(code **)(plVar10[1] + 0x30))(plVar10);
        }
        if (lVar11 == 0) {
          plVar14 = *(longlong **)(param_1 + 0x60);
          uVar8 = 0xc;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801eee3a;
        }
        FUN_1801ddb80();
        uVar4 = DAT_1803ca5d8;
        uVar3 = DAT_1803bd828;
/* novels.prompts.inkos.writer_prompts#218="length_governance" */
        uVar2 = DAT_1803bd820;
        uVar8 = DAT_1803ab258;
        pcVar12 = _Py_NoneStruct_exref;
        *(undefined4 *)(lVar9 + 0x28) = 0xd;
        plVar10 = (longlong *)FUN_1803717e0(param_1,uVar2,uVar4,pcVar12,uVar3,uVar8);
        if (plVar10 == (longlong *)0x0) {
          plVar14 = *(longlong **)(param_1 + 0x60);
          uVar8 = 0xd;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801eee3a;
        }
        if (((code *)plVar10[1] == PyModule_Type_exref) ||
           (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
/* novels.prompts.inkos.writer_prompts#220="LengthSpec" */
          lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd830,DAT_1803ab250);
        }
        else {
/* novels.prompts.inkos.writer_prompts#220="LengthSpec" */
          lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd830);
        }
        if (lVar11 != 0) {
          FUN_1801ddb80();
          if (((code *)plVar10[1] == PyModule_Type_exref) ||
             (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
/* novels.prompts.inkos.writer_prompts#3="build_length_spec" */
            lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd168,DAT_1803ab250);
          }
          else {
/* novels.prompts.inkos.writer_prompts#3="build_length_spec" */
            lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd168);
          }
          if (lVar11 != 0) {
            FUN_1801ddb80();
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            lVar11 = FUN_1801ddc10();
            if (lVar11 == 0) {
/* novels.prompts.inkos.writer_prompts#208="Literal" */
              FUN_180373340(param_1,&local_res20,DAT_1803bd7d0);
              uVar8 = 0xf;
              plVar14 = local_res20;
            }
            else {
              lVar11 = FUN_180024e50(param_1,lVar11,DAT_1803bd838);
              if (lVar11 == 0) {
                plVar14 = *(longlong **)(param_1 + 0x60);
                uVar8 = 0xf;
                *(undefined8 *)(param_1 + 0x60) = 0;
              }
              else {
                FUN_1801ddb80();
                lVar11 = FUN_1801ddc10();
                if (lVar11 == 0) {
/* novels.prompts.inkos.writer_prompts#208="Literal" */
                  FUN_180373340(param_1,&local_res20,DAT_1803bd7d0);
                  uVar8 = 0x10;
                  plVar14 = local_res20;
                }
                else {
                  lVar11 = FUN_180024e50(param_1,lVar11,DAT_1803bd848);
                  if (lVar11 == 0) {
                    plVar14 = *(longlong **)(param_1 + 0x60);
                    uVar8 = 0x10;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                  }
                  else {
                    FUN_1801ddb80();
                    lVar11 = FUN_1801ddc10();
                    if (lVar11 == 0) {
/* novels.prompts.inkos.writer_prompts#208="Literal" */
                      FUN_180373340(param_1,&local_res20,DAT_1803bd7d0);
                      uVar8 = 0x11;
                      plVar14 = local_res20;
                    }
                    else {
                      lVar11 = FUN_180024e50(param_1,lVar11,DAT_1803bd858);
                      if (lVar11 == 0) {
                        plVar14 = *(longlong **)(param_1 + 0x60);
                        uVar8 = 0x11;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                      }
                      else {
                        FUN_1801ddb80();
/* novels.prompts.inkos.writer_prompts#209="TypedDict" */
                        puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bd7d8);
                        if ((puVar7 == (undefined8 *)0x0) ||
                           (local_res10[0] = (code *)*puVar7, local_res10[0] == (code *)0x0)) {
/* novels.prompts.inkos.writer_prompts#209="TypedDict" */
                          puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd7d8);
                          local_res10[0] = pcVar20;
                          if (puVar7 != (undefined8 *)0x0) {
                            local_res10[0] = (code *)*puVar7;
                          }
                          if (local_res10[0] == (code *)0x0) {
/* novels.prompts.inkos.writer_prompts#209="TypedDict" */
                            FUN_180373340(param_1,&local_res20,DAT_1803bd7d8);
                            uVar8 = 0x14;
                            plVar14 = local_res20;
                            goto LAB_1801eee3a;
                          }
                        }
                        lVar11 = *(longlong *)(param_1 + 0x10);
                        pcVar12 = *(code **)(lVar11 + 0x41d98);
                        if (pcVar12 == (code *)0x0) {
                          pcVar12 = (code *)FUN_18035ebd0(PyTuple_Type_exref,1);
                        }
                        else {
                          *(longlong *)(lVar11 + 0x41d98) = *(longlong *)(pcVar12 + 0x18);
                          *(int *)(lVar11 + 0x41e38) = *(int *)(lVar11 + 0x41e38) + -1;
                          *(longlong *)pcVar12 = 1;
                        }
                        *(longlong *)(pcVar12 + 0x18) = 0;
                        pcVar17 = pcVar12 + -0x10;
                        lVar11 = _PyThreadState_GetCurrent();
                        lVar11 = *(longlong *)(*(longlong *)(lVar11 + 0x10) + 0xd0);
                        puVar7 = *(undefined8 **)(lVar11 + 8);
                        *puVar7 = pcVar17;
                        *(ulonglong *)(pcVar12 + -8) = *(ulonglong *)(pcVar12 + -8) & 3;
                        *(ulonglong *)(pcVar12 + -8) =
                             *(ulonglong *)(pcVar12 + -8) | (ulonglong)puVar7;
                        *(longlong *)pcVar17 = lVar11;
                        *(code **)(lVar11 + 8) = pcVar17;
                        if (*(int *)local_res10[0] + 1 != 0) {
                          *(int *)local_res10[0] = *(int *)local_res10[0] + 1;
                        }
                        *(code **)(pcVar12 + 0x18) = local_res10[0];
                        if (*(int *)pcVar12 + 1 != 0) {
                          *(int *)pcVar12 = *(int *)pcVar12 + 1;
                        }
                        local_res10[0] = pcVar12;
                        plVar10 = (longlong *)FUN_180095960(param_1,local_res10);
                        pcVar17 = pcVar20;
                        if (plVar10 == (longlong *)0x0) {
                          plVar14 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          pcVar13 = pcVar20;
                        }
                        else {
                          pcVar13 = (code *)FUN_180370ec0(param_1);
                          local_res10[0] = pcVar13;
                          iVar19 = FUN_180001510(plVar10);
                          local_78 = plVar10;
                          if (iVar19 == -1) {
                            plVar14 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                          }
                          else {
                            pcVar17 = pcVar21;
                            if (iVar19 == 0) {
                              local_58 = PyType_Type_exref;
                              if (*(int *)PyType_Type_exref + 1 != 0) {
                                *(int *)PyType_Type_exref = *(int *)PyType_Type_exref + 1;
                              }
                            }
                            else {
                              plVar14 = (longlong *)FUN_180058050(param_1,plVar10,DAT_1803ab250);
                              if (plVar14 == (longlong *)0x0) {
                                plVar14 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                goto LAB_1801eeda8;
                              }
                              local_58 = (code *)plVar14[1];
                              if (*(int *)local_58 + 1 != 0) {
                                *(int *)local_58 = *(int *)local_58 + 1;
                              }
                              if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0))
                              {
                                (**(code **)(plVar14[1] + 0x30))(plVar14);
                              }
                            }
                            plVar14 = (longlong *)FUN_18036f490(param_1,local_58,plVar10);
                            if ((-1 < *(int *)local_58) &&
                               (*(longlong *)local_58 = *(longlong *)local_58 + -1,
                               *(longlong *)local_58 == 0)) {
                              (**(code **)(*(longlong *)(local_58 + 8) + 0x30))(local_58);
                            }
                            if (plVar14 == (longlong *)0x0) {
                              plVar14 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                            }
                            else {
/* novels.prompts.inkos.writer_prompts#227="__prepare__" */
                              iVar19 = FUN_18037f2f0(param_1,plVar14,DAT_1803bd868);
                              local_60 = plVar14;
                              if (iVar19 == -1) {
                                plVar14 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                              }
                              else if (iVar19 == 0) {
                                pcVar17 = (code *)FUN_180370ec0(param_1);
LAB_1801eef09:
                                if (*(int *)pcVar17 + 1 != 0) {
                                  *(int *)pcVar17 = *(int *)pcVar17 + 1;
                                }
/* novels.prompts.inkos.writer_prompts#231="__name__" */
                                puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bd888);
                                if (((puVar7 == (undefined8 *)0x0) ||
                                    (pcVar21 = (code *)*puVar7, (code *)*puVar7 == (code *)0x0)) &&
                                   (puVar7 = (undefined8 *)
/* novels.prompts.inkos.writer_prompts#231="__name__" */
                                             FUN_1800015a0(DAT_1803ab778,DAT_1803bd888),
                                   pcVar21 = pcVar20, puVar7 != (undefined8 *)0x0)) {
                                  pcVar21 = (code *)*puVar7;
                                }
/* novels.prompts.inkos.writer_prompts#233="__module__" */
                                iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd898,pcVar21);
                                if ((iVar19 == 0) &&
/* novels.prompts.inkos.writer_prompts#234="__qualname__" ; novels.prompts.inkos.writer_prompts#228="FanficContext" */
                                   (iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd8a0,DAT_1803bd870),
                                   iVar19 == 0)) {
                                  plVar15 = (longlong *)FUN_180370ec0(param_1);
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                  iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd8a8,plVar15);
                                  local_res10[0] = (code *)CONCAT44(local_res10[0]._4_4_,iVar19);
                                  if ((-1 < (int)*plVar15) &&
                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                                    iVar19 = (int)local_res10[0];
                                  }
                                  if (iVar19 != 0) goto LAB_1801ef4f5;
                                  lVar11 = FUN_18036cd80(param_1,DAT_1803bdae0,DAT_1803ca5e0);
                                  FUN_180001d00(param_1,lVar11);
/* novels.prompts.inkos.writer_prompts#236="str" */
                                  local_res10[0] = DAT_1803bd8b0;
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                  plVar15 = (longlong *)PyObject_GetItem(pcVar17,DAT_1803bd8a8);
                                  if (plVar15 == (longlong *)0x0) {
                                    plVar14 = *(longlong **)(param_1 + 0x60);
                                    if (plVar14 != (longlong *)0x0) {
                                      plVar10 = (longlong *)plVar14[1];
                                      if ((*(uint *)((longlong *)plVar14[1] + 0x15) & 0x40000000) ==
                                          0) {
                                        plVar10 = plVar14;
                                      }
                                      if ((plVar10 != *(longlong **)PyExc_KeyError_exref) &&
                                         (((-1 < *(int *)(plVar10[1] + 0xa8) ||
                                           ((*(uint *)(plVar10 + 0x15) & 0x40000000) == 0)) ||
                                          (cVar6 = FUN_180382640(), cVar6 != '\x01')))) {
                                        uVar8 = 0x15;
                                        goto LAB_1801ef2fe;
                                      }
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      if ((-1 < (int)*plVar14) &&
                                         (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                        (**(code **)(plVar14[1] + 0x30))(plVar14);
                                      }
                                    }
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                    FUN_180373340(param_1,&local_res20,DAT_1803bd8a8);
                                    uVar8 = 0x15;
                                    plVar14 = local_res20;
                                  }
                                  else {
/* novels.prompts.inkos.writer_prompts#32="fanficCanon" */
                                    cVar6 = FUN_18000d470(param_1,plVar15,DAT_1803bd250,
                                                          local_res10[0]);
                                    local_res10[0] = (code *)CONCAT71(local_res10[0]._1_7_,cVar6);
                                    if ((-1 < (int)*plVar15) &&
                                       (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                      (**(code **)(plVar15[1] + 0x30))(plVar15);
                                      cVar6 = (char)local_res10[0];
                                    }
                                    if (cVar6 == '\0') {
                                      uVar8 = 0x15;
LAB_1801ef2fa:
                                      plVar14 = *(longlong **)(param_1 + 0x60);
LAB_1801ef2fe:
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                    }
                                    else {
/* novels.prompts.inkos.writer_prompts#214="FanficMode" */
                                      local_res10[0] = DAT_1803bd800;
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                      plVar15 = (longlong *)PyObject_GetItem(pcVar17,DAT_1803bd8a8);
                                      if (plVar15 != (longlong *)0x0) {
/* novels.prompts.inkos.writer_prompts#33="fanficMode" */
                                        cVar6 = FUN_18000d470(param_1,plVar15,DAT_1803bd258,
                                                              local_res10[0]);
                                        local_res10[0] =
                                             (code *)CONCAT71(local_res10[0]._1_7_,cVar6);
                                        if ((-1 < (int)*plVar15) &&
                                           (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                          (**(code **)(plVar15[1] + 0x30))(plVar15);
                                          cVar6 = (char)local_res10[0];
                                        }
                                        if (cVar6 == '\0') {
                                          uVar8 = 0x16;
                                        }
                                        else {
/* novels.prompts.inkos.writer_prompts#237="list[str]" */
                                          local_res10[0] = DAT_1803bd8b8;
                                          plVar15 = (longlong *)
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                                    PyObject_GetItem(pcVar17,DAT_1803bd8a8);
                                          if (plVar15 == (longlong *)0x0) {
                                            plVar14 = *(longlong **)(param_1 + 0x60);
                                            if (plVar14 != (longlong *)0x0) {
                                              plVar10 = (longlong *)plVar14[1];
                                              if ((*(uint *)((longlong *)plVar14[1] + 0x15) &
                                                  0x40000000) == 0) {
                                                plVar10 = plVar14;
                                              }
                                              if ((plVar10 != *(longlong **)PyExc_KeyError_exref) &&
                                                 (((-1 < *(int *)(plVar10[1] + 0xa8) ||
                                                   ((*(uint *)(plVar10 + 0x15) & 0x40000000) == 0))
                                                  || (cVar6 = FUN_180382640(), cVar6 != '\x01')))) {
                                                uVar8 = 0x17;
                                                goto LAB_1801ef2fe;
                                              }
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                              if ((-1 < (int)*plVar14) &&
                                                 (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                (**(code **)(plVar14[1] + 0x30))(plVar14);
                                              }
                                            }
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                            FUN_180373340(param_1,&local_res20,DAT_1803bd8a8);
                                            uVar8 = 0x17;
                                            plVar14 = local_res20;
                                            goto LAB_1801ef302;
                                          }
/* novels.prompts.inkos.writer_prompts#37="allowedDeviations" */
                                          cVar6 = FUN_18000d470(param_1,plVar15,DAT_1803bd278,
                                                                local_res10[0]);
                                          local_res10[0] =
                                               (code *)CONCAT71(local_res10[0]._1_7_,cVar6);
                                          if ((-1 < (int)*plVar15) &&
                                             (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                            (**(code **)(plVar15[1] + 0x30))(plVar15);
                                            cVar6 = (char)local_res10[0];
                                          }
                                          if (cVar6 != '\0') {
                                            FUN_180001d80(param_1);
                                            iVar19 = FUN_180386a00(plVar10,pcVar12);
                                            if ((iVar19 != -1) &&
                                               ((iVar19 != 1 ||
/* novels.prompts.inkos.writer_prompts#238="__orig_bases__" */
                                                (iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd8c0,
                                                                           pcVar12), iVar19 == 0))))
                                            {
                                              lVar11 = *(longlong *)(param_1 + 0x10);
/* novels.prompts.inkos.writer_prompts#228="FanficContext" */
                                              local_res10[0] = DAT_1803bd870;
                                              plVar15 = *(longlong **)(lVar11 + 0x41da8);
                                              if (plVar15 == (longlong *)0x0) {
                                                plVar15 = (longlong *)
                                                          FUN_18035ebd0(PyTuple_Type_exref,3);
                                              }
                                              else {
                                                *(longlong *)(lVar11 + 0x41da8) = plVar15[3];
                                                *(int *)(lVar11 + 0x41e40) =
                                                     *(int *)(lVar11 + 0x41e40) + -1;
                                                *plVar15 = 1;
                                              }
                                              plVar15[3] = 0;
                                              plVar15[4] = 0;
                                              plVar15[5] = 0;
                                              FUN_1801f33c0(plVar15);
                                              if (*(int *)local_res10[0] + 1 != 0) {
                                                *(int *)local_res10[0] = *(int *)local_res10[0] + 1;
                                              }
                                              plVar15[3] = (longlong)local_res10[0];
                                              iVar19 = (int)*plVar10 + 1;
                                              if (iVar19 != 0) {
                                                *(int *)plVar10 = iVar19;
                                              }
                                              plVar15[4] = (longlong)plVar10;
                                              if (*(int *)pcVar17 + 1 != 0) {
                                                *(int *)pcVar17 = *(int *)pcVar17 + 1;
                                              }
                                              plVar15[5] = (longlong)pcVar17;
                                              *(undefined4 *)(lVar9 + 0x28) = 0x14;
                                              local_res10[0] =
                                                   (code *)FUN_180001790(param_1,plVar14,plVar15,
                                                                         pcVar13);
                                              if ((-1 < (int)*plVar15) &&
                                                 (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                (**(code **)(plVar15[1] + 0x30))(plVar15);
                                              }
                                              if (local_res10[0] != (code *)0x0) {
                                                if (*(int *)local_res10[0] + 1 != 0) {
                                                  *(int *)local_res10[0] =
                                                       *(int *)local_res10[0] + 1;
                                                }
                                                if ((-1 < *(int *)pcVar17) &&
                                                   (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1
                                                   , *(longlong *)pcVar17 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))
                                                            (pcVar17);
                                                }
                                                if ((-1 < *(int *)local_res10[0]) &&
                                                   (*(longlong *)local_res10[0] =
                                                         *(longlong *)local_res10[0] + -1,
                                                   *(longlong *)local_res10[0] == 0)) {
                                                  (**(code **)(*(longlong *)(local_res10[0] + 8) +
                                                              0x30))(local_res10[0]);
                                                }
                                                FUN_1801ddb80();
                                                FUN_180001d80(param_1);
                                                if ((-1 < *(int *)pcVar12) &&
                                                   (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1
                                                   , *(longlong *)pcVar12 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))
                                                            (pcVar12);
                                                }
                                                if ((-1 < (int)*plVar10) &&
                                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                                }
                                                if ((-1 < *(int *)pcVar13) &&
                                                   (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1
                                                   , *(longlong *)pcVar13 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))
                                                            (pcVar13);
                                                }
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  (**(code **)(plVar14[1] + 0x30))();
                                                }
                                                if ((-1 < *(int *)pcVar17) &&
                                                   (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1
                                                   , *(longlong *)pcVar17 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))
                                                            (pcVar17);
                                                }
                                                piVar1 = DAT_1803bd8c8;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8d0);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
/* novels.prompts.inkos.writer_prompts#241="build_writer_system_prompt" */
                                                FUN_1803609a0(FUN_1801dfdd0,DAT_1803bd8d8,0,
                                                              DAT_1803bd0f8,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e0);
/* novels.prompts.inkos.writer_prompts#38="build_genre_intro" */
                                                FUN_1803609a0(FUN_1801e2da0,DAT_1803bd280,0,
                                                              DAT_1803bcf90,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e8);
/* novels.prompts.inkos.writer_prompts#15="build_governed_input_contract" */
                                                FUN_1803609a0(FUN_1801e31a0,DAT_1803bd1c8,0,
                                                              DAT_1803bd0e0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e8);
/* novels.prompts.inkos.writer_prompts#16="build_chapter_memo_contract" */
                                                FUN_1803609a0(FUN_1801e3500,DAT_1803bd1d0,0,
                                                              DAT_1803bceb0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f0);
/* novels.prompts.inkos.writer_prompts#17="build_length_guidance" */
                                                FUN_1803609a0(FUN_1801e3860,DAT_1803bd1d8,0,
                                                              DAT_1803bceb8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f8);
/* novels.prompts.inkos.writer_prompts#39="build_core_rules" */
                                                FUN_1803609a0(FUN_1801e4130,DAT_1803bd288,0,
                                                              DAT_1803bd110,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#248="build_anti_ai_examples" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd910,0,
                                                                      DAT_1803bd068,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#247="## 去AI味：反例→正例对照\n\n以下对照表展示AI常犯的\"味道\"问题和修正方法。正文必须贴近正例风格。\n\n### 情绪描写\n| 反例（AI味） | 正例（人味） | 要点 |\n|---|---|---|\n| 他感到非常愤怒。 | 他捏碎了手中的茶杯，滚烫的茶水流过指缝，但他像没感觉一样。 | 用动作外化情绪 |\n| 她心里很悲伤，眼泪流了下来。 | 她攥紧手机，指节发白，屏幕上的聊天记录模糊成一片。 | 用身体细节替代直白标签 |\n| 他感到一阵恐惧。 | 他后背的汗毛竖了" */
                                                uVar8 = DAT_1803bd908;
/* novels.prompts.inkos.writer_prompts#247="## 去AI味：反例→正例对照\n\n以下对照表展示AI常犯的\"味道\"问题和修正方法。正文必须贴近正例风格。\n\n### 情绪描写\n| 反例（AI味） | 正例（人味） | 要点 |\n|---|---|---|\n| 他感到非常愤怒。 | 他捏碎了手中的茶杯，滚烫的茶水流过指缝，但他像没感觉一样。 | 用动作外化情绪 |\n| 她心里很悲伤，眼泪流了下来。 | 她攥紧手机，指节发白，屏幕上的聊天记录模糊成一片。 | 用身体细节替代直白标签 |\n| 他感到一阵恐惧。 | 他后背的汗毛竖了" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd908;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#248="build_anti_ai_examples" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd910,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#250="build_character_psychology_method" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd920,0,
                                                                      DAT_1803bdac0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#249="## 六步走人物心理分析\n\n每个重要角色在关键场景中的行为，必须经过以下六步推导：\n\n1. **当前处境**：角色此刻面临什么局面？手上有什么牌？\n2. **核心动机**：角色最想要什么？最害怕什么？\n3. **信息边界**：角色知道什么？不知道什么？对局势有什么误判？\n4. **性格过滤**：同样的局面，这个角色的性格会怎么反应？（冲动/谨慎/阴险/果断）\n5. **行为选择**：基于以上四点，角色会做出什么选择？\n6. **情绪外化**：这个选择伴随什么情绪？用什么身体语" */
                                                uVar8 = DAT_1803bd918;
/* novels.prompts.inkos.writer_prompts#249="## 六步走人物心理分析\n\n每个重要角色在关键场景中的行为，必须经过以下六步推导：\n\n1. **当前处境**：角色此刻面临什么局面？手上有什么牌？\n2. **核心动机**：角色最想要什么？最害怕什么？\n3. **信息边界**：角色知道什么？不知道什么？对局势有什么误判？\n4. **性格过滤**：同样的局面，这个角色的性格会怎么反应？（冲动/谨慎/阴险/果断）\n5. **行为选择**：基于以上四点，角色会做出什么选择？\n6. **情绪外化**：这个选择伴随什么情绪？用什么身体语" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd918;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#250="build_character_psychology_method" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd920,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#252="build_supporting_character_method" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd930,0,
                                                                      DAT_1803bd078,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#251="## 配角设计方法论\n\n### 配角B面原则\n配角必须有反击，有自己的算盘。主角的强大在于压服聪明人，而不是碾压傻子。\n\n### 构建方法\n1. **动机绑定主线**：每个配角的行为动机必须与主线产生关联\n   - 反派对抗主角不是因为\"反派脸谱\"，而是有自己的诉求（如保护家人、争夺生存资源）\n   - 盟友帮助主角是因为有共同敌人或欠了人情，而非无条件忠诚\n2. **核心标签 + 反差细节**：让配角\"活\"过来\n   - 表面冷硬的角色有不为人知的温柔一面（如偷偷照顾流浪动" */
                                                uVar8 = DAT_1803bd928;
/* novels.prompts.inkos.writer_prompts#251="## 配角设计方法论\n\n### 配角B面原则\n配角必须有反击，有自己的算盘。主角的强大在于压服聪明人，而不是碾压傻子。\n\n### 构建方法\n1. **动机绑定主线**：每个配角的行为动机必须与主线产生关联\n   - 反派对抗主角不是因为\"反派脸谱\"，而是有自己的诉求（如保护家人、争夺生存资源）\n   - 盟友帮助主角是因为有共同敌人或欠了人情，而非无条件忠诚\n2. **核心标签 + 反差细节**：让配角\"活\"过来\n   - 表面冷硬的角色有不为人知的温柔一面（如偷偷照顾流浪动" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd928;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#252="build_supporting_character_method" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd930,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#254="build_reader_psychology_method" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd940,0,
                                                                      DAT_1803bce90,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#253="## 读者心理学框架\n\n写作时同步考虑读者的心理状态：\n\n- **期待管理**：在读者期待释放时，适当延迟以增强快感；在读者即将失去耐心时，立即给反馈\n- **信息落差**：让读者比角色多知道一点（制造紧张），或比角色少知道一点（制造好奇）\n- **情绪节拍**：压制→释放→更大的压制→更大的释放。释放时要超过读者心理预期。递进式升级——不是一次到位，而是层层加码（被骂→手机掉下水道→被噎住→有人敲门），每次比上一次更过分\n- **锚定效应**：先给读者一个参照（对手有多强/" */
                                                uVar8 = DAT_1803bd938;
/* novels.prompts.inkos.writer_prompts#253="## 读者心理学框架\n\n写作时同步考虑读者的心理状态：\n\n- **期待管理**：在读者期待释放时，适当延迟以增强快感；在读者即将失去耐心时，立即给反馈\n- **信息落差**：让读者比角色多知道一点（制造紧张），或比角色少知道一点（制造好奇）\n- **情绪节拍**：压制→释放→更大的压制→更大的释放。释放时要超过读者心理预期。递进式升级——不是一次到位，而是层层加码（被骂→手机掉下水道→被噎住→有人敲门），每次比上一次更过分\n- **锚定效应**：先给读者一个参照（对手有多强/" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd938;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#254="build_reader_psychology_method" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd940,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#256="build_emotional_pacing_method" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd950,0,
                                                                      DAT_1803bcfe0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#255="## 情感节点设计\n\n关系发展（友情、爱情、从属）必须经过事件驱动的节点递进：\n\n1. **设计3-5个关键事件**：共同御敌、秘密分享、利益冲突、信任考验、牺牲/妥协\n2. **递进升温**：每个事件推进关系一个层级，禁止跨越式发展（初见即死忠、一面之缘即深情）\n3. **情绪用场景传达**：环境烘托（暴雨中独坐）+ 微动作（攥拳指尖发白）替代直白抒情\n4. **情感与题材匹配**：末世侧重\"共患难的信任\"、悬疑侧重\"试探与默契\"、玄幻侧重\"利益捆绑到真正认可\"\n5. **" */
                                                uVar8 = DAT_1803bd948;
/* novels.prompts.inkos.writer_prompts#255="## 情感节点设计\n\n关系发展（友情、爱情、从属）必须经过事件驱动的节点递进：\n\n1. **设计3-5个关键事件**：共同御敌、秘密分享、利益冲突、信任考验、牺牲/妥协\n2. **递进升温**：每个事件推进关系一个层级，禁止跨越式发展（初见即死忠、一面之缘即深情）\n3. **情绪用场景传达**：环境烘托（暴雨中独坐）+ 微动作（攥拳指尖发白）替代直白抒情\n4. **情感与题材匹配**：末世侧重\"共患难的信任\"、悬疑侧重\"试探与默契\"、玄幻侧重\"利益捆绑到真正认可\"\n5. **" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd948;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#256="build_emotional_pacing_method" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd950,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#258="build_immersion_techniques" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd960,0,
                                                                      DAT_1803bd100,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#257="## 代入感技法\n\n- **自然信息交代**：角色身份/外貌/背景通过行动和对话带出，禁止\"资料卡式\"直接罗列\n- **画面代入法**：开场先给画面（动作、环境、声音），再给信息，让读者\"看到\"而非\"被告知\"\n- **共鸣锚点**：主角的困境必须有普遍性（被欺压、不公待遇、被低估），让读者觉得\"这也是我\"\n- **欲望钩子**：每章至少让读者产生一个\"接下来会怎样\"的好奇心\n- **信息落差应用**：让读者比角色多知道一点（紧张感）或少知道一点（好奇心），动态切换\n- **具" */
                                                uVar8 = DAT_1803bd958;
/* novels.prompts.inkos.writer_prompts#257="## 代入感技法\n\n- **自然信息交代**：角色身份/外貌/背景通过行动和对话带出，禁止\"资料卡式\"直接罗列\n- **画面代入法**：开场先给画面（动作、环境、声音），再给信息，让读者\"看到\"而非\"被告知\"\n- **共鸣锚点**：主角的困境必须有普遍性（被欺压、不公待遇、被低估），让读者觉得\"这也是我\"\n- **欲望钩子**：每章至少让读者产生一个\"接下来会怎样\"的好奇心\n- **信息落差应用**：让读者比角色多知道一点（紧张感）或少知道一点（好奇心），动态切换\n- **具" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd958;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#258="build_immersion_techniques" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd960);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
/* novels.prompts.inkos.writer_prompts#19="build_writing_craft_card" */
                                                FUN_1803609a0(FUN_1801e4550,DAT_1803bd1e8,0,
                                                              DAT_1803bcf38,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
/* novels.prompts.inkos.writer_prompts#22="build_creative_constitution" */
                                                FUN_1803609a0(FUN_1801e47a0,DAT_1803bd200,0,
                                                              DAT_1803bdad8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
/* novels.prompts.inkos.writer_prompts#23="build_immersion_pillars" */
                                                FUN_1803609a0(FUN_1801e49f0,DAT_1803bd208,0,
                                                              DAT_1803bcfa8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd970);
/* novels.prompts.inkos.writer_prompts#24="build_golden_opening_discipline" */
                                                FUN_1803609a0(FUN_1801e4c40,DAT_1803bd210,0,
                                                              DAT_1803bcf20,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                piVar1 = DAT_1803bd978;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd980);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
/* novels.prompts.inkos.writer_prompts#41="build_golden_chapters_rules" */
                                                FUN_1803609a0(FUN_1801e5240,DAT_1803bd298,0,
                                                              DAT_1803bd028,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
/* novels.prompts.inkos.writer_prompts#44="build_full_cast_tracking" */
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd2b0,0,
                                                                      DAT_1803bcfc0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
/* novels.prompts.inkos.writer_prompts#263="## 全员追踪\n\n本书启用全员追踪模式。每章结束时，POST_SETTLEMENT 必须额外包含：\n- 本章出场角色清单（名字 + 一句话状态变化）\n- 角色间关系变动（如有）\n- 未出场但被提及的角色（名字 + 提及原因）" */
                                                uVar8 = DAT_1803bd988;
/* novels.prompts.inkos.writer_prompts#263="## 全员追踪\n\n本书启用全员追踪模式。每章结束时，POST_SETTLEMENT 必须额外包含：\n- 本章出场角色清单（名字 + 一句话状态变化）\n- 角色间关系变动（如有）\n- 未出场但被提及的角色（名字 + 提及原因）" */
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd988;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
/* novels.prompts.inkos.writer_prompts#44="build_full_cast_tracking" */
                                                FUN_1801ddb80(uVar8,DAT_1803bd2b0,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd990);
/* novels.prompts.inkos.writer_prompts#25="build_genre_rules" */
                                                FUN_1803609a0(FUN_1801e5d00,DAT_1803bd218,0,
                                                              DAT_1803bd0b8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd998);
/* novels.prompts.inkos.writer_prompts#27="build_narrative_person_rule" */
                                                FUN_1803609a0(FUN_1801e7520,DAT_1803bd228,0,
                                                              DAT_1803bdab8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
/* novels.prompts.inkos.writer_prompts#21="build_prose_execution_rules" */
                                                FUN_1803609a0(FUN_1801e7950,DAT_1803bd1f8,0,
                                                              DAT_1803bd070,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9a0);
/* novels.prompts.inkos.writer_prompts#26="build_protagonist_rules" */
                                                FUN_1803609a0(FUN_1801e7ba0,DAT_1803bd220,0,
                                                              DAT_1803bd118,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9a8);
/* novels.prompts.inkos.writer_prompts#28="build_book_rules_body" */
                                                FUN_1803609a0(FUN_1801e9380,DAT_1803bd230,0,
                                                              DAT_1803bcf50,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9b0);
/* novels.prompts.inkos.writer_prompts#29="build_style_guide" */
                                                FUN_1803609a0(FUN_1801e9720,DAT_1803bd238,0,
                                                              DAT_1803bd048,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                piVar1 = DAT_1803bd978;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9b8);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
/* novels.prompts.inkos.writer_prompts#30="build_style_fingerprint" */
                                                FUN_1803609a0(FUN_1801e9ae0,DAT_1803bd240,0,
                                                              DAT_1803bcea0,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e0);
/* novels.prompts.inkos.writer_prompts#270="build_pre_write_checklist" */
                                                FUN_1803609a0(FUN_1801e9e80,DAT_1803bd9c0,0,
                                                              DAT_1803bd008,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9c8);
/* novels.prompts.inkos.writer_prompts#9="build_creative_output_format" */
                                                FUN_1803609a0(FUN_1801eb0b0,DAT_1803bd198,0,
                                                              DAT_1803bcf30,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9c8);
/* novels.prompts.inkos.writer_prompts#10="build_output_format" */
                                                FUN_1803609a0(FUN_1801eb930,DAT_1803bd1a0,0,
                                                              DAT_1803bd0c8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d0);
/* novels.prompts.inkos.writer_prompts#191="build_english_pre_write_table" */
                                                FUN_1803609a0(FUN_1801ec5a0,DAT_1803bd748,0,
                                                              DAT_1803bd098,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f8);
/* novels.prompts.inkos.writer_prompts#192="build_english_content_blocks" */
                                                FUN_1803609a0(FUN_1801ecac0,DAT_1803bd750,0,
                                                              DAT_1803bcf78,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d8);
/* novels.prompts.inkos.writer_prompts#7="build_english_creative_output_format" */
                                                FUN_1803609a0(FUN_1801ecee0,DAT_1803bd188,0,
                                                              DAT_1803bdab0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d8);
/* novels.prompts.inkos.writer_prompts#8="build_english_output_format" */
                                                FUN_1803609a0(FUN_1801ed3b0,DAT_1803bd190,0,
                                                              DAT_1803bd108,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                piVar1 = DAT_1803ca5e0;
                                                if (*DAT_1803ca5e0 + 1 == 0) {
                                                  return DAT_1803ca5e0;
                                                }
                                                *DAT_1803ca5e0 = *DAT_1803ca5e0 + 1;
                                                return piVar1;
                                              }
                                            }
                                            goto LAB_1801ef4f5;
                                          }
                                          uVar8 = 0x17;
                                        }
                                        goto LAB_1801ef2fa;
                                      }
                                      plVar14 = *(longlong **)(param_1 + 0x60);
                                      if (plVar14 != (longlong *)0x0) {
                                        plVar10 = (longlong *)plVar14[1];
                                        if ((*(uint *)((longlong *)plVar14[1] + 0x15) & 0x40000000)
                                            == 0) {
                                          plVar10 = plVar14;
                                        }
                                        if ((plVar10 != *(longlong **)PyExc_KeyError_exref) &&
                                           (((-1 < *(int *)(plVar10[1] + 0xa8) ||
                                             ((*(uint *)(plVar10 + 0x15) & 0x40000000) == 0)) ||
                                            (cVar6 = FUN_180382640(), cVar6 != '\x01')))) {
                                          uVar8 = 0x16;
                                          goto LAB_1801ef2fe;
                                        }
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                        if ((-1 < (int)*plVar14) &&
                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                                        }
                                      }
/* novels.prompts.inkos.writer_prompts#235="__annotations__" */
                                      FUN_180373340(param_1,&local_res20,DAT_1803bd8a8);
                                      uVar8 = 0x16;
                                      plVar14 = local_res20;
                                    }
                                  }
LAB_1801ef302:
                                  local_res10[0] = (code *)plVar14[5];
                                  pcVar21 = pcVar20;
                                  if (local_res10[0] == (code *)0x0) {
                                    pcVar18 = (code *)FUN_18038e2a0(lVar11,uVar8);
                                    plVar10 = (longlong *)plVar14[5];
                                    if (((pcVar18 != _Py_NoneStruct_exref) &&
                                        (pcVar21 = pcVar18, pcVar18 != (code *)0x0)) &&
                                       (*(int *)pcVar18 + 1 != 0)) {
                                      *(int *)pcVar18 = *(int *)pcVar18 + 1;
                                    }
LAB_1801ef388:
                                    plVar14[5] = (longlong)pcVar21;
                                    if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                                       (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                      (**(code **)(plVar10[1] + 0x30))();
                                    }
                                    if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
                                       (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                                       *(longlong *)pcVar18 == 0)) {
                                      (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                                    }
                                  }
                                  else if (*(longlong *)(local_res10[0] + 0x18) != lVar11) {
                                    pcVar18 = (code *)FUN_18038e2a0(lVar11,uVar8);
                                    *(code **)(pcVar18 + 0x10) = local_res10[0];
                                    if (*(int *)local_res10[0] + 1 != 0) {
                                      *(int *)local_res10[0] = *(int *)local_res10[0] + 1;
                                    }
                                    plVar10 = (longlong *)plVar14[5];
                                    if (((pcVar18 != _Py_NoneStruct_exref) &&
                                        (pcVar21 = pcVar18, pcVar18 != (code *)0x0)) &&
                                       (*(int *)pcVar18 + 1 != 0)) {
                                      *(int *)pcVar18 = *(int *)pcVar18 + 1;
                                    }
                                    goto LAB_1801ef388;
                                  }
                                  FUN_18036d0a0(lVar11,&DAT_180394940,0);
                                  FUN_180001d80(param_1);
                                }
                                else {
LAB_1801ef4f5:
                                  plVar14 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                if ((-1 < *(int *)pcVar17) &&
                                   (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1,
                                   *(longlong *)pcVar17 == 0)) {
                                  (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
                                }
                              }
                              else {
                                plVar15 = (longlong *)FUN_18037f0c0();
/* novels.prompts.inkos.writer_prompts#228="FanficContext" */
                                pcVar18 = DAT_1803bd870;
                                if (plVar15 == (longlong *)0x0) {
                                  plVar14 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                }
                                else {
                                  lVar11 = *(longlong *)(param_1 + 0x10);
                                  plVar16 = *(longlong **)(lVar11 + 0x41da0);
                                  if (plVar16 == (longlong *)0x0) {
                                    plVar16 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
                                  }
                                  else {
                                    *(longlong *)(lVar11 + 0x41da0) = plVar16[3];
                                    *(int *)(lVar11 + 0x41e3c) = *(int *)(lVar11 + 0x41e3c) + -1;
                                    *plVar16 = 1;
                                  }
                                  plVar16[3] = 0;
                                  plVar16[4] = 0;
                                  FUN_1801f33c0(plVar16);
                                  if (*(int *)pcVar18 + 1 != 0) {
                                    *(int *)pcVar18 = *(int *)pcVar18 + 1;
                                  }
                                  plVar16[3] = (longlong)pcVar18;
                                  iVar19 = (int)*plVar10 + 1;
                                  if (iVar19 != 0) {
                                    *(int *)plVar10 = iVar19;
                                  }
                                  plVar16[4] = (longlong)plVar10;
                                  *(undefined4 *)(lVar9 + 0x28) = 0x14;
                                  pcVar17 = (code *)FUN_180001790(param_1,plVar15,plVar16,pcVar13);
                                  if ((-1 < (int)*plVar15) &&
                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                                  }
                                  if ((-1 < (int)*plVar16) &&
                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                    (**(code **)(plVar16[1] + 0x30))(plVar16);
                                  }
                                  if (pcVar17 == (code *)0x0) {
                                    plVar14 = *(longlong **)(param_1 + 0x60);
                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                    pcVar17 = pcVar21;
                                  }
                                  else {
/* novels.prompts.inkos.writer_prompts#229="__getitem__" */
                                    iVar19 = FUN_18037f2f0(param_1,pcVar17,DAT_1803bd878);
                                    if (iVar19 == -1) {
                                      plVar14 = *(longlong **)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                    }
                                    else {
                                      pcVar21 = _Py_FalseStruct_exref;
                                      if (iVar19 != 0) {
                                        pcVar21 = _Py_TrueStruct_exref;
                                      }
                                      iVar19 = FUN_180001510(pcVar21);
/* novels.prompts.inkos.writer_prompts#230="%s.__prepare__() must return a mapping, not %s" */
                                      uVar8 = DAT_1803bd880;
                                      if (iVar19 != 0) goto LAB_1801eef09;
                                      local_res20 = (longlong *)
/* novels.prompts.inkos.writer_prompts#231="__name__" */
                                                    FUN_18036f270(param_1,plVar14,DAT_1803bd888,
/* novels.prompts.inkos.writer_prompts#232="<metaclass>" */
                                                                  DAT_1803bd890);
                                      if (local_res20 == (longlong *)0x0) {
                                        plVar14 = *(longlong **)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                      }
                                      else {
                                        lVar11 = *(longlong *)(param_1 + 0x10);
                                        plVar10 = *(longlong **)(lVar11 + 0x41da0);
                                        if (plVar10 == (longlong *)0x0) {
                                          plVar10 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
                                        }
                                        else {
                                          *(longlong *)(lVar11 + 0x41da0) = plVar10[3];
                                          *(int *)(lVar11 + 0x41e3c) =
                                               *(int *)(lVar11 + 0x41e3c) + -1;
                                          *plVar10 = 1;
                                        }
                                        plVar10[3] = 0;
                                        plVar10[4] = 0;
                                        FUN_1801f33c0(plVar10);
                                        plVar10[3] = (longlong)local_res20;
                                        plVar14 = *(longlong **)(pcVar17 + 8);
                                        iVar19 = (int)*plVar14 + 1;
                                        if (iVar19 != 0) {
                                          *(int *)plVar14 = iVar19;
                                        }
/* novels.prompts.inkos.writer_prompts#231="__name__" */
                                        lVar11 = FUN_18037f0c0(iVar19,plVar14,DAT_1803bd888);
                                        if ((-1 < (int)*plVar14) &&
                                           (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                          (**(code **)(plVar14[1] + 0x30))(plVar14);
                                        }
                                        if (lVar11 == 0) {
                                          plVar14 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          pcVar13 = local_res10[0];
                                          if ((-1 < (int)*plVar10) &&
                                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                                            pcVar13 = local_res10[0];
                                          }
                                        }
                                        else {
                                          plVar10[4] = lVar11;
                                          plVar15 = (longlong *)FUN_180381520(uVar8,plVar10);
                                          if ((-1 < (int)*plVar10) &&
                                             (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                                          }
                                          pcVar21 = PyExc_TypeError_exref;
                                          if (plVar15 == (longlong *)0x0) {
                                            plVar14 = *(longlong **)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                            pcVar13 = local_res10[0];
                                          }
                                          else {
                                            *(undefined4 *)(lVar9 + 0x28) = 0x14;
                                            plVar14 = (longlong *)
                                                      FUN_180376770(param_1,*(undefined8 *)pcVar21,
                                                                    plVar15);
                                            if ((-1 < (int)*plVar15) &&
                                               (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                              (**(code **)(plVar15[1] + 0x30))(plVar15);
                                            }
                                            FUN_18000d1d0(param_1,plVar14);
                                            pcVar13 = local_res10[0];
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
LAB_1801eeda8:
                        uVar8 = 0x14;
                        if ((-1 < *(int *)pcVar12) &&
                           (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1,
                           *(longlong *)pcVar12 == 0)) {
                          (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                        }
                        if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                           (*local_78 = *local_78 + -1, *local_78 == 0)) {
                          (**(code **)(local_78[1] + 0x30))();
                        }
                        if (((pcVar13 != (code *)0x0) && (-1 < *(int *)pcVar13)) &&
                           (*(longlong *)pcVar13 = *(longlong *)pcVar13 + -1,
                           *(longlong *)pcVar13 == 0)) {
                          (**(code **)(*(longlong *)(pcVar13 + 8) + 0x30))(pcVar13);
                        }
                        if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
                           (*local_60 = *local_60 + -1, *local_60 == 0)) {
                          (**(code **)(local_60[1] + 0x30))();
                        }
                        if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
                           (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1,
                           *(longlong *)pcVar17 == 0)) {
                          (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
                        }
                      }
                    }
                  }
                }
              }
            }
            goto LAB_1801eee3a;
          }
        }
        uVar8 = 0xd;
        goto LAB_1801ee1be;
      }
      plVar14 = *(longlong **)(param_1 + 0x60);
      uVar8 = 0xc;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
  }
LAB_1801eee3a:
  piVar1 = (int *)plVar14[5];
  if (piVar1 == (int *)0x0) {
    pcVar21 = (code *)FUN_18038e2a0(lVar9,uVar8);
    bVar22 = pcVar21 == _Py_NoneStruct_exref;
  }
  else {
    if (*(longlong *)(piVar1 + 6) == lVar9) goto LAB_1801ef5a2;
    pcVar21 = (code *)FUN_18038e2a0(lVar9,uVar8);
    *(int **)(pcVar21 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    bVar22 = pcVar21 == _Py_NoneStruct_exref;
  }
  plVar10 = (longlong *)plVar14[5];
  if (((!bVar22) && (pcVar20 = pcVar21, pcVar21 != (code *)0x0)) && (*(int *)pcVar21 + 1 != 0)) {
    *(int *)pcVar21 = *(int *)pcVar21 + 1;
  }
  plVar14[5] = (longlong)pcVar20;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  if (((pcVar21 != (code *)0x0) && (-1 < *(int *)pcVar21)) &&
     (*(longlong *)pcVar21 = *(longlong *)pcVar21 + -1, *(longlong *)pcVar21 == 0)) {
    (**(code **)(*(longlong *)(pcVar21 + 8) + 0x30))(pcVar21);
  }
LAB_1801ef5a2:
  FUN_180001d80(param_1);
  plVar10 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar14;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  return (int *)0x0;
}
