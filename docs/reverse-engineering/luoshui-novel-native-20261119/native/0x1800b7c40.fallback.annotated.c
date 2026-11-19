/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1800b7c40 */

longlong * FUN_1800b7c40(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  longlong *plVar3;
  longlong *plVar4;
  longlong lVar5;
  int *piVar6;
  longlong *plVar7;
  char cVar8;
  int iVar9;
  int iVar10;
  longlong *plVar11;
  code *pcVar12;
  longlong *plVar13;
  longlong *plVar14;
  longlong *plVar15;
  longlong *plVar16;
  code *pcVar17;
  code *pcVar18;
  undefined8 uVar19;
  undefined8 *puVar20;
  code *local_res20;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  
  plVar7 = DAT_1803b1ed0;
  plVar1 = (longlong *)param_3[1];
  plVar2 = (longlong *)*param_3;
  plVar3 = (longlong *)param_3[2];
  plVar4 = (longlong *)param_3[3];
  local_88 = (longlong *)0x0;
  local_a0 = (longlong *)0x0;
  local_90 = (longlong *)0x0;
  local_98 = (longlong *)0x0;
  if (DAT_1803b1ed0 == (longlong *)0x0) {
LAB_1800b7cf9:
    DAT_1803b1ed0 = (longlong *)FUN_18036cd10(param_1,DAT_1803b1ec8,DAT_1803ca7c0,0x50);
  }
  else if ((1 < *DAT_1803b1ed0) || (DAT_1803b1ed0[2] != 0)) {
    if ((DAT_1803b1ed0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b1ed0 && (*DAT_1803b1ed0 = *DAT_1803b1ed0 + -1, *plVar7 == 0)))) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    goto LAB_1800b7cf9;
  }
  plVar7 = DAT_1803b1ed0;
  FUN_180001d00(param_1,DAT_1803b1ed0);
  plVar11 = (longlong *)FUN_180370dc0(param_1,plVar2);
  if (plVar11 == (longlong *)0x0) {
    plVar13 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar19 = 0x53;
    local_res20 = (code *)0x0;
    pcVar12 = (code *)0x0;
  }
  else {
/* novels.pipeline.chapter_review_cycle#24="issues" */
    pcVar17 = (code *)FUN_18036f6e0(param_1,plVar11,DAT_1803b2020);
    if (pcVar17 == (code *)0x0) {
      pcVar17 = _Py_NoneStruct_exref;
    }
    iVar9 = FUN_180001510(pcVar17);
    local_88 = plVar11;
    if (iVar9 == -1) {
      plVar13 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < *(int *)pcVar17) &&
         (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
        (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
      }
      uVar19 = 0x54;
      local_res20 = (code *)0x0;
      pcVar12 = (code *)0x0;
    }
    else {
      if (iVar9 != 1) {
        if ((-1 < *(int *)pcVar17) &&
           (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
          (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
        }
        pcVar17 = (code *)FUN_180372160(param_1,0);
      }
      pcVar12 = (code *)FUN_180372690(param_1,pcVar17);
      if ((-1 < *(int *)pcVar17) &&
         (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
        (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
      }
      local_res20 = _Py_FalseStruct_exref;
      if (pcVar12 == (code *)0x0) {
        plVar13 = *(longlong **)(param_1 + 0x60);
        uVar19 = 0x54;
        *(undefined8 *)(param_1 + 0x60) = 0;
        local_res20 = (code *)0x0;
        pcVar12 = (code *)0x0;
      }
      else {
        if (*(int *)_Py_FalseStruct_exref + 1 != 0) {
          *(int *)_Py_FalseStruct_exref = *(int *)_Py_FalseStruct_exref + 1;
        }
        iVar9 = FUN_180001510(plVar1);
        if (iVar9 == -1) {
          plVar13 = *(longlong **)(param_1 + 0x60);
          uVar19 = 0x57;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else if (iVar9 == 0) {
LAB_1800b840b:
          puVar20 = (undefined8 *)(param_1 + 0x60);
          iVar9 = FUN_180001510(plVar3);
          if (iVar9 == -1) {
            plVar13 = (longlong *)*puVar20;
            *puVar20 = 0;
            uVar19 = 0x62;
          }
          else if (iVar9 == 0) {
LAB_1800b88eb:
            puVar20 = (undefined8 *)(param_1 + 0x60);
            iVar9 = FUN_180001510(plVar4);
            if (iVar9 == -1) {
              plVar13 = (longlong *)*puVar20;
              *puVar20 = 0;
              uVar19 = 0x6c;
            }
            else if (iVar9 == 0) {
LAB_1800b8eaa:
/* novels.pipeline.chapter_review_cycle#24="issues" */
              PyDict_SetItem(plVar11,DAT_1803b2020,pcVar12);
              if (local_res20 == (code *)0x0) {
                plVar11 = (longlong *)
                          PyUnicode_FromFormat
                                    ("cannot access local variable \'%U\' where it is not associated with a value"
/* novels.pipeline.chapter_review_cycle#37="blocked" */
                                     ,DAT_1803b2088);
                plVar13 = (longlong *)
                          FUN_180376770(param_1,*(undefined8 *)PyExc_UnboundLocalError_exref,plVar11
                                       );
                if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                }
                FUN_18000d1d0(param_1,plVar13);
                uVar19 = 0x78;
              }
              else {
                iVar9 = FUN_180001510(local_res20);
                if (iVar9 != -1) {
                  if (iVar9 != 0) {
/* novels.pipeline.chapter_review_cycle#38="passed" */
                    PyDict_SetItem(plVar11,DAT_1803b2090,_Py_FalseStruct_exref);
                  }
                  FUN_180001d80(param_1);
                  iVar9 = (int)*plVar11 + 1;
                  iVar10 = (int)*plVar11;
                  if (iVar9 != 0) {
                    *(int *)plVar11 = iVar9;
                    iVar10 = iVar9;
                  }
                  if ((-1 < iVar10) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                  }
                  if ((-1 < *(int *)pcVar12) &&
                     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
                  }
                  if ((-1 < *(int *)local_res20) &&
                     (*(longlong *)local_res20 = *(longlong *)local_res20 + -1,
                     *(longlong *)local_res20 == 0)) {
                    (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))(local_res20);
                  }
                  if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                    (**(code **)(local_a0[1] + 0x30))();
                  }
                  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
                     (*local_90 = *local_90 + -1, *local_90 == 0)) {
                    (**(code **)(local_90[1] + 0x30))();
                  }
                  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                     (*local_98 = *local_98 + -1, *local_98 == 0)) {
                    (**(code **)(local_98[1] + 0x30))();
                  }
                  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
                    (**(code **)(plVar2[1] + 0x30))();
                  }
                  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                    (**(code **)(plVar1[1] + 0x30))();
                  }
                  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
                    (**(code **)(plVar3[1] + 0x30))();
                  }
                  if ((int)*plVar4 < 0) {
                    return plVar11;
                  }
                  *plVar4 = *plVar4 + -1;
                  if (*plVar4 != 0) {
                    return plVar11;
                  }
                  (**(code **)(plVar4[1] + 0x30))(plVar4);
                  return plVar11;
                }
                plVar13 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar19 = 0x78;
              }
            }
            else {
              plVar13 = (longlong *)FUN_18037f0c0();
              if (plVar13 == (longlong *)0x0) {
                plVar13 = (longlong *)*puVar20;
                *puVar20 = 0;
                uVar19 = 0x6d;
              }
              else {
                plVar14 = (longlong *)FUN_18000d280();
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
                if (plVar14 == (longlong *)0x0) {
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar19 = 0x6d;
                }
                else {
                  do {
                    pcVar18 = local_res20;
                    pcVar17 = *(code **)(plVar14[1] + 0xe0);
                    if (pcVar17 == (code *)0x0) {
                      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                   "\'%s\' object is not an iterator",
                                   *(undefined8 *)(plVar14[1] + 0x18));
LAB_1800b8d87:
                      plVar13 = *(longlong **)(param_1 + 0x60);
                      if (plVar13 != (longlong *)0x0) {
                        plVar15 = (longlong *)plVar13[1];
                        if ((*(uint *)((longlong *)plVar13[1] + 0x15) & 0x40000000) == 0) {
                          plVar15 = plVar13;
                        }
                        if ((plVar15 != *(longlong **)PyExc_StopIteration_exref) &&
                           (((-1 < *(int *)(plVar15[1] + 0xa8) ||
                             ((*(uint *)(plVar15 + 0x15) & 0x40000000) == 0)) ||
                            (cVar8 = FUN_180382640(), cVar8 != '\x01')))) {
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          uVar19 = 0x6d;
                          if (local_98 == (longlong *)0x0) goto LAB_1800b8e10;
                          goto LAB_1800b8df8;
                        }
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                          (**(code **)(plVar13[1] + 0x30))(plVar13);
                        }
                      }
                      if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                         (*local_98 = *local_98 + -1, *local_98 == 0)) {
                        (**(code **)(local_98[1] + 0x30))(local_98);
                      }
                      if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                        (**(code **)(plVar14[1] + 0x30))(plVar14);
                      }
                      goto LAB_1800b8eaa;
                    }
                    plVar13 = (longlong *)(*pcVar17)(plVar14);
                    if (plVar13 == (longlong *)0x0) goto LAB_1800b8d87;
                    if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                       (*local_98 = *local_98 + -1, *local_98 == 0)) {
                      (**(code **)(local_98[1] + 0x30))(local_98);
                    }
                    iVar9 = (int)*plVar13 + 1;
                    if (iVar9 != 0) {
                      *(int *)plVar13 = iVar9;
                    }
                    if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
                       (*local_98 = *local_98 + -1, *local_98 == 0)) {
                      (**(code **)(local_98[1] + 0x30))();
                    }
