
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
  plVar10 = (longlong *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bdaa8);
  if ((plVar10 == (longlong *)0x0) || (*plVar10 == 0)) {
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
      lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7c8,DAT_1803ab250);
    }
    else {
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
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7d0,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd7d0);
      }
      if (lVar11 == 0) goto LAB_1801ee1b8;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd7d8,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd7d8);
      }
      if (lVar11 == 0) goto LAB_1801ee1b8;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd7e8;
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
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd1c0,DAT_1803ab250);
      }
      else {
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
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd1a8,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd1a8);
      }
      if (lVar11 == 0) goto LAB_1801ee330;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd7f8;
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
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd800,DAT_1803ab250);
      }
      else {
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
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd268,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd268);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd248,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd248);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if (((code *)plVar10[1] == PyModule_Type_exref) || (iVar19 = PyType_IsSubtype(), iVar19 != 0))
      {
        lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd270,DAT_1803ab250);
      }
      else {
        lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd270);
      }
      if (lVar11 == 0) goto LAB_1801ee555;
      FUN_1801ddb80();
      if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
        (**(code **)(plVar10[1] + 0x30))(plVar10);
      }
      uVar4 = DAT_1803ca5d8;
      uVar3 = DAT_1803bd810;
      uVar2 = DAT_1803bd808;
      uVar8 = DAT_1803ab258;
      pcVar12 = _Py_NoneStruct_exref;
      *(undefined4 *)(lVar9 + 0x28) = 0xc;
      plVar10 = (longlong *)FUN_1803717e0(param_1,uVar2,uVar4,pcVar12,uVar3,uVar8);
      if (plVar10 != (longlong *)0x0) {
        if (((code *)plVar10[1] == PyModule_Type_exref) ||
           (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
          lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd818,DAT_1803ab250);
        }
        else {
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
          lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd830,DAT_1803ab250);
        }
        else {
          lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd830);
        }
        if (lVar11 != 0) {
          FUN_1801ddb80();
          if (((code *)plVar10[1] == PyModule_Type_exref) ||
             (iVar19 = PyType_IsSubtype(), iVar19 != 0)) {
            lVar11 = FUN_180371d80(param_1,plVar10,DAT_1803ca5d8,DAT_1803bd168,DAT_1803ab250);
          }
          else {
            lVar11 = FUN_180371890(param_1,plVar10,DAT_1803bd168);
          }
          if (lVar11 != 0) {
            FUN_1801ddb80();
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            lVar11 = FUN_1801ddc10();
            if (lVar11 == 0) {
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
                        puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bd7d8);
                        if ((puVar7 == (undefined8 *)0x0) ||
                           (local_res10[0] = (code *)*puVar7, local_res10[0] == (code *)0x0)) {
                          puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ab778,DAT_1803bd7d8);
                          local_res10[0] = pcVar20;
                          if (puVar7 != (undefined8 *)0x0) {
                            local_res10[0] = (code *)*puVar7;
                          }
                          if (local_res10[0] == (code *)0x0) {
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
                                puVar7 = (undefined8 *)FUN_1800015a0(DAT_1803ca5d8,DAT_1803bd888);
                                if (((puVar7 == (undefined8 *)0x0) ||
                                    (pcVar21 = (code *)*puVar7, (code *)*puVar7 == (code *)0x0)) &&
                                   (puVar7 = (undefined8 *)
                                             FUN_1800015a0(DAT_1803ab778,DAT_1803bd888),
                                   pcVar21 = pcVar20, puVar7 != (undefined8 *)0x0)) {
                                  pcVar21 = (code *)*puVar7;
                                }
                                iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd898,pcVar21);
                                if ((iVar19 == 0) &&
                                   (iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd8a0,DAT_1803bd870),
                                   iVar19 == 0)) {
                                  plVar15 = (longlong *)FUN_180370ec0(param_1);
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
                                  local_res10[0] = DAT_1803bd8b0;
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
                                    FUN_180373340(param_1,&local_res20,DAT_1803bd8a8);
                                    uVar8 = 0x15;
                                    plVar14 = local_res20;
                                  }
                                  else {
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
                                      local_res10[0] = DAT_1803bd800;
                                      plVar15 = (longlong *)PyObject_GetItem(pcVar17,DAT_1803bd8a8);
                                      if (plVar15 != (longlong *)0x0) {
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
                                          local_res10[0] = DAT_1803bd8b8;
                                          plVar15 = (longlong *)
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
                                            FUN_180373340(param_1,&local_res20,DAT_1803bd8a8);
                                            uVar8 = 0x17;
                                            plVar14 = local_res20;
                                            goto LAB_1801ef302;
                                          }
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
                                                (iVar19 = PyObject_SetItem(pcVar17,DAT_1803bd8c0,
                                                                           pcVar12), iVar19 == 0))))
                                            {
                                              lVar11 = *(longlong *)(param_1 + 0x10);
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
                                                FUN_1803609a0(FUN_1801dfdd0,DAT_1803bd8d8,0,
                                                              DAT_1803bd0f8,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e0);
                                                FUN_1803609a0(FUN_1801e2da0,DAT_1803bd280,0,
                                                              DAT_1803bcf90,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e8);
                                                FUN_1803609a0(FUN_1801e31a0,DAT_1803bd1c8,0,
                                                              DAT_1803bd0e0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e8);
                                                FUN_1803609a0(FUN_1801e3500,DAT_1803bd1d0,0,
                                                              DAT_1803bceb0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f0);
                                                FUN_1803609a0(FUN_1801e3860,DAT_1803bd1d8,0,
                                                              DAT_1803bceb8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f8);
                                                FUN_1803609a0(FUN_1801e4130,DAT_1803bd288,0,
                                                              DAT_1803bd110,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd910,0,
                                                                      DAT_1803bd068,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd908;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd908;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd910,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd920,0,
                                                                      DAT_1803bdac0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd918;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd918;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd920,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd930,0,
                                                                      DAT_1803bd078,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd928;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd928;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd930,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd940,0,
                                                                      DAT_1803bce90,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd938;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd938;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd940,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd950,0,
                                                                      DAT_1803bcfe0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd948;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd948;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd950,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd960,0,
                                                                      DAT_1803bd100,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd958;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd958;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd960);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
                                                FUN_1803609a0(FUN_1801e4550,DAT_1803bd1e8,0,
                                                              DAT_1803bcf38,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
                                                FUN_1803609a0(FUN_1801e47a0,DAT_1803bd200,0,
                                                              DAT_1803bdad8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
                                                FUN_1803609a0(FUN_1801e49f0,DAT_1803bd208,0,
                                                              DAT_1803bcfa8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd970);
                                                FUN_1803609a0(FUN_1801e4c40,DAT_1803bd210,0,
                                                              DAT_1803bcf20,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                piVar1 = DAT_1803bd978;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd980);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_1803609a0(FUN_1801e5240,DAT_1803bd298,0,
                                                              DAT_1803bd028,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd900);
                                                lVar9 = FUN_1803609a0(0,DAT_1803bd2b0,0,
                                                                      DAT_1803bcfc0,0,0,uVar8,
                                                                      DAT_1803ca5e0,0,0,0);
                                                uVar8 = DAT_1803bd988;
                                                *(undefined8 *)(lVar9 + 200) = DAT_1803bd988;
                                                *(code **)(lVar9 + 0x78) = FUN_180360930;
                                                FUN_1801ddb80(uVar8,DAT_1803bd2b0,lVar9);
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd990);
                                                FUN_1803609a0(FUN_1801e5d00,DAT_1803bd218,0,
                                                              DAT_1803bd0b8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd998);
                                                FUN_1803609a0(FUN_1801e7520,DAT_1803bd228,0,
                                                              DAT_1803bdab8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd968);
                                                FUN_1803609a0(FUN_1801e7950,DAT_1803bd1f8,0,
                                                              DAT_1803bd070,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9a0);
                                                FUN_1803609a0(FUN_1801e7ba0,DAT_1803bd220,0,
                                                              DAT_1803bd118,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9a8);
                                                FUN_1803609a0(FUN_1801e9380,DAT_1803bd230,0,
                                                              DAT_1803bcf50,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9b0);
                                                FUN_1803609a0(FUN_1801e9720,DAT_1803bd238,0,
                                                              DAT_1803bd048,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                piVar1 = DAT_1803bd978;
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9b8);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_1803609a0(FUN_1801e9ae0,DAT_1803bd240,0,
                                                              DAT_1803bcea0,piVar1,0,uVar8,
                                                              DAT_1803ca5e0,0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8e0);
                                                FUN_1803609a0(FUN_1801e9e80,DAT_1803bd9c0,0,
                                                              DAT_1803bd008,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9c8);
                                                FUN_1803609a0(FUN_1801eb0b0,DAT_1803bd198,0,
                                                              DAT_1803bcf30,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9c8);
                                                FUN_1803609a0(FUN_1801eb930,DAT_1803bd1a0,0,
                                                              DAT_1803bd0c8,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d0);
                                                FUN_1803609a0(FUN_1801ec5a0,DAT_1803bd748,0,
                                                              DAT_1803bd098,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd8f8);
                                                FUN_1803609a0(FUN_1801ecac0,DAT_1803bd750,0,
                                                              DAT_1803bcf78,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d8);
                                                FUN_1803609a0(FUN_1801ecee0,DAT_1803bd188,0,
                                                              DAT_1803bdab0,0,0,uVar8,DAT_1803ca5e0,
                                                              0,0,0);
                                                FUN_1801ddb80();
                                                uVar8 = FUN_18036fb80(param_1,DAT_1803bd9d8);
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
                                      uVar8 = DAT_1803bd880;
                                      if (iVar19 != 0) goto LAB_1801eef09;
                                      local_res20 = (longlong *)
                                                    FUN_18036f270(param_1,plVar14,DAT_1803bd888,
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