/* novels.pipeline.chapter_review_cycle#25="severity" */
                    uVar19 = DAT_1803b2028;
                    plVar15 = (longlong *)FUN_18037f0c0();
                    local_98 = plVar13;
                    if (plVar15 == (longlong *)0x0) {
LAB_1800b8d44:
                      plVar13 = *(longlong **)(param_1 + 0x60);
                      uVar19 = 0x6f;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      goto LAB_1800b8df8;
                    }
/* novels.pipeline.chapter_review_cycle#35="block" */
                    iVar9 = FUN_1803836d0(plVar15,DAT_1803b2078);
                    if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                      (**(code **)(plVar15[1] + 0x30))(plVar15);
                    }
/* novels.pipeline.chapter_review_cycle#27="critical" */
                    plVar13 = DAT_1803b2038;
                    if (iVar9 == -1) goto LAB_1800b8d44;
                    if (iVar9 == 1) {
/* novels.pipeline.chapter_review_cycle#27="critical" */
                      iVar9 = (int)*DAT_1803b2038 + 1;
                      if (iVar9 != 0) {
/* novels.pipeline.chapter_review_cycle#27="critical" */
                        *(int *)DAT_1803b2038 = iVar9;
                      }
                    }
                    else {
                      plVar13 = (longlong *)FUN_18037f0c0();
                      if (plVar13 == (longlong *)0x0) goto LAB_1800b8d44;
                    }
                    plVar15 = (longlong *)_PyDict_NewPresized(4);
                    PyDict_SetItem(plVar15,uVar19,plVar13);
                    if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                    }
/* novels.pipeline.chapter_review_cycle#28="category" */
                    uVar19 = DAT_1803b2040;
                    plVar13 = (longlong *)FUN_18037f0c0();
                    if (plVar13 == (longlong *)0x0) {
                      uVar19 = 0x70;
LAB_1800b8d0e:
                      plVar13 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                        (**(code **)(plVar15[1] + 0x30))(plVar15);
                      }
                      goto LAB_1800b8df8;
                    }
                    PyDict_SetItem(plVar15,uVar19,plVar13);
                    if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                    }
/* novels.pipeline.chapter_review_cycle#29="summary" */
                    uVar19 = DAT_1803b2048;
                    plVar13 = (longlong *)FUN_18037f0c0();
                    if (plVar13 == (longlong *)0x0) {
                      uVar19 = 0x71;
                      goto LAB_1800b8d0e;
                    }
                    PyDict_SetItem(plVar15,uVar19,plVar13);
                    if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                    }
/* novels.pipeline.chapter_review_cycle#31="source" ; novels.pipeline.chapter_review_cycle#36="sensitive" */
                    PyDict_SetItem(plVar15,DAT_1803b2058,DAT_1803b2080);
                    lVar5 = *(longlong *)(pcVar12 + 0x10);
                    cVar8 = FUN_1803722b0(pcVar12,lVar5 + 1);
                    if (cVar8 != '\0') {
                      *(longlong **)(*(longlong *)(pcVar12 + 0x18) + lVar5 * 8) = plVar15;
                    }
                    plVar13 = (longlong *)FUN_18037f0c0();
                    if (plVar13 == (longlong *)0x0) {
LAB_1800b8cdf:
                      plVar13 = *(longlong **)(param_1 + 0x60);
                      uVar19 = 0x74;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      goto LAB_1800b8df8;
                    }
/* novels.pipeline.chapter_review_cycle#35="block" */
                    iVar9 = FUN_1803836d0(plVar13,DAT_1803b2078);
                    if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                    }
                    if (iVar9 == -1) goto LAB_1800b8cdf;
                    if (iVar9 == 1) {
                      local_res20 = _Py_TrueStruct_exref;
                      if (*(int *)_Py_TrueStruct_exref + 1 != 0) {
                        *(int *)_Py_TrueStruct_exref = *(int *)_Py_TrueStruct_exref + 1;
                      }
                      if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
                         (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                         *(longlong *)pcVar18 == 0)) {
                        (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))();
                      }
                    }
                    lVar5 = *(longlong *)(param_1 + 0x10);
                    if (((*(int *)(*(longlong *)(lVar5 + 0x60) + 0x4d4) != 0) ||
                        (*(int *)(lVar5 + 0x1a0) != 0)) &&
                       ((iVar9 = Py_MakePendingCalls(), iVar9 < 0 &&
                        (*(longlong *)(param_1 + 0x60) != 0)))) goto LAB_1800b8cc6;
                    if (*(int *)(lVar5 + 0x174) != 0) {
                      PyEval_SaveThread();
                      PyEval_AcquireThread(param_1);
                    }
                  } while (*(longlong *)(param_1 + 0x80) == 0);
                  plVar11 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x80) = 0;
                  uVar19 = FUN_180373490(param_1);
                  *(undefined8 *)(param_1 + 0x60) = uVar19;
                  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                    (**(code **)(plVar11[1] + 0x30))(plVar11);
                  }
LAB_1800b8cc6:
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  uVar19 = 0x6d;
LAB_1800b8df8:
                  if ((-1 < (int)*local_98) && (*local_98 = *local_98 + -1, *local_98 == 0)) {
                    (**(code **)(local_98[1] + 0x30))(local_98);
                  }
LAB_1800b8e10:
                  if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                    (**(code **)(plVar14[1] + 0x30))(plVar14);
                  }
                }
              }
            }
          }
          else {
            plVar13 = (longlong *)FUN_18037f0c0();
            if (plVar13 == (longlong *)0x0) {
              plVar13 = (longlong *)*puVar20;
              *puVar20 = 0;
              uVar19 = 99;
            }
            else {
              plVar14 = (longlong *)FUN_18000d280();
              if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                (**(code **)(plVar13[1] + 0x30))(plVar13);
              }
              if (plVar14 == (longlong *)0x0) {
                plVar13 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar19 = 99;
              }
              else {
                do {
                  pcVar17 = *(code **)(plVar14[1] + 0xe0);
                  if (pcVar17 == (code *)0x0) {
                    PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                 "\'%s\' object is not an iterator",
                                 *(undefined8 *)(plVar14[1] + 0x18));
LAB_1800b879a:
                    plVar13 = *(longlong **)(param_1 + 0x60);
                    if (plVar13 != (longlong *)0x0) {
                      plVar15 = (longlong *)plVar13[1];
                      if ((*(uint *)((longlong *)plVar13[1] + 0x15) & 0x40000000) == 0) {
                        plVar15 = plVar13;
                      }
                      if ((plVar15 != *(longlong **)PyExc_StopIteration_exref) &&
                         (((-1 < *(int *)(plVar15[1] + 0xa8) ||
                           ((*(uint *)(plVar15 + 0x15) & 0x40000000) == 0)) ||
                          (cVar8 = FUN_180382640(), cVar8 != '\x01')))) {
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        uVar19 = 99;
                        if (local_90 != (longlong *)0x0) goto LAB_1800b8812;
                        goto LAB_1800b882c;
                      }
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                        (**(code **)(plVar13[1] + 0x30))(plVar13);
                      }
                    }
                    if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
                       (*local_90 = *local_90 + -1, *local_90 == 0)) {
                      (**(code **)(local_90[1] + 0x30))(local_90);
                    }
                    if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                      (**(code **)(plVar14[1] + 0x30))(plVar14);
                    }
                    goto LAB_1800b88eb;
                  }
                  plVar13 = (longlong *)(*pcVar17)(plVar14);
                  if (plVar13 == (longlong *)0x0) goto LAB_1800b879a;
                  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
                     (*local_90 = *local_90 + -1, *local_90 == 0)) {
                    (**(code **)(local_90[1] + 0x30))();
                  }
                  iVar9 = (int)*plVar13 + 1;
                  if (iVar9 != 0) {
                    *(int *)plVar13 = iVar9;
                  }
                  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
                     (*local_90 = *local_90 + -1, *local_90 == 0)) {
                    (**(code **)(local_90[1] + 0x30))();
                  }
/* novels.pipeline.chapter_review_cycle#25="severity" */
                  uVar19 = DAT_1803b2028;
                  plVar15 = (longlong *)FUN_18037f0c0();
                  local_90 = plVar13;
                  if (plVar15 == (longlong *)0x0) {
                    plVar13 = *(longlong **)(param_1 + 0x60);
                    uVar19 = 0x65;
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    goto LAB_1800b8812;
                  }
                  plVar16 = (longlong *)_PyDict_NewPresized(5);
                  PyDict_SetItem(plVar16,uVar19,plVar15);
                  if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                  }
/* novels.pipeline.chapter_review_cycle#28="category" */
                  uVar19 = DAT_1803b2040;
                  plVar13 = (longlong *)FUN_18037f0c0();
                  if (plVar13 == (longlong *)0x0) {
                    uVar19 = 0x66;
LAB_1800b8738:
                    plVar13 = *(longlong **)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                      (**(code **)(plVar16[1] + 0x30))(plVar16);
                    }
                    goto LAB_1800b8812;
                  }
                  PyDict_SetItem(plVar16,uVar19,plVar13);
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
/* novels.pipeline.chapter_review_cycle#29="summary" */
                  uVar19 = DAT_1803b2048;
                  plVar13 = (longlong *)FUN_18037f0c0();
                  if (plVar13 == (longlong *)0x0) {
                    uVar19 = 0x67;
                    goto LAB_1800b8738;
                  }
                  PyDict_SetItem(plVar16,uVar19,plVar13);
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
/* novels.pipeline.chapter_review_cycle#33="suggestion" */
                  uVar19 = DAT_1803b2068;
                  plVar13 = (longlong *)FUN_18037f0c0();
                  if (plVar13 == (longlong *)0x0) {
                    uVar19 = 0x68;
                    goto LAB_1800b8738;
                  }
                  PyDict_SetItem(plVar16,uVar19,plVar13);
                  if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                    (**(code **)(plVar13[1] + 0x30))(plVar13);
                  }
/* novels.pipeline.chapter_review_cycle#31="source" ; novels.pipeline.chapter_review_cycle#34="ai_tells" */
                  PyDict_SetItem(plVar16,DAT_1803b2058,DAT_1803b2070);
                  lVar5 = *(longlong *)(pcVar12 + 0x10);
                  cVar8 = FUN_1803722b0(pcVar12,lVar5 + 1);
                  if (cVar8 != '\0') {
                    *(longlong **)(*(longlong *)(pcVar12 + 0x18) + lVar5 * 8) = plVar16;
                  }
                  lVar5 = *(longlong *)(param_1 + 0x10);
                  if ((((*(int *)(*(longlong *)(lVar5 + 0x60) + 0x4d4) != 0) ||
                       (*(int *)(lVar5 + 0x1a0) != 0)) && (iVar9 = Py_MakePendingCalls(), iVar9 < 0)
                      ) && (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_1800b8711;
                  if (*(int *)(lVar5 + 0x174) != 0) {
                    PyEval_SaveThread();
                    PyEval_AcquireThread(param_1);
                  }
                } while (*(longlong *)(param_1 + 0x80) == 0);
                plVar11 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x80) = 0;
                uVar19 = FUN_180373490(param_1);
                *(undefined8 *)(param_1 + 0x60) = uVar19;
                if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                }
LAB_1800b8711:
                plVar13 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                uVar19 = 99;
LAB_1800b8812:
                if ((-1 < (int)*local_90) && (*local_90 = *local_90 + -1, *local_90 == 0)) {
                  (**(code **)(local_90[1] + 0x30))(local_90);
                }
LAB_1800b882c:
                if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                  (**(code **)(plVar14[1] + 0x30))(plVar14);
                }
              }
            }
          }
        }
        else {
          plVar13 = (longlong *)FUN_18037f0c0();
          if (plVar13 == (longlong *)0x0) {
            plVar13 = *(longlong **)(param_1 + 0x60);
            uVar19 = 0x58;
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            plVar14 = (longlong *)FUN_18000d280();
            if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
              (**(code **)(plVar13[1] + 0x30))(plVar13);
            }
            if (plVar14 == (longlong *)0x0) {
              plVar13 = *(longlong **)(param_1 + 0x60);
              uVar19 = 0x58;
              *(undefined8 *)(param_1 + 0x60) = 0;
            }
            else {
              do {
                pcVar18 = local_res20;
                pcVar17 = *(code **)(plVar14[1] + 0xe0);
                if (pcVar17 == (code *)0x0) {
                  PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                               "\'%s\' object is not an iterator",*(undefined8 *)(plVar14[1] + 0x18)
                              );
LAB_1800b82e1:
                  puVar20 = (undefined8 *)(param_1 + 0x60);
                  plVar13 = (longlong *)*puVar20;
                  if (plVar13 != (longlong *)0x0) {
                    plVar15 = (longlong *)plVar13[1];
                    if ((*(uint *)((longlong *)plVar13[1] + 0x15) & 0x40000000) == 0) {
                      plVar15 = plVar13;
                    }
                    if ((plVar15 != *(longlong **)PyExc_StopIteration_exref) &&
                       (((-1 < *(int *)(plVar15[1] + 0xa8) ||
                         ((*(uint *)(plVar15 + 0x15) & 0x40000000) == 0)) ||
                        (cVar8 = FUN_180382640(), cVar8 != '\x01')))) {
                      *puVar20 = 0;
                      uVar19 = 0x58;
                      if (local_a0 != (longlong *)0x0) goto LAB_1800b8353;
                      goto LAB_1800b836f;
                    }
                    *puVar20 = 0;
                    if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                      (**(code **)(plVar13[1] + 0x30))(plVar13);
                    }
                  }
                  if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                    (**(code **)(local_a0[1] + 0x30))(local_a0);
                  }
                  if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                    (**(code **)(plVar14[1] + 0x30))(plVar14);
                  }
                  goto LAB_1800b840b;
                }
                plVar13 = (longlong *)(*pcVar17)(plVar14);
                if (plVar13 == (longlong *)0x0) goto LAB_1800b82e1;
                if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                   (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                  (**(code **)(local_a0[1] + 0x30))();
                }
                iVar9 = (int)*plVar13 + 1;
                if (iVar9 != 0) {
                  *(int *)plVar13 = iVar9;
                }
                if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
                   (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                  (**(code **)(local_a0[1] + 0x30))();
                }
/* novels.pipeline.chapter_review_cycle#25="severity" */
                uVar19 = DAT_1803b2028;
                plVar15 = (longlong *)FUN_18037f0c0();
                local_a0 = plVar13;
                if (plVar15 == (longlong *)0x0) {
LAB_1800b8296:
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  uVar19 = 0x5a;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_1800b8353;
                }
/* novels.pipeline.chapter_review_cycle#26="error" */
                iVar9 = FUN_1803836d0(plVar15,DAT_1803b2030);
                if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                  (**(code **)(plVar15[1] + 0x30))(plVar15);
                }
/* novels.pipeline.chapter_review_cycle#27="critical" */
                plVar13 = DAT_1803b2038;
                if (iVar9 == -1) goto LAB_1800b8296;
                if (iVar9 == 1) {
/* novels.pipeline.chapter_review_cycle#27="critical" */
                  iVar9 = (int)*DAT_1803b2038 + 1;
                  if (iVar9 != 0) {
/* novels.pipeline.chapter_review_cycle#27="critical" */
                    *(int *)DAT_1803b2038 = iVar9;
                  }
                }
                else {
                  plVar13 = (longlong *)FUN_18037f0c0();
                  if (plVar13 == (longlong *)0x0) goto LAB_1800b8296;
                }
                plVar15 = (longlong *)_PyDict_NewPresized(4);
                PyDict_SetItem(plVar15,uVar19,plVar13);
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
/* novels.pipeline.chapter_review_cycle#28="category" */
                uVar19 = DAT_1803b2040;
                plVar13 = (longlong *)FUN_18037f0c0();
                if (plVar13 == (longlong *)0x0) {
                  uVar19 = 0x5b;
LAB_1800b825d:
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  if ((-1 < (int)*plVar15) && (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                    (**(code **)(plVar15[1] + 0x30))(plVar15);
                  }
                  goto LAB_1800b8353;
                }
                PyDict_SetItem(plVar15,uVar19,plVar13);
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
/* novels.pipeline.chapter_review_cycle#29="summary" */
                uVar19 = DAT_1803b2048;
                plVar13 = (longlong *)FUN_18037f0c0();
                if (plVar13 == (longlong *)0x0) {
                  uVar19 = 0x5c;
                  goto LAB_1800b825d;
                }
                PyDict_SetItem(plVar15,uVar19,plVar13);
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
/* novels.pipeline.chapter_review_cycle#31="source" ; novels.pipeline.chapter_review_cycle#32="post_write" */
                PyDict_SetItem(plVar15,DAT_1803b2058,DAT_1803b2060);
                lVar5 = *(longlong *)(pcVar12 + 0x10);
                cVar8 = FUN_1803722b0(pcVar12,lVar5 + 1);
                if (cVar8 != '\0') {
                  *(longlong **)(*(longlong *)(pcVar12 + 0x18) + lVar5 * 8) = plVar15;
                }
                plVar13 = (longlong *)FUN_18037f0c0();
                if (plVar13 == (longlong *)0x0) {
LAB_1800b822b:
                  plVar13 = *(longlong **)(param_1 + 0x60);
                  uVar19 = 0x5f;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  goto LAB_1800b8353;
                }
/* novels.pipeline.chapter_review_cycle#26="error" */
                iVar9 = FUN_1803836d0(plVar13,DAT_1803b2030);
                if ((-1 < (int)*plVar13) && (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                }
                if (iVar9 == -1) goto LAB_1800b822b;
                if (iVar9 == 1) {
                  local_res20 = _Py_TrueStruct_exref;
                  if (*(int *)_Py_TrueStruct_exref + 1 != 0) {
                    *(int *)_Py_TrueStruct_exref = *(int *)_Py_TrueStruct_exref + 1;
                  }
                  if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
                     (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1, *(longlong *)pcVar18 == 0))
                  {
                    (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))();
                  }
                }
                lVar5 = *(longlong *)(param_1 + 0x10);
                if (((*(int *)(*(longlong *)(lVar5 + 0x60) + 0x4d4) != 0) ||
                    (*(int *)(lVar5 + 0x1a0) != 0)) &&
                   ((iVar9 = Py_MakePendingCalls(), iVar9 < 0 &&
                    (*(longlong *)(param_1 + 0x60) != 0)))) goto LAB_1800b820f;
                if (*(int *)(lVar5 + 0x174) != 0) {
                  PyEval_SaveThread();
                  PyEval_AcquireThread(param_1);
                }
              } while (*(longlong *)(param_1 + 0x80) == 0);
              plVar11 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x80) = 0;
              uVar19 = FUN_180373490(param_1);
              *(undefined8 *)(param_1 + 0x60) = uVar19;
              if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                (**(code **)(plVar11[1] + 0x30))(plVar11);
              }
LAB_1800b820f:
              plVar13 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              uVar19 = 0x58;
LAB_1800b8353:
              if ((-1 < (int)*local_a0) && (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
                (**(code **)(local_a0[1] + 0x30))(local_a0);
              }
LAB_1800b836f:
              if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                (**(code **)(plVar14[1] + 0x30))(plVar14);
              }
            }
          }
        }
      }
    }
  }
  piVar6 = (int *)plVar13[5];
  if (piVar6 == (int *)0x0) {
    pcVar17 = (code *)FUN_18038e2a0(plVar7,uVar19);
    plVar11 = (longlong *)plVar13[5];
    pcVar18 = pcVar17;
    if (pcVar17 == _Py_NoneStruct_exref) {
      pcVar18 = (code *)0x0;
    }
    else {
LAB_1800b8f81:
      pcVar17 = pcVar18;
      if ((pcVar18 != (code *)0x0) && (*(int *)pcVar18 + 1 != 0)) {
        *(int *)pcVar18 = *(int *)pcVar18 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar6 + 6) == plVar7) goto LAB_1800b8fca;
    pcVar17 = (code *)FUN_18038e2a0(plVar7,uVar19);
    *(int **)(pcVar17 + 0x10) = piVar6;
    if (*piVar6 + 1 != 0) {
      *piVar6 = *piVar6 + 1;
    }
    plVar11 = (longlong *)plVar13[5];
    pcVar18 = pcVar17;
    if (pcVar17 != _Py_NoneStruct_exref) goto LAB_1800b8f81;
    pcVar18 = (code *)0x0;
  }
  plVar13[5] = (longlong)pcVar18;
  if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
     (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))();
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
LAB_1800b8fca:
  FUN_18036d0a0(plVar7,"oooooooooo",plVar2,plVar1,plVar3,plVar4,local_88,pcVar12,local_res20,
                local_a0,local_90,local_98);
  plVar11 = DAT_1803b1ed0;
  if (plVar7 == DAT_1803b1ed0) {
    if ((-1 < (int)*DAT_1803b1ed0) && (*DAT_1803b1ed0 = *DAT_1803b1ed0 + -1, *plVar11 == 0)) {
      (**(code **)(plVar11[1] + 0x30))();
    }
    DAT_1803b1ed0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
     (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
  if (((local_res20 != (code *)0x0) && (-1 < *(int *)local_res20)) &&
     (*(longlong *)local_res20 = *(longlong *)local_res20 + -1, *(longlong *)local_res20 == 0)) {
    (**(code **)(*(longlong *)(local_res20 + 8) + 0x30))(local_res20);
  }
  if (((local_a0 != (longlong *)0x0) && (-1 < (int)*local_a0)) &&
     (*local_a0 = *local_a0 + -1, *local_a0 == 0)) {
    (**(code **)(local_a0[1] + 0x30))(local_a0);
  }
  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
     (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))(local_90);
  }
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))(local_98);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))();
  }
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar13;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
